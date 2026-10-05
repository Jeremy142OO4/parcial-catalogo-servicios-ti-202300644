package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID         int64  `json:"id"`
	FullName   string `json:"full_name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	IsActive   bool   `json:"is_active"`
	PositionID int64  `json:"position_id"`
}

type Session struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type Service struct {
	pool       *pgxpool.Pool
	sessionTTL time.Duration
}

func NewService(pool *pgxpool.Pool, sessionTTL time.Duration) *Service {
	if sessionTTL <= 0 {
		sessionTTL = 8 * time.Hour
	}
	return &Service{pool: pool, sessionTTL: sessionTTL}
}

func HashPassword(password string) (string, error) {
	if len(password) < 8 {
		return "", fmt.Errorf("la contraseña debe tener al menos 8 caracteres")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func (s *Service) Login(ctx context.Context, identifier, password string) (*Session, error) {
	identifier = strings.TrimSpace(identifier)
	var user User
	var passwordHash string
	err := s.pool.QueryRow(ctx, `
		SELECT id, full_name, username, email, password_hash, role, is_active, position_id
		FROM app_users
		WHERE (LOWER(username) = LOWER($1) OR LOWER(email) = LOWER($1))`, identifier).
		Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &passwordHash, &user.Role, &user.IsActive, &user.PositionID)
	if err != nil || !user.IsActive || bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password)) != nil {
		return nil, fmt.Errorf("credenciales inválidas")
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	expiresAt := time.Now().Add(s.sessionTTL)
	if _, err := s.pool.Exec(ctx, `
		INSERT INTO sessions (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`, user.ID, hashToken(token), expiresAt); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return &Session{Token: token, User: user}, nil
}

func (s *Service) Authenticate(ctx context.Context, token string) (*User, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, fmt.Errorf("sesión requerida")
	}
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT u.id, u.full_name, u.username, u.email, u.role, u.is_active, u.position_id
		FROM sessions s
		JOIN app_users u ON u.id = s.user_id
		WHERE s.token_hash = $1 AND s.expires_at > NOW() AND u.is_active = TRUE`, hashToken(token)).
		Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.PositionID)
	if err != nil {
		return nil, fmt.Errorf("sesión inválida")
	}
	return &user, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE token_hash = $1`, hashToken(token))
	return err
}

func (s *Service) CreateUser(ctx context.Context, positionID int64, fullName, username, email, password, role string) (User, error) {
	var positionActive bool
	if err := s.pool.QueryRow(ctx, `SELECT is_active FROM positions WHERE id = $1`, positionID).Scan(&positionActive); err != nil || !positionActive {
		return User{}, fmt.Errorf("el puesto no existe o está inactivo")
	}
	hash, err := HashPassword(password)
	if err != nil {
		return User{}, err
	}
	var user User
	err = s.pool.QueryRow(ctx, `
		INSERT INTO app_users (position_id, full_name, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, full_name, username, email, role, is_active, position_id`,
		positionID, strings.TrimSpace(fullName), strings.TrimSpace(username), strings.TrimSpace(email), hash, role).
		Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.PositionID)
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return user, nil
}

func (s *Service) SetUserActive(ctx context.Context, userID int64, active bool) error {
	if !active {
		var assigned bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM service_assignments WHERE responsible_user_id = $1)`, userID).Scan(&assigned); err != nil {
			return err
		}
		if assigned {
			return fmt.Errorf("no se puede desactivar: el usuario tiene servicios asignados")
		}
	}
	_, err := s.pool.Exec(ctx, `UPDATE app_users SET is_active = $1 WHERE id = $2`, active, userID)
	return err
}

func (s *Service) UpdateUser(ctx context.Context, userID int64, positionID *int64, fullName, email, role *string) (User, error) {
	if role != nil && *role != "admin" && *role != "consulta" {
		return User{}, fmt.Errorf("el rol debe ser admin o consulta")
	}
	if positionID != nil {
		var incompatible bool
		if err := s.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM service_assignments a JOIN positions p ON p.id=$2 WHERE a.responsible_user_id=$1 AND a.section_id<>p.section_id)`, userID, *positionID).Scan(&incompatible); err != nil {
			return User{}, err
		}
		if incompatible {
			return User{}, fmt.Errorf("no se puede cambiar de sección: el usuario tiene servicios asignados")
		}
		var active bool
		if err := s.pool.QueryRow(ctx, `SELECT is_active FROM positions WHERE id = $1`, *positionID).Scan(&active); err != nil || !active {
			return User{}, fmt.Errorf("el puesto no existe o está inactivo")
		}
	}
	var user User
	err := s.pool.QueryRow(ctx, `
		UPDATE app_users SET
			position_id = COALESCE($1, position_id),
			full_name = COALESCE(NULLIF($2, ''), full_name),
			email = COALESCE(NULLIF($3, ''), email),
			role = COALESCE(NULLIF($4, ''), role)
		WHERE id = $5
		RETURNING id, full_name, username, email, role, is_active, position_id`, positionID, valueOrEmpty(fullName), valueOrEmpty(email), valueOrEmpty(role), userID).
		Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.PositionID)
	if err != nil {
		return User{}, fmt.Errorf("actualizar usuario: %w", err)
	}
	return user, nil
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (s *Service) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, full_name, username, email, role, is_active, position_id
		FROM app_users ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	users := make([]User, 0)
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.FullName, &user.Username, &user.Email, &user.Role, &user.IsActive, &user.PositionID); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func randomToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	return hex.EncodeToString(buffer), nil
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
