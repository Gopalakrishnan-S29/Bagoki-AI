
package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID           string
	Name         string
	Email        string
	PasswordHash string
}

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(
	ctx context.Context,
	name string,
	email string,
	passwordHash string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`INSERT INTO users (name, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id, name, email, password_hash`,
		name,
		email,
		passwordHash,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(
	ctx context.Context,
	email string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, email, password_hash
		 FROM users
		 WHERE email = $1`,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
func (r *UserRepository) GetUserByID(
	ctx context.Context,
	id string,
) (*User, error) {
	user := &User{}

	err := r.db.QueryRow(
		ctx,
		`SELECT id, name, email, password_hash
		 FROM users
		 WHERE id = $1`,
		id,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}