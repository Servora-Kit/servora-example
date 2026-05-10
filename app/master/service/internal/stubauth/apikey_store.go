// Package stubauth provides TEST-ONLY stub backends for the lighthouse
// demo authn middleware (jwt + apikey). It is wired only by the
// demo/authn branch of servora-example to exercise the engine-agnostic
// authn dispatcher end-to-end. DO NOT use in production.
package stubauth

import (
	"context"
	"errors"

	"github.com/Servora-Kit/servora/core/actor"
	"github.com/Servora-Kit/servora/security/authn/apikey"
)

// DemoAPIKey is the single hard-coded API key the lighthouse demo accepts.
// Tests / curl scripts pass it via the `X-API-Key` header.
const DemoAPIKey = "lighthouse-demo-key"

// NewAPIKeyStore returns an apikey.Store backed by an in-memory map with
// a single demo key mapped to a service actor. It is intentionally trivial:
// the lighthouse demo only cares about success/failure paths, not realistic
// key management.
func NewAPIKeyStore() apikey.Store {
	return &store{m: map[string]actor.Actor{
		DemoAPIKey: actor.NewServiceActor("lighthouse-svc", "Lighthouse Demo Service"),
	}}
}

type store struct {
	m map[string]actor.Actor
}

func (s *store) Lookup(_ context.Context, key string) (actor.Actor, error) {
	a, ok := s.m[key]
	if !ok {
		return nil, errors.New("apikey: unknown key")
	}
	return a, nil
}
