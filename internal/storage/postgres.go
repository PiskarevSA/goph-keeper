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

	switch sec.Kind {
	case domain.KindCredentials:
		_, err = s.db.Exec(
			`INSERT INTO credentials_secrets (uuid, login, password)
			VALUES ($1,$2,$3)`,
			sec.UUID, sec.Details.Credentials.Login, sec.Details.Credentials.Password,
		)
	case domain.KindCard:
		_, err = s.db.Exec(
			`INSERT INTO card_secrets (uuid, number, holder, expires, verification_code)
			VALUES ($1,$2,$3,$4,$5)`,
			sec.UUID, sec.Details.Card.Number, sec.Details.Card.Holder,
			sec.Details.Card.Expires, sec.Details.Card.VerificationCode,
		)
	case domain.KindText:
		_, err = s.db.Exec(
			`INSERT INTO text_secrets (uuid, filename, content)
			VALUES ($1,$2,$3)`,
			sec.UUID, sec.Details.Text.Filename, sec.Details.Text.Content,
		)
	case domain.KindRaw:
		_, err = s.db.Exec(
			`INSERT INTO raw_secrets (uuid, path, filename, size)
			VALUES ($1,$2,$3,$4)`,
			sec.UUID, sec.Details.Raw.Path,
			sec.Details.Raw.Filename, sec.Details.Raw.Size,
		)
	}

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

	switch sec.Kind {
	case domain.KindCredentials:
		c := domain.Credentials{}
		_ = s.db.QueryRow(
			`SELECT login, password
			FROM credentials_secrets
			WHERE uuid=$1`,
			sec.UUID).
			Scan(&c.Login, &c.Password)
		sec.Details.Credentials = &c
	case domain.KindCard:
		c := domain.Card{}
		_ = s.db.QueryRow(
			`SELECT number, holder, expires, verification_code
			FROM card_secrets
			WHERE uuid=$1`,
			sec.UUID).
			Scan(&c.Number, &c.Holder, &c.Expires, &c.VerificationCode)
		sec.Details.Card = &c
	case domain.KindText:
		t := domain.Text{}
		_ = s.db.QueryRow(
			`SELECT filename, content
			FROM text_secrets
			WHERE uuid=$1`,
			sec.UUID).
			Scan(&t.Filename, &t.Content)
		sec.Details.Text = &t
	case domain.KindRaw:
		r := domain.Raw{}
		_ = s.db.QueryRow(
			`SELECT path, filename, size
			FROM raw_secrets
			WHERE uuid=$1`,
			sec.UUID).
			Scan(&r.Path, &r.Filename, &r.Size)
		sec.Details.Raw = &r
	}

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
	if err != nil {
		return modified, err
	}

	switch sec.Kind {
	case domain.KindCredentials:
		_, err = s.db.Exec(
			`UPDATE credentials_secrets
			SET login=$1, password=$2
			WHERE uuid=$3`,
			sec.Details.Credentials.Login,
			sec.Details.Credentials.Password,
			sec.UUID)
	case domain.KindCard:
		_, err = s.db.Exec(
			`UPDATE card_secrets
			SET number=$1, holder=$2, expires=$3, verification_code=$4
			WHERE uuid=$5`,
			sec.Details.Card.Number,
			sec.Details.Card.Holder,
			sec.Details.Card.Expires,
			sec.Details.Card.VerificationCode,
			sec.UUID)
	case domain.KindText:
		_, err = s.db.Exec(
			`UPDATE text_secrets
			SET filename=$1, content=$2
			WHERE uuid=$3`,
			sec.Details.Text.Filename,
			sec.Details.Text.Content,
			sec.UUID)
	case domain.KindRaw:
		_, err = s.db.Exec(
			`UPDATE raw_secrets
			SET path=$1, filename=$2, size=$3
			WHERE uuid=$4`,
			sec.Details.Raw.Path,
			sec.Details.Raw.Filename,
			sec.Details.Raw.Size,
			sec.UUID)
	}

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
