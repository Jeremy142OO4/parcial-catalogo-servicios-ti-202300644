package desk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
)

type Repository struct{ pool *pgxpool.Pool }
type Service struct{ repo *Repository }

func NewService(p *pgxpool.Pool) *Service { return &Service{&Repository{p}} }

type Error struct {
	Status  int
	Message string
}

func (e *Error) Error() string { return e.Message }
func bad(s string) error       { return &Error{400, s} }
func forbidden() error         { return &Error{403, "operación no autorizada"} }
func conflict(s string) error  { return &Error{409, s} }

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func objects(ctx context.Context, q queryer, sql string, args ...any) ([]map[string]any, error) {
	rows, err := q.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]any{}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return nil, err
		}
		var item map[string]any
		if err = json.Unmarshal(raw, &item); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}
func object(ctx context.Context, q queryer, sql string, args ...any) (map[string]any, error) {
	var raw []byte
	if err := q.QueryRow(ctx, sql, args...).Scan(&raw); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &Error{404, "registro no encontrado"}
		}
		return nil, err
	}
	var m map[string]any
	err := json.Unmarshal(raw, &m)
	return m, err
}
func (s *Service) transaction(ctx context.Context, u *auth.User, run func(pgx.Tx) error) error {
	if u.ReadOnly() {
		return forbidden()
	}
	tx, err := s.repo.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	// Serialize grants/role revocation with operational mutations.
	var active bool
	if err = tx.QueryRow(ctx, `SELECT is_active FROM app_users WHERE id=$1 FOR SHARE`, u.ID).Scan(&active); err != nil || !active {
		return forbidden()
	}
	var readonly bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role_code='consulta')`, u.ID).Scan(&readonly); err != nil {
		return err
	}
	if readonly {
		return forbidden()
	}
	var currentRoles []string
	if err = tx.QueryRow(ctx, `SELECT ARRAY(SELECT role_code FROM user_roles WHERE user_id=$1 ORDER BY role_code)`, u.ID).Scan(&currentRoles); err != nil {
		return err
	}
	expected := len(u.Roles)
	if expected == 0 {
		expected = 1
	}
	if len(currentRoles) != expected {
		return forbidden()
	}
	for _, role := range currentRoles {
		if !u.HasRole(role) {
			return forbidden()
		}
	}
	if err = run(tx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Service) allowed(ctx context.Context, q queryer, u *auth.User, queue int64, permission string) bool {
	if permission != "read" && u.ReadOnly() {
		return false
	}
	if u.HasRole("admin") {
		return true
	}
	if u.ReadOnly() {
		return permission == "read"
	}
	columns := map[string]string{"read": "can_read", "handle": "can_handle", "transfer": "can_transfer", "approve": "can_approve", "admin": "can_admin"}
	col, ok := columns[permission]
	if !ok {
		return false
	}
	var yes bool
	err := q.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM queue_grants g WHERE queue_id=$1 AND %s AND (g.user_id=$2 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$2)))`, col), queue, u.ID).Scan(&yes)
	return err == nil && yes
}
func jsonBytes(v any) []byte { b, _ := json.Marshal(v); return b }
func (s *Service) audit(ctx context.Context, tx pgx.Tx, id, actor int64, action string, before, after any, reason string) error {
	_, err := tx.Exec(ctx, `INSERT INTO ticket_history(ticket_id,actor_id,action,before_value,after_value,reason) VALUES($1,$2,$3,$4,$5,$6)`, id, actor, action, jsonBytes(before), jsonBytes(after), reason)
	return err
}
func page(limit, offset int) (int, int) {
	if limit < 1 {
		limit = 25
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
func requireAdmin(u *auth.User) error {
	if !u.HasRole("admin") || u.ReadOnly() {
		return forbidden()
	}
	return nil
}
