package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/vdyabkin01-ux/goledger/internal/user"
)

// PostgreSQL error codes we react to.
// Full list: https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	codeUniqueViolation   = "23505"
	codeInvalidTextRepr   = "22P02" // e.g. "abc" passed where a uuid is expected
	constraintUniqueEmail = "users_email_key"
	userColumns           = "id::text, email, password_hash, created_at"
)

// UserRepo is a user.Repository backed by PostgreSQL.
type UserRepo struct {
	pool *pgxpool.Pool
}

// NewUserRepo creates a UserRepo.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// Create implements user.Repository.
func (r *UserRepo) Create(ctx context.Context, email, passwordHash string) (user.User, error) {
	row := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING `+userColumns,
		email, passwordHash)

	u, err := scanUser(row)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codeUniqueViolation && pgErr.ConstraintName == constraintUniqueEmail {
		return user.User{}, user.ErrEmailTaken
	}
	if err != nil {
		return user.User{}, fmt.Errorf("insert user: %w", err)
	}
	return u, nil
}

// GetByEmail implements user.Repository.
func (r *UserRepo) GetByEmail(ctx context.Context, email string) (user.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE email = $1`, email)
	return r.get(row)
}

// GetByID implements user.Repository.
func (r *UserRepo) GetByID(ctx context.Context, id string) (user.User, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+userColumns+` FROM users WHERE id = $1`, id)
	return r.get(row)
}

func (r *UserRepo) get(row pgx.Row) (user.User, error) {
	u, err := scanUser(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, user.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == codeInvalidTextRepr {
		return user.User{}, user.ErrNotFound
	}
	if err != nil {
		return user.User{}, fmt.Errorf("select user: %w", err)
	}
	return u, nil
}

func scanUser(row pgx.Row) (user.User, error) {
	var u user.User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	return u, err
}
