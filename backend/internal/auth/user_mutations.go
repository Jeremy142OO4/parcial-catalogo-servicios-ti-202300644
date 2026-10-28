package auth

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type userQuery interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

// HTTP actors are supplied by authentication, never by the submitted JSON.
func (s *Service) administratorMutation(ctx context.Context, actor int64, run func(pgx.Tx) error) error {
	if actor <= 0 {
		return fmt.Errorf("operación no autorizada")
	}
	tx, e := s.pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	if e = lockAdministrator(ctx, tx, actor); e != nil {
		return e
	}
	if e = run(tx); e != nil {
		return e
	}
	return tx.Commit(ctx)
}

func lockAdministrator(ctx context.Context, tx pgx.Tx, actor int64) error {
	var active, authorized bool
	if e := tx.QueryRow(ctx, `SELECT is_active FROM app_users WHERE id=$1 FOR UPDATE`, actor).Scan(&active); e != nil || !active {
		return fmt.Errorf("operación no autorizada")
	}
	// A separate statement obtains a fresh snapshot after a lock wait.
	if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role_code='admin') AND NOT EXISTS(SELECT 1 FROM user_roles WHERE user_id=$1 AND role_code='consulta')`, actor).Scan(&authorized); e != nil || !authorized {
		return fmt.Errorf("operación no autorizada")
	}
	_, e := tx.Exec(ctx, `SELECT set_config('app.actor_id',$1,true)`, fmt.Sprint(actor))
	return e
}

func (s *Service) CreateUserBy(ctx context.Context, actor, position int64, name, username, email, password, role string) (User, error) {
	var user User
	e := s.administratorMutation(ctx, actor, func(tx pgx.Tx) error {
		var e error
		user, e = createUser(ctx, tx, position, name, username, email, password, role)
		return e
	})
	return user, e
}
func (s *Service) UpdateUserBy(ctx context.Context, actor, id int64, position *int64, name, email, role *string) (User, error) {
	var user User
	e := s.administratorMutation(ctx, actor, func(tx pgx.Tx) error {
		var e error
		user, e = updateUser(ctx, tx, id, position, name, email, role)
		return e
	})
	return user, e
}
func (s *Service) SetUserActiveBy(ctx context.Context, actor, id int64, active bool) error {
	return s.administratorMutation(ctx, actor, func(tx pgx.Tx) error { return setUserActive(ctx, tx, id, active) })
}
