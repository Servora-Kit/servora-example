package biz

import (
	"errors"

	"github.com/google/wire"
)

// ErrNotFound is a sentinel error returned by data layer when a record is not found.
// Biz layer uses errors.Is(err, ErrNotFound) to distinguish "not found" from other failures.
var ErrNotFound = errors.New("record not found")

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewAuthnUsecase, NewUserUsecase, NewApplicationUsecase)
