package server

import (
	"context"
	"testing"
	"time"

	masterpb "github.com/Servora-Kit/servora-example/api/gen/go/master/service/v1"
	"github.com/Servora-Kit/servora-example/app/master/service/internal/stubauth"
	"github.com/Servora-Kit/servora/security/authn"
	"github.com/Servora-Kit/servora/security/authn/apikey"
	authjwt "github.com/Servora-Kit/servora/security/authn/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	gojwt "github.com/golang-jwt/jwt/v5"
)

type authnTestTransport struct {
	op      string
	headers map[string]string
}

func (t *authnTestTransport) Kind() transport.Kind            { return transport.KindHTTP }
func (t *authnTestTransport) Endpoint() string                { return "" }
func (t *authnTestTransport) Operation() string               { return t.op }
func (t *authnTestTransport) RequestHeader() transport.Header { return authnTestHeader(t.headers) }
func (t *authnTestTransport) ReplyHeader() transport.Header   { return authnTestHeader{} }

type authnTestHeader map[string]string

func (h authnTestHeader) Get(key string) string      { return h[key] }
func (h authnTestHeader) Set(key, value string)      { h[key] = value }
func (h authnTestHeader) Add(key, value string)      { h[key] = value }
func (h authnTestHeader) Keys() []string             { return nil }
func (h authnTestHeader) Values(key string) []string { return nil }

func TestExampleAuthnJWT(t *testing.T) {
	signer, verifier := stubauth.SharedKeypair()
	raw, err := signer.Sign(gojwt.MapClaims{
		"sub":  "lighthouse-user",
		"name": "Lighthouse User",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(time.Hour).Unix(),
	})
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	mw := authn.Server(
		authn.Multi(
			authn.Named(authjwt.Scheme, authjwt.NewAuthenticator(authjwt.WithVerifier(verifier))),
			authn.Named(apikey.Scheme, apikey.NewAuthenticator(apikey.WithStore(stubauth.NewAPIKeyStore()))),
		),
		authn.WithRulesFuncs(masterpb.AuthnRules),
	)
	ctx := transport.NewServerContext(context.Background(), &authnTestTransport{
		op:      "/master.service.v1.MasterService/Hello",
		headers: map[string]string{"Authorization": "Bearer " + raw},
	})

	_, err = mw(func(ctx context.Context, _ any) (any, error) {
		sub, ok := authjwt.SubjectFrom(ctx)
		if !ok || sub != "lighthouse-user" {
			t.Fatalf("SubjectFrom = (%q,%v), want (lighthouse-user,true)", sub, ok)
		}
		authType, ok := authn.AuthTypeFrom(ctx)
		if !ok || authType != authjwt.Scheme {
			t.Fatalf("AuthTypeFrom = (%q,%v), want (jwt,true)", authType, ok)
		}
		return "ok", nil
	})(ctx, nil)
	if err != nil {
		t.Fatalf("authn middleware returned error: %v", err)
	}
}
