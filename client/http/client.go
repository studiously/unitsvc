package http

import (
	"github.com/go-kit/kit/log"
	"github.com/studiously/unitsvc/unitsvc"
	"strings"
	"net/url"
)

func New(instance string, logger log.Logger) (unitsvc.Service, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}

	return unitsvc.MakeClientEndpoints(instance)
}