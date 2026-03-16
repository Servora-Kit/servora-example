package actor

type UserActor struct {
	id             string
	displayName    string
	email          string
	metadata       map[string]string
	organizationID string
	projectID      string
}

func NewUserActor(id, displayName, email string, metadata map[string]string) *UserActor {
	return &UserActor{
		id:          id,
		displayName: displayName,
		email:       email,
		metadata:    metadata,
	}
}

func (u *UserActor) ID() string           { return u.id }
func (u *UserActor) Type() Type           { return TypeUser }
func (u *UserActor) DisplayName() string  { return u.displayName }
func (u *UserActor) Email() string        { return u.email }

func (u *UserActor) Metadata() map[string]string {
	if u.metadata == nil {
		return map[string]string{}
	}
	return u.metadata
}

func (u *UserActor) Meta(key string) string {
	if u.metadata == nil {
		return ""
	}
	return u.metadata[key]
}

func (u *UserActor) OrganizationID() string     { return u.organizationID }
func (u *UserActor) ProjectID() string           { return u.projectID }
func (u *UserActor) SetOrganizationID(id string) { u.organizationID = id }
func (u *UserActor) SetProjectID(id string)      { u.projectID = id }
