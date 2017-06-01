package oauth2

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pkg/errors"
	"golang.org/x/oauth2/clientcredentials"
)

type HTTPRecovator struct {
	Config   *clientcredentials.Config
	Dry      bool
	Endpoint *url.URL
	Client   *http.Client
}

func (r *HTTPRecovator) RevokeToken(ctx context.Context, token string) error {
	var ep = *r.Endpoint
	ep.Path = RevocationPath

	data := url.Values{"token": []string{token}}
	hreq, err := http.NewRequest("POST", ep.String(), bytes.NewBufferString(data.Encode()))
	if err != nil {
		return errors.WithStack(err)
	}

	hreq.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	hreq.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	hreq.SetBasicAuth(r.Config.ClientID, r.Config.ClientSecret)
	if r.Client == nil {
		r.Client = http.DefaultClient
	}
	hres, err := r.Client.Do(hreq)
	if err != nil {
		return errors.WithStack(err)
	}
	defer hres.Body.Close()

	body, _ := ioutil.ReadAll(hres.Body)
	if hres.StatusCode < 200 || hres.StatusCode >= 300 {
		return errors.Errorf("Expected 2xx status code but got %d.\n%s", hres.StatusCode, body)
	}
	return nil
}
