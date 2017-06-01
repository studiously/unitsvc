package warden_test

import (
	"log"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"context"

	"github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/ory/fosite"
	foauth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/herodot"
	"github.com/ory/hydra/firewall"
	"github.com/ory/hydra/oauth2"
	"github.com/ory/hydra/pkg"
	"github.com/ory/hydra/warden"
	"github.com/ory/hydra/warden/group"
	"github.com/ory/ladon"
	"github.com/stretchr/testify/assert"
	coauth2 "golang.org/x/oauth2"
)

var ts *httptest.Server

var wardens = map[string]firewall.Firewall{}

var ladonWarden = pkg.LadonWarden(map[string]ladon.Policy{
	"1": &ladon.DefaultPolicy{
		ID:        "1",
		Subjects:  []string{"alice", "group1"},
		Resources: []string{"matrix", "forbidden_matrix", "rn:hydra:token<.*>"},
		Actions:   []string{"create", "decide"},
		Effect:    ladon.AllowAccess,
	},
	"2": &ladon.DefaultPolicy{
		ID:        "2",
		Subjects:  []string{"siri"},
		Resources: []string{"<.*>"},
		Actions:   []string{"decide"},
		Effect:    ladon.AllowAccess,
	},
	"3": &ladon.DefaultPolicy{
		ID:        "3",
		Subjects:  []string{"group1"},
		Resources: []string{"forbidden_matrix", "rn:hydra:token<.*>"},
		Actions:   []string{"create", "decide"},
		Effect:    ladon.DenyAccess,
	},
})

var fositeStore = pkg.FositeStore()

var now = time.Now().Round(time.Second)

var tokens = pkg.Tokens(4)

func init() {
	wardens["local"] = &warden.LocalWarden{
		Warden: ladonWarden,
		L:      logrus.New(),
		OAuth2: &fosite.Fosite{
			Store: fositeStore,
			TokenIntrospectionHandlers: fosite.TokenIntrospectionHandlers{
				&foauth2.CoreValidator{
					CoreStrategy:  pkg.HMACStrategy,
					CoreStorage:   fositeStore,
					ScopeStrategy: fosite.HierarchicScopeStrategy,
				},
			},
			ScopeStrategy: fosite.HierarchicScopeStrategy,
		},
		Groups: &group.MemoryManager{
			Groups: map[string]group.Group{
				"group1": {
					ID:      "group1",
					Members: []string{"ken"},
				},
			},
		},
		Issuer:              "tests",
		AccessTokenLifespan: time.Hour,
	}

	r := httprouter.New()
	serv := &warden.WardenHandler{
		H:      herodot.NewJSONWriter(nil),
		Warden: wardens["local"],
	}
	serv.SetRoutes(r)
	ts = httptest.NewServer(r)

	url, err := url.Parse(ts.URL + warden.TokenAllowedHandlerPath)
	if err != nil {
		log.Fatalf("%s", err)
	}

	ar := fosite.NewAccessRequest(oauth2.NewSession("alice"))
	ar.GrantedScopes = fosite.Arguments{"core", "hydra.warden"}
	ar.RequestedAt = now
	ar.Client = &fosite.DefaultClient{ID: "siri"}
	ar.Session.SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour).Round(time.Second))
	fositeStore.CreateAccessTokenSession(nil, tokens[0][0], ar)

	ar2 := fosite.NewAccessRequest(oauth2.NewSession("siri"))
	ar2.GrantedScopes = fosite.Arguments{"core", "hydra.warden"}
	ar2.RequestedAt = now
	ar2.Client = &fosite.DefaultClient{ID: "bob"}
	ar2.Session.SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour).Round(time.Second))
	fositeStore.CreateAccessTokenSession(nil, tokens[1][0], ar2)

	ar3 := fosite.NewAccessRequest(oauth2.NewSession("siri"))
	ar3.GrantedScopes = fosite.Arguments{"core", "hydra.warden"}
	ar3.RequestedAt = now
	ar3.Client = &fosite.DefaultClient{ID: "doesnt-exist"}
	ar3.Session.SetExpiresAt(fosite.AccessToken, time.Now().Add(-time.Hour).Round(time.Second))
	fositeStore.CreateAccessTokenSession(nil, tokens[2][0], ar3)

	ar4 := fosite.NewAccessRequest(oauth2.NewSession("ken"))
	ar4.GrantedScopes = fosite.Arguments{"core", "hydra.warden"}
	ar4.RequestedAt = now
	ar4.Client = &fosite.DefaultClient{ID: "siri"}
	ar4.Session.SetExpiresAt(fosite.AccessToken, time.Now().Add(time.Hour).Round(time.Second))
	fositeStore.CreateAccessTokenSession(nil, tokens[3][0], ar4)

	conf := &coauth2.Config{
		Scopes:   []string{},
		Endpoint: coauth2.Endpoint{},
	}
	wardens["http"] = &warden.HTTPWarden{
		Endpoint: url,
		Client: conf.Client(coauth2.NoContext, &coauth2.Token{
			AccessToken: tokens[1][1],
			Expiry:      time.Now().Add(time.Hour),
			TokenType:   "bearer",
		}),
	}
}

