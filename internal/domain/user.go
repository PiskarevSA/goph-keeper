package domain

type User struct {
	ID           int64
	Email        string
	PasswordHash string
}

type UserStorage interface {
	CreateUser(email, passwordHash string) (int64, error)
	GetUserByEmail(email string) (*User, error)
}
