package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/dexciuq/yummy-express-backend/internal/validator"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
	AnonymousUser     = &User{}
)

type User struct {
	ID          int64     `json:"id"`
	FirstName   string    `json:"firstname"`
	LastName    string    `json:"lastname"`
	PhoneNumber string    `json:"phone_number"`
	Email       string    `json:"email"`
	Password    password  `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	Role_ID     int64     `json:"role_id"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type password struct {
	plaintext *string
	hash      []byte
}

type UserModel struct {
	DB *sql.DB
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}
	p.plaintext = &plaintextPassword
	p.hash = hash
	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}
	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateUser(v *validator.Validator, user *User) {
	//v.Check(user.Name != "", "name", "must be provided")
	//v.Check(len(user.Name) <= 20, "name", "must not be more than 20 bytes long")
	ValidateEmail(v, user.Email)
	if user.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *user.Password.plaintext)
	}
	if user.Password.hash == nil {
		panic("missing password hash for user")
	}
}

func (u UserModel) Insert(user *User) error {
	query := `
	INSERT INTO users (firstname, lastname, phone_number, email, password_hash, role_id)
	VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING id, created_at`

	args := []any{user.FirstName, user.LastName, user.PhoneNumber, user.Email, user.Password.hash, user.Role_ID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		switch {
		case err.Error() == `pq: повторяющееся значение ключа нарушает ограничение уникальности "users_email_key"` || err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (u UserModel) GetById(id int64) (*User, error) {
	query := `
	SELECT id, firstname, lastname, phone_number, email, password_hash, created_at, role_id
	FROM users
	WHERE id = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role_ID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u UserModel) GetByEmail(email string) (*User, error) {
	query := `
	SELECT id, firstname, lastname, phone_number, email, password_hash, created_at, role_id
	FROM users
	WHERE email = $1`

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.PhoneNumber,
		&user.Email,
		&user.Password.hash,
		&user.CreatedAt,
		&user.Role_ID,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (u UserModel) Update(user *User) error {
	query := `
	UPDATE users
	SET firstname = $1, lastname = $2, phone_number = $3, email = $4, password_hash = $5, role_id = $6
	WHERE id = $7
	RETURNING id`

	args := []any{
		user.FirstName,
		user.LastName,
		user.PhoneNumber,
		user.Email,
		user.Password.hash,
		user.Role_ID,
		user.ID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := u.DB.QueryRowContext(ctx, query, args...).Scan(&user.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: повторяющееся значение ключа нарушает ограничение уникальности "users_email_key"` || err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (u UserModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `DELETE FROM users
		WHERE id = $1`

	result, err := u.DB.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}
