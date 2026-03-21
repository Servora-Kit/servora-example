package data

import (
	"github.com/google/uuid"

	apppb "github.com/Servora-Kit/servora/api/gen/go/application/service/v1"
	userpb "github.com/Servora-Kit/servora/api/gen/go/user/service/v1"
	"github.com/Servora-Kit/servora/app/iam/service/internal/data/ent"
	"github.com/Servora-Kit/servora/pkg/mapper"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var userMapper = mapper.NewForwardMapper(func(u *ent.User) *userpb.User {
	pbUser := &userpb.User{
		Id:            u.ID.String(),
		Username:      u.Username,
		Email:         u.Email,
		Role:          u.Role,
		Status:        u.Status,
		EmailVerified: u.EmailVerified,
		Phone:         u.Phone,
		PhoneVerified: u.PhoneVerified,
		CreatedAt:     timestamppb.New(u.CreatedAt),
		UpdatedAt:     timestamppb.New(u.UpdatedAt),
	}
	if u.EmailVerifiedAt != nil {
		pbUser.EmailVerifiedAt = timestamppb.New(*u.EmailVerifiedAt)
	}
	if u.Profile != nil {
		pbUser.Profile = profileFromJSON(u.Profile)
	}
	return pbUser
})

func profileFromJSON(m map[string]interface{}) *userpb.UserProfile {
	if m == nil {
		return nil
	}
	p := &userpb.UserProfile{}
	if v, ok := m["name"].(string); ok {
		p.Name = v
	}
	if v, ok := m["given_name"].(string); ok {
		p.GivenName = v
	}
	if v, ok := m["family_name"].(string); ok {
		p.FamilyName = v
	}
	if v, ok := m["nickname"].(string); ok {
		p.Nickname = v
	}
	if v, ok := m["picture"].(string); ok {
		p.Picture = v
	}
	if v, ok := m["gender"].(string); ok {
		p.Gender = v
	}
	if v, ok := m["birthdate"].(string); ok {
		p.Birthdate = v
	}
	if v, ok := m["zoneinfo"].(string); ok {
		p.Zoneinfo = v
	}
	if v, ok := m["locale"].(string); ok {
		p.Locale = v
	}
	return p
}

var applicationMapper = func() *mapper.CopierMapper[apppb.Application, ent.Application] {
	m := mapper.NewCopierMapper[apppb.Application, ent.Application]()
	m.AppendConverters(mapper.AllBuiltinConverters())
	m.AppendConverters(mapper.NewGenericConverterPair[uuid.UUID, string](
		func(id uuid.UUID) (string, error) { return id.String(), nil },
		func(s string) (uuid.UUID, error) { return uuid.Parse(s) },
	))
	m.AppendConverters(mapper.NewGenericConverterPair[int, int32](
		func(i int) (int32, error) { return int32(i), nil },
		func(i int32) (int, error) { return int(i), nil },
	))
	m.WithFieldMapping(map[string]string{
		"ID":               "Id",
		"ClientID":         "ClientId",
		"IDTokenLifetime":  "IdTokenLifetime",
	})
	return m
}()
