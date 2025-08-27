package domain

import "time"

type SecretKind int

const (
	KindUnspecified SecretKind = iota
	KindCredentials
	KindCard
	KindText
	KindRaw
)

type Secret struct {
	UUID     string
	Name     string
	Kind     SecretKind
	Created  time.Time
	Modified time.Time
	Details  SecretDetails
}

type SecretDetails struct {
	Description string
	Credentials *Credentials
	Card        *Card
	Text        *Text
	Raw         *Raw
}

type Credentials struct {
	Login    string
	Password string
}

type Card struct {
	Number           string
	Holder           string
	Expires          string
	VerificationCode string
}

type Text struct {
	Filename string
	Content  string
}

type Raw struct {
	Path     string
	Size     int64
	Filename string
}

type SecretStorage interface {
	ListSecrets(userID int64) ([]Secret, error)
	CreateSecret(userID int64, s Secret) (Secret, error)
	GetSecret(userID int64, uuid string) (*Secret, error)
	UpdateSecret(userID int64, s Secret) (time.Time, error)
	DeleteSecret(userID int64, uuid string) error
}