func TestActionAllowed(t *testing.T) {
	for n, w := range wardens {
		for k, c := range []struct {
			token     string
			req       *firewall.TokenAccessRequest
			scopes    []string
			expectErr bool
			assert    func(*firewall.Context)
		}{
			{
				token:     "invalid",
				req:       &firewall.TokenAccessRequest{},
				scopes:    []string{},
				expectErr: true,
			},
			{
				token:     tokens[0][1],
				req:       &firewall.TokenAccessRequest{},
				scopes:    []string{"core"},
				expectErr: true,
			},
			{
				token:     tokens[0][1],
				req:       &firewall.TokenAccessRequest{},
				scopes:    []string{"foo"},
				expectErr: true,
			},
			{
				token: tokens[0][1],
				req: &firewall.TokenAccessRequest{
					Resource: "matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				scopes:    []string{"foo"},
				expectErr: true,
			},
			{
				token: tokens[0][1],
				req: &firewall.TokenAccessRequest{
					Resource: "matrix",
					Action:   "delete",
					Context:  ladon.Context{},
				},
				scopes:    []string{"core"},
				expectErr: true,
			},
			{
				token: tokens[0][1],
				req: &firewall.TokenAccessRequest{
					Resource: "matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				scopes:    []string{"illegal"},
				expectErr: true,
			},
			{
				token: tokens[0][1],
				req: &firewall.TokenAccessRequest{
					Resource: "matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				scopes:    []string{"core"},
				expectErr: false,
				assert: func(c *firewall.Context) {
					assert.Equal(t, "siri", c.Audience)
					assert.Equal(t, "alice", c.Subject)
					assert.Equal(t, "tests", c.Issuer)
					assert.Equal(t, now.Add(time.Hour).Unix(), c.ExpiresAt.Unix())
					assert.Equal(t, now.Unix(), c.IssuedAt.Unix())
				},
			},
			{
				token: tokens[3][1],
				req: &firewall.TokenAccessRequest{
					Resource: "forbidden_matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				scopes:    []string{"core"},
				expectErr: true,
			},
			{
				token: tokens[3][1],
				req: &firewall.TokenAccessRequest{
					Resource: "matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				scopes:    []string{"core"},
				expectErr: false,
				assert: func(c *firewall.Context) {
					assert.Equal(t, "siri", c.Audience)
					assert.Equal(t, "ken", c.Subject)
					assert.Equal(t, "tests", c.Issuer)
					assert.Equal(t, now.Add(time.Hour).Unix(), c.ExpiresAt.Unix())
					assert.Equal(t, now.Unix(), c.IssuedAt.Unix())
				},
			},
		} {
			ctx, err := w.TokenAllowed(context.Background(), c.token, c.req, c.scopes...)
			pkg.AssertError(t, c.expectErr, err, "ActionAllowed case", n, k)
			if err == nil && c.assert != nil {
				c.assert(ctx)
			}
		}
	}
}

func TestAllowed(t *testing.T) {
	for n, w := range wardens {
		for k, c := range []struct {
			req       *firewall.AccessRequest
			expectErr bool
			assert    func(*firewall.Context)
		}{
			{
				req: &firewall.AccessRequest{
					Subject:  "alice",
					Resource: "other-thing",
					Action:   "create",
					Context:  ladon.Context{},
				},
				expectErr: true,
			},
			{
				req: &firewall.AccessRequest{
					Subject:  "alice",
					Resource: "matrix",
					Action:   "delete",
					Context:  ladon.Context{},
				},
				expectErr: true,
			},
			{
				req: &firewall.AccessRequest{
					Subject:  "alice",
					Resource: "matrix",
					Action:   "create",
					Context:  ladon.Context{},
				},
				expectErr: false,
			},
		} {
			err := w.IsAllowed(context.Background(), c.req)
			pkg.AssertError(t, c.expectErr, err, "TestAllowed case", n, k)
			t.Logf("Passed test case %d\n", k)
		}
		t.Logf("Passed tests %s\n", n)
	}

}
