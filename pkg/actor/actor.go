package actor

type Type string

const (
	TypeUser      Type = "user"
	TypeSystem    Type = "system"
	TypeAnonymous Type = "anonymous"
)

// Actor represents the identity of a request initiator.
type Actor interface {
	ID() string
	Type() Type
	DisplayName() string
}
