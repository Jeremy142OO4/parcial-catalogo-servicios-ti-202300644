package auth

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
)

func ValidRole(role string) bool {
	switch role {
	case "admin", "consulta", "solicitante", "agente", "aprobador":
		return true
	}
	return false
}
func (u *User) HasRole(role string) bool {
	if len(u.Roles) == 0 {
		return u.Role == role
	}
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}
func (u *User) ReadOnly() bool { return u.HasRole("consulta") }
func (s *Service) loadRoles(ctx context.Context, u *User) error {
	err := s.pool.QueryRow(ctx, `SELECT coalesce(array_agg(role_code ORDER BY role_code),'{}') FROM user_roles WHERE user_id=$1`, u.ID).Scan(&u.Roles)
	if u.ReadOnly() {
		u.Role = "consulta"
	} else if u.HasRole("admin") {
		u.Role = "admin"
	}
	return err
}
func (s *Service) SetRoles(ctx context.Context, id int64, roles []string) error {
	return s.SetRolesBy(ctx, 0, id, roles)
}

// Actor zero is reserved for internal setup jobs, never supplied by HTTP input.
func (s *Service) SetRolesBy(ctx context.Context, actor, id int64, roles []string) error {
	if len(roles) == 0 {
		return fmt.Errorf("al menos un rol requerido")
	}
	seen := map[string]bool{}
	for _, r := range roles {
		if !ValidRole(r) || seen[r] {
			return fmt.Errorf("roles inválidos o repetidos")
		}
		seen[r] = true
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if actor > 0 {
		if err = lockAdministrator(ctx, tx, actor); err != nil {
			return err
		}
	}
	var old string
	if err = tx.QueryRow(ctx, `SELECT role FROM app_users WHERE id=$1 FOR UPDATE`, id).Scan(&old); err != nil {
		return err
	}
	var pending bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tickets WHERE assignee_id=$1 AND state NOT IN ('cerrado','rechazado'))`, id).Scan(&pending); err != nil {
		return err
	}
	if pending {
		return fmt.Errorf("reasignar tickets antes de cambiar roles")
	}
	primary := roles[0]
	if seen["admin"] {
		primary = "admin"
	}
	if seen["consulta"] {
		primary = "consulta"
	}
	if _, err = tx.Exec(ctx, `UPDATE app_users SET role=$2 WHERE id=$1`, id, primary); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM user_roles WHERE user_id=$1`, id); err != nil {
		return err
	}
	batch := &pgx.Batch{}
	for _, r := range roles {
		batch.Queue(`INSERT INTO user_roles VALUES($1,$2)`, id, r)
	}
	if err = tx.SendBatch(ctx, batch).Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
