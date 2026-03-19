package data

import (
	"time"

	"github.com/Servora-Kit/servora/app/iam/service/internal/biz/entity"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/pkg/mapper"
)

var userMapper = mapper.NewForwardMapper(func(u *ent.User) *entity.User {
	e := &entity.User{
		ID:              u.ID.String(),
		Username:        u.Username,
		Email:           u.Email,
		Password:        u.Password,
		Phone:           u.Phone,
		PhoneVerified:   u.PhoneVerified,
		Role:            u.Role,
		Status:          u.Status,
		EmailVerified:   u.EmailVerified,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
		UpdatedAt:       u.UpdatedAt,
	}
	if u.Profile != nil {
		if v, ok := u.Profile["name"].(string); ok {
			e.Profile.Name = v
		}
		if v, ok := u.Profile["given_name"].(string); ok {
			e.Profile.GivenName = v
		}
		if v, ok := u.Profile["family_name"].(string); ok {
			e.Profile.FamilyName = v
		}
		if v, ok := u.Profile["nickname"].(string); ok {
			e.Profile.Nickname = v
		}
		if v, ok := u.Profile["picture"].(string); ok {
			e.Profile.Picture = v
		}
		if v, ok := u.Profile["gender"].(string); ok {
			e.Profile.Gender = v
		}
		if v, ok := u.Profile["birthdate"].(string); ok {
			e.Profile.Birthdate = v
		}
		if v, ok := u.Profile["zoneinfo"].(string); ok {
			e.Profile.Zoneinfo = v
		}
		if v, ok := u.Profile["locale"].(string); ok {
			e.Profile.Locale = v
		}
	}
	return e
})

var applicationMapper = mapper.NewForwardMapper(func(a *ent.Application) *entity.Application {
	return &entity.Application{
		ID:               a.ID.String(),
		ClientID:         a.ClientID,
		ClientSecretHash: a.ClientSecretHash,
		Name:             a.Name,
		RedirectURIs:     a.RedirectUris,
		Scopes:           a.Scopes,
		GrantTypes:       a.GrantTypes,
		ApplicationType:  a.ApplicationType,
		AccessTokenType:  a.AccessTokenType,
		Type:             a.Type,
		IDTokenLifetime:  time.Duration(a.IDTokenLifetime) * time.Second,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
})
