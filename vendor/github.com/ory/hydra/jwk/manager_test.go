package jwk_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/ory/fosite"
	"github.com/ory/herodot"
	"github.com/ory/hydra/compose"
	"github.com/ory/hydra/integration"
	. "github.com/ory/hydra/jwk"
	"github.com/ory/hydra/pkg"
	"github.com/ory/ladon"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var managers = map[string]Manager{}

var testGenerator = &RS256Generator{}

var ts *httptest.Server
var httpManager *HTTPManager

func init() {
	localWarden, httpClient := compose.NewMockFirewall(
		"tests",
		"alice",
		fosite.Arguments{
			"hydra.keys.create",
			"hydra.keys.get",
			"hydra.keys.delete",
			"hydra.keys.update",
		}, &ladon.DefaultPolicy{
			ID:        "1",
			Subjects:  []string{"alice"},
			Resources: []string{"rn:hydra:keys:<faz|bar|foo|anonymous><.*>"},
			Actions:   []string{"create", "get", "delete", "update"},
			Effect:    ladon.AllowAccess,
		}, &ladon.DefaultPolicy{
			ID:        "2",
			Subjects:  []string{"alice", ""},
			Resources: []string{"rn:hydra:keys:anonymous<.*>"},
			Actions:   []string{"get"},
			Effect:    ladon.AllowAccess,
		},
	)

	router := httprouter.New()
	h := Handler{
		Manager: &MemoryManager{},
		W:       localWarden,
		H:       herodot.NewJSONWriter(nil),
	}
	h.SetRoutes(router)
	ts := httptest.NewServer(router)
	u, _ := url.Parse(ts.URL + "/keys")
	managers["memory"] = &MemoryManager{}
	httpManager = &HTTPManager{Client: httpClient, Endpoint: u}
	managers["http"] = httpManager
}

func randomBytes(n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, bytes); err != nil {
		return []byte{}, errors.WithStack(err)
	}
	return bytes, nil
}

var encryptionKey, _ = randomBytes(32)

func TestMain(m *testing.M) {
	connectToPG()
	connectToMySQL()

	s := m.Run()
	integration.KillAll()
	os.Exit(s)
}

func connectToPG() {
	var db = integration.ConnectToPostgres()
	s := &SQLManager{DB: db, Cipher: &AEAD{Key: encryptionKey}}
	if _, err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["postgres"] = s
}

func connectToMySQL() {
	var db = integration.ConnectToMySQL()
	s := &SQLManager{DB: db, Cipher: &AEAD{Key: encryptionKey}}

	if _, err := s.CreateSchemas(); err != nil {
		log.Fatalf("Could not create postgres schema: %v", err)
	}

	managers["mysql"] = s
}

func TestHTTPManagerPublicKeyGet(t *testing.T) {
	anonymous := &HTTPManager{Endpoint: httpManager.Endpoint, Client: http.DefaultClient}
	ks, _ := testGenerator.Generate("")
	priv := ks.Key("private")

	name := "http"
	m := httpManager

	_, err := m.GetKey("anonymous", "baz")
	pkg.AssertError(t, true, err, name)

	err = m.AddKey("anonymous", First(priv))
	pkg.AssertError(t, false, err, name)

	time.Sleep(time.Millisecond * 100)

	got, err := anonymous.GetKey("anonymous", "private")
	pkg.RequireError(t, false, err, name)
	assert.Equal(t, priv, got.Keys, "%s", name)
}

func TestManagerKey(t *testing.T) {
	ks, _ := testGenerator.Generate("")
	priv := ks.Key("private")
	pub := ks.Key("public")

	for name, m := range managers {
		t.Run(fmt.Sprintf("case=%s", name), func(t *testing.T) {
			_, err := m.GetKey("faz", "baz")
			assert.NotNil(t, err)

			err = m.AddKey("faz", First(priv))
			assert.Nil(t, err)

			time.Sleep(time.Millisecond * 100)

			got, err := m.GetKey("faz", "private")
			assert.Nil(t, err)
			assert.Equal(t, priv, got.Keys, "%s", name)

			err = m.AddKey("faz", First(pub))
			assert.Nil(t, err)

			time.Sleep(time.Millisecond * 100)

			got, err = m.GetKey("faz", "private")
			assert.Nil(t, err)
			assert.Equal(t, priv, got.Keys, "%s", name)

			got, err = m.GetKey("faz", "public")
			assert.Nil(t, err)
			assert.Equal(t, pub, got.Keys, "%s", name)

			err = m.DeleteKey("faz", "public")
			assert.Nil(t, err)

			time.Sleep(time.Millisecond * 100)

			ks, err = m.GetKey("faz", "public")
			assert.NotNil(t, err)
		})
	}

	err := managers["http"].AddKey("nonono", First(priv))
	pkg.AssertError(t, true, err, "%s")
}

func TestManagerKeySet(t *testing.T) {
	ks, _ := testGenerator.Generate("")
	ks.Key("private")

	for name, m := range managers {
		t.Run(fmt.Sprintf("case=%s", name), func(t *testing.T) {
			_, err := m.GetKeySet("foo")
			pkg.AssertError(t, true, err, name)

			err = m.AddKeySet("bar", ks)
			assert.Nil(t, err)

			time.Sleep(time.Millisecond * 100)

			got, err := m.GetKeySet("bar")
			assert.Nil(t, err)
			assert.Equal(t, ks.Key("public"), got.Key("public"), name)
			assert.Equal(t, ks.Key("private"), got.Key("private"), name)

			err = m.DeleteKeySet("bar")
			assert.Nil(t, err)

			time.Sleep(time.Millisecond * 100)

			_, err = m.GetKeySet("bar")
			assert.NotNil(t, err)
		})
	}

	err := managers["http"].AddKeySet("nonono", ks)
	pkg.AssertError(t, true, err, "%s")
}
