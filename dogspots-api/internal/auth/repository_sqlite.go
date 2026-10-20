package auth

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) (*SQLiteRepository, error) {
	return &SQLiteRepository{db: db}, nil
}

func (r *SQLiteRepository) CreateUser(ctx context.Context, user UserCreateParams) (User, error) {
	query := `INSERT INTO users (id, email, password_hash, name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := r.db.ExecContext(ctx, query,
		user.ID.String(),
		user.Email,
		user.PasswordHash,
		user.Name,
		user.CreatedAt.UTC().Format(time.RFC3339Nano),
		user.UpdatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return User{}, err
	}
	return r.GetUserByID(ctx, user.ID)
}

func (r *SQLiteRepository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = ? LIMIT 1`, email)
	return scanUser(row)
}

func (r *SQLiteRepository) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id = ? LIMIT 1`, id.String())
	return scanUser(row)
}

func (r *SQLiteRepository) UpdateUser(ctx context.Context, id uuid.UUID, name string) (User, error) {
	_, err := r.db.ExecContext(ctx, "UPDATE users SET name = ?, updated_at = ? WHERE id = ?",
		name,
		time.Now().UTC().Format(time.RFC3339Nano),
		id.String(),
	)
	if err != nil {
		return User{}, err
	}
	return r.GetUserByID(ctx, id)
}

func (r *SQLiteRepository) CreateRefreshToken(ctx context.Context, token RefreshTokenParams) (RefreshToken, error) {
	_, err := r.db.ExecContext(ctx, `INSERT INTO refresh_tokens (id, user_id, token_hash, family_id, expires_at, created_at, is_revoked)
	VALUES (?, ?, ?, ?, ?, ?, 0)`,
		token.ID.String(),
		token.UserID.String(),
		token.TokenHash,
		token.FamilyID.String(),
		token.ExpiresAt.UTC().Format(time.RFC3339Nano),
		token.CreatedAt.UTC().Format(time.RFC3339Nano),
	)
	if err != nil {
		return RefreshToken{}, err
	}
	return r.GetRefreshToken(ctx, token.ID)
}

func (r *SQLiteRepository) GetRefreshToken(ctx context.Context, id uuid.UUID) (RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, token_hash, family_id, expires_at, created_at, is_revoked FROM refresh_tokens WHERE id = ? LIMIT 1`, id.String())
	return scanRefreshToken(row)
}

func (r *SQLiteRepository) GetRefreshTokenByHash(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, user_id, token_hash, family_id, expires_at, created_at, is_revoked FROM refresh_tokens WHERE token_hash = ? LIMIT 1`, tokenHash)
	return scanRefreshToken(row)
}

func (r *SQLiteRepository) ListRefreshTokensByFamily(ctx context.Context, familyID uuid.UUID) ([]RefreshToken, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, user_id, token_hash, family_id, expires_at, created_at, is_revoked FROM refresh_tokens WHERE family_id = ?`, familyID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []RefreshToken
	for rows.Next() {
		token, err := scanRefreshTokenFromRows(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, rows.Err()
}

func (r *SQLiteRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET is_revoked = 1 WHERE id = ?`, id.String())
	return err
}

func (r *SQLiteRepository) RevokeAllTokensByFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refresh_tokens SET is_revoked = 1 WHERE family_id = ?`, familyID.String())
	return err
}

func scanUser(row *sql.Row) (User, error) {
	var user User
	var idStr, email, passwordHash, name, createdAt, updatedAt string
	if err := row.Scan(&idStr, &email, &passwordHash, &name, &createdAt, &updatedAt); err != nil {
		return User{}, err
	}
	userID, err := uuid.Parse(idStr)
	if err != nil {
		return User{}, err
	}
	user.ID = userID
	user.Email = email
	user.PasswordHash = passwordHash
	user.Name = name
	user.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return User{}, err
	}
	user.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return User{}, err
	}
	return user, nil
}

func scanRefreshToken(row *sql.Row) (RefreshToken, error) {
	var token RefreshToken
	var idStr, userIDStr, tokenHash, familyIDStr, expiresAt, createdAt string
	var revoked int
	if err := row.Scan(&idStr, &userIDStr, &tokenHash, &familyIDStr, &expiresAt, &createdAt, &revoked); err != nil {
		return RefreshToken{}, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RefreshToken{}, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return RefreshToken{}, err
	}
	familyID, err := uuid.Parse(familyIDStr)
	if err != nil {
		return RefreshToken{}, err
	}
	token.ID = id
	token.UserID = userID
	token.TokenHash = tokenHash
	token.FamilyID = familyID
	token.ExpiresAt, err = time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return RefreshToken{}, err
	}
	token.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return RefreshToken{}, err
	}
	token.IsRevoked = revoked == 1
	return token, nil
}

func scanRefreshTokenFromRows(rows *sql.Rows) (RefreshToken, error) {
	var token RefreshToken
	var idStr, userIDStr, tokenHash, familyIDStr, expiresAt, createdAt string
	var revoked int
	if err := rows.Scan(&idStr, &userIDStr, &tokenHash, &familyIDStr, &expiresAt, &createdAt, &revoked); err != nil {
		return RefreshToken{}, err
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return RefreshToken{}, err
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return RefreshToken{}, err
	}
	familyID, err := uuid.Parse(familyIDStr)
	if err != nil {
		return RefreshToken{}, err
	}
	token.ID = id
	token.UserID = userID
	token.TokenHash = tokenHash
	token.FamilyID = familyID
	token.ExpiresAt, err = time.Parse(time.RFC3339Nano, expiresAt)
	if err != nil {
		return RefreshToken{}, err
	}
	token.CreatedAt, err = time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return RefreshToken{}, err
	}
	token.IsRevoked = revoked == 1
	return token, nil
}
