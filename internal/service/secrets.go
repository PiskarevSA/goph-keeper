package service

import (
	"fmt"
	"os"
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

func (s *SecretsService) SaveRawFile(
	userID int64, uuid string, path string,
) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}, err
	}

	sec, err := s.storage.GetSecret(userID, uuid)
	if err != nil {
		return time.Time{}, err
	}
	if sec == nil {
		return time.Time{}, fmt.Errorf("secret not found")
	}

	sec.Details.Raw = &domain.Raw{
		Path:     path,
		Size:     info.Size(),
		Filename: sec.Details.Raw.Filename,
	}

	modified, err := s.storage.UpdateSecret(userID, *sec)
	if err != nil {
		return time.Time{}, err
	}

	return modified, nil
}
