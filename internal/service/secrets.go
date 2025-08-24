package service

import (
	"time"

	"github.com/PiskarevSA/goph-keeper/internal/domain"
)

type SecretsService struct {
	storage domain.SecretStorage
}

func NewSecretsService(s domain.SecretStorage) *SecretsService {
	return &SecretsService{storage: s}
}

func (s *SecretsService) List(userID int64) ([]domain.Secret, error) {
	return s.storage.ListSecrets(userID)
}

func (s *SecretsService) Create(userID int64, sec domain.Secret,
) (domain.Secret, error) {
	return s.storage.CreateSecret(userID, sec)
}

func (s *SecretsService) Get(userID int64, uuid string,
) (*domain.Secret, error) {
	return s.storage.GetSecret(userID, uuid)
}

func (s *SecretsService) Update(userID int64, sec domain.Secret,
) (time.Time, error) {
	return s.storage.UpdateSecret(userID, sec)
}

func (s *SecretsService) Delete(userID int64, uuid string) error {
	return s.storage.DeleteSecret(userID, uuid)
}
