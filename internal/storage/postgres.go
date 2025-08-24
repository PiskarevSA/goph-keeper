package storage

import (
	"database/sql"
	"errors"
	"time"

	"github.com/PiskarevSA/goph-keeper/internal/domain"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage(db *sql.DB) *PostgresStorage {
	return &PostgresStorage{db: db}
}

// UserStorage
func (s *PostgresStorage) CreateUser(email, passwordHash string,
) (int64, error) {
	var id int64
	err := s.db.QueryRow(
		`INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id`,
		email, passwordHash,
	).Scan(&id)
	return id, err
}

func (s *PostgresStorage) GetUserByEmail(email string,
) (*domain.User, error) {
	row := s.db.QueryRow(
		`SELECT id, email, password_hash
		FROM users
		WHERE email=$1`,
		email)
	u := domain.User{}
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &u, err
}

// SecretStorage
func (s *PostgresStorage) ListSecrets(userID int64,
) ([]domain.Secret, error) {
	rows, err := s.db.Query(
		`SELECT uuid, name, kind, created, modified, description
		FROM secrets
		WHERE user_id=$1`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Secret
	for rows.Next() {
		var sec domain.Secret
		var kind int

		err = rows.Scan(&sec.UUID, &sec.Name, &kind,
			&sec.Created, &sec.Modified, &sec.Details.Description)
		if err != nil {
			return nil, err
		}
		sec.Kind = domain.SecretKind(kind)
		res = append(res, sec)
	}
	return res, nil
}

func (s *PostgresStorage) CreateSecret(userID int64, sec domain.Secret,
) (domain.Secret, error) {
	if sec.UUID == "" {
		sec.UUID = uuid.New().String()
	}

	err := s.db.QueryRow(
		`INSERT INTO secrets (uuid, user_id, name, kind, created, modified, description) 
		 VALUES ($1, $2, $3, $4, NOW(), NOW(), $5)
		 RETURNING created, modified`,
		sec.UUID, userID, sec.Name, int(sec.Kind), sec.Details.Description,
	).Scan(&sec.Created, &sec.Modified)

	return sec, err
}

func (s *PostgresStorage) GetSecret(userID int64, uuid string,
) (*domain.Secret, error) {
	row := s.db.QueryRow(
		`SELECT uuid, name, kind, created, modified, description
		FROM secrets
		WHERE user_id=$1 AND uuid=$2`,
		userID, uuid)
	var sec domain.Secret
	var kind int
	err := row.Scan(&sec.UUID, &sec.Name, &kind,
		&sec.Created, &sec.Modified, &sec.Details.Description)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	sec.Kind = domain.SecretKind(kind)
	return &sec, err
}

func (s *PostgresStorage) UpdateSecret(userID int64, sec domain.Secret,
) (time.Time, error) {
	var modified time.Time
	err := s.db.QueryRow(
		`UPDATE secrets 
         SET name=$1, kind=$2, modified=NOW(), description=$3 
         WHERE user_id=$4 AND uuid=$5 
         RETURNING modified`,
		sec.Name, int(sec.Kind), sec.Details.Description, userID, sec.UUID,
	).Scan(&modified)
	return modified, err
}

func (s *PostgresStorage) DeleteSecret(userID int64, uuid string) error {
	res, err := s.db.Exec(`DELETE FROM secrets
	WHERE user_id=$1
	AND uuid=$2`, userID, uuid)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("not found")
	}
	return nil
}
