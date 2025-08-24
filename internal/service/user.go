package service

import (
	"errors"

	"github.com/PiskarevSA/goph-keeper/internal/domain"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	storage domain.UserStorage
}

func NewUserService(s domain.UserStorage) *UserService {
	return &UserService{storage: s}
}

func (s *UserService) Register(email, password string) (int64, error) {
	hash, _ := bcrypt.GenerateFromPassword(
		[]byte(password), bcrypt.DefaultCost)
	return s.storage.CreateUser(email, string(hash))
}

func (s *UserService) Authenticate(email, password string,
) (*domain.User, error) {
	user, err := s.storage.GetUserByEmail(email)
	if user == nil || err != nil {
		return nil, errors.New("invalid credentials")
	}
	if bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	) != nil {
		return nil, errors.New("invalid credentials")
	}
	return user, nil
}
