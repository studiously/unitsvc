package oauth2

import (
	"crypto/rsa"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/sessions"
	"github.com/ory/fosite"
	"github.com/ory/fosite/handler/openid"
	ejwt "github.com/ory/fosite/token/jwt"
	"github.com/ory/hydra/jwk"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
)

const (
	ConsentChallengeKey = "hydra.consent.challenge"
	ConsentEndpointKey  = "hydra.consent.response"
)

type DefaultConsentStrategy struct {
	Issuer string

	DefaultIDTokenLifespan   time.Duration
	DefaultChallengeLifespan time.Duration
	KeyManager               jwk.Manager
}

func (s *DefaultConsentStrategy) ValidateResponse(a fosite.AuthorizeRequester, token string, session *sessions.Session) (claims *Session, err error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.Errorf("Unexpected signing method: %v", t.Header["alg"])
		}

		pk, err := s.KeyManager.GetKey(ConsentEndpointKey, "public")
		if err != nil {
			return nil, err
		}

		rsaKey, ok := jwk.First(pk.Keys).Key.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("Could not convert to RSA Private Key")
		}
		return rsaKey, nil
	})
	if err != nil {
		return nil, errors.Wrap(err, "The consent response is not a valid JSON Web Token")
	}

	// make sure to use MapClaims since that is the default..
	jwtClaims, ok := t.Claims.(jwt.MapClaims)
	if err != nil || !ok {
		return nil, errors.Errorf("Couldn't parse token: %v", err)
	} else if !t.Valid {
		return nil, errors.Errorf("Token is invalid")
	}

	if j, ok := session.Values["consent_jti"]; !ok {
		return nil, errors.Errorf("Session cookie is missing anti-replay token")
	} else if js, ok := j.(string); !ok {
		return nil, errors.Errorf("Session cookie anti-replay value is not a string")
	} else if js != ejwt.ToString(jwtClaims["jti"]) {
		return nil, errors.Errorf("Session cookie anti-replay value does not match value from consent response")
	}
	delete(session.Values, "jti")

	if time.Now().After(ejwt.ToTime(jwtClaims["exp"])) {
		return nil, errors.Errorf("Token expired")
	}

	if ejwt.ToString(jwtClaims["aud"]) != a.GetClient().GetID() {
		return nil, errors.Errorf("Audience mismatch")
	}

	subject := ejwt.ToString(jwtClaims["sub"])
	scopes := toStringSlice(jwtClaims["scp"])
	for _, scope := range scopes {
		a.GrantScope(scope)
	}

	var idExt map[string]interface{}
	var atExt map[string]interface{}
	if ext, ok := jwtClaims["id_ext"].(map[string]interface{}); ok {
		idExt = ext
	}
	if ext, ok := jwtClaims["at_ext"].(map[string]interface{}); ok {
		atExt = ext
	}

	// add key id to session headers
	extHeader := map[string]interface{}{
		"kid": "public",
	}

	return &Session{
		DefaultSession: &openid.DefaultSession{
			Claims: &ejwt.IDTokenClaims{
				Audience:  a.GetClient().GetID(),
				Subject:   subject,
				Issuer:    s.Issuer,
				IssuedAt:  time.Now(),
				ExpiresAt: time.Now().Add(s.DefaultIDTokenLifespan),
				Extra:     idExt,
			},
			Headers: &ejwt.Headers{extHeader},
			Subject: subject,
		},
		Extra: atExt,
	}, err

}

func toStringSlice(i interface{}) []string {
	if r, ok := i.([]string); ok {
		return r
	} else if r, ok := i.(fosite.Arguments); ok {
		return r
	} else if r, ok := i.([]interface{}); ok {
		ret := make([]string, 0)
		for _, y := range r {
			s, ok := y.(string)
			if ok {
				ret = append(ret, s)
			}
		}
		return ret
	}
	return []string{}
}

func (s *DefaultConsentStrategy) IssueChallenge(authorizeRequest fosite.AuthorizeRequester, redirectURL string, session *sessions.Session) (string, error) {
	token := jwt.New(jwt.SigningMethodRS256)
	jti := uuid.New()
	token.Claims = jwt.MapClaims{
		"jti":   jti,
		"scp":   authorizeRequest.GetRequestedScopes(),
		"aud":   authorizeRequest.GetClient().GetID(),
		"exp":   time.Now().Add(s.DefaultChallengeLifespan).Unix(),
		"redir": redirectURL,
	}

	session.Values["consent_jti"] = jti
	ks, err := s.KeyManager.GetKey(ConsentChallengeKey, "private")
	if err != nil {
		return "", errors.WithStack(err)
	}

	rsaKey, ok := jwk.First(ks.Keys).Key.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("Could not convert to RSA Private Key")
	}

	var signature, encoded string
	if encoded, err = token.SigningString(); err != nil {
		return "", errors.WithStack(err)
	} else if signature, err = token.Method.Sign(encoded, rsaKey); err != nil {
		return "", errors.WithStack(err)
	}

	return fmt.Sprintf("%s.%s", encoded, signature), nil

}
