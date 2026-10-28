package desk

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"strings"
)

type QueueInput struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Process     string `json:"process_code"`
	Level       int    `json:"support_level"`
	First       *int   `json:"first_response_minutes"`
	Resolution  *int   `json:"resolution_minutes"`
	Active      bool   `json:"is_active"`
}

func (s *Service) Queues(ctx context.Context, u *auth.User) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(q)||jsonb_build_object('can_admin',($3 AND ($1 OR EXISTS(SELECT 1 FROM queue_grants g WHERE g.queue_id=q.id AND g.can_admin AND(g.user_id=$2 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$2)))))) FROM queues q WHERE $1 OR $4 OR EXISTS(SELECT 1 FROM queue_grants g WHERE g.queue_id=q.id AND (g.can_read OR ($3 AND g.can_admin)) AND(g.user_id=$2 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$2))) ORDER BY q.id`, u.HasRole("admin"), u.ID, !u.ReadOnly(), u.ReadOnly())
}
func (s *Service) Processes(ctx context.Context) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(p)||jsonb_build_object('states',(SELECT jsonb_agg(code ORDER BY code) FROM workflow_states WHERE process_code=p.code)) FROM processes p ORDER BY code`)
}
func (s *Service) SaveQueue(ctx context.Context, u *auth.User, id int64, in QueueInput) (map[string]any, error) {
	if u.ReadOnly() || (id == 0 && !u.HasRole("admin")) {
		return nil, forbidden()
	}
	if strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.Name) == "" {
		return nil, bad("código y nombre requeridos")
	}
	var out map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error {
		if id > 0 {
			if e := s.lockQueueAdministration(ctx, tx, u, id); e != nil {
				return e
			}
			var used bool
			if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tickets WHERE queue_id=$1)`, id).Scan(&used); e != nil {
				return e
			}
			if used {
				var old string
				if e := tx.QueryRow(ctx, `SELECT process_code FROM queues WHERE id=$1 FOR UPDATE`, id).Scan(&old); e != nil {
					return e
				}
				if old != in.Process {
					return conflict("no cambiar proceso de cola con historial")
				}
			}
		}
		var e error
		if id == 0 {
			out, e = object(ctx, tx, `INSERT INTO queues(code,name,description,process_code,support_level,first_response_minutes,resolution_minutes,is_active) VALUES($1,$2,$3,$4,$5,$6,$7,$8) RETURNING to_jsonb(queues)`, in.Code, in.Name, in.Description, in.Process, in.Level, in.First, in.Resolution, in.Active)
		} else {
			out, e = object(ctx, tx, `UPDATE queues SET code=$2,name=$3,description=$4,process_code=$5,support_level=$6,first_response_minutes=$7,resolution_minutes=$8,is_active=$9 WHERE id=$1 RETURNING to_jsonb(queues)`, id, in.Code, in.Name, in.Description, in.Process, in.Level, in.First, in.Resolution, in.Active)
		}
		return e
	})
	return out, err
}

type GrantInput struct {
	UserID   *int64  `json:"user_id"`
	Role     *string `json:"role_code"`
	Read     bool    `json:"can_read"`
	Handle   bool    `json:"can_handle"`
	Transfer bool    `json:"can_transfer"`
	Approve  bool    `json:"can_approve"`
	Admin    bool    `json:"can_admin"`
}

func (s *Service) Grant(ctx context.Context, u *auth.User, queue int64, in GrantInput) error {
	if u.ReadOnly() {
		return forbidden()
	}
	if (in.UserID == nil) == (in.Role == nil) {
		return bad("elegir un usuario o un rol")
	}
	return s.transaction(ctx, u, func(tx pgx.Tx) error {
		if e := s.lockQueueAdministration(ctx, tx, u, queue); e != nil {
			return e
		}
		var pending bool
		if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tickets t WHERE t.queue_id=$1 AND t.assignee_id IS NOT NULL AND t.state NOT IN ('cerrado','rechazado') AND (t.assignee_id=$2 OR EXISTS(SELECT 1 FROM user_roles ur WHERE ur.user_id=t.assignee_id AND ur.role_code=$3)))`, queue, in.UserID, in.Role).Scan(&pending); e != nil {
			return e
		}
		if pending {
			return conflict("reasignar tickets antes de editar permisos")
		}
		_, e := tx.Exec(ctx, `DELETE FROM queue_grants WHERE queue_id=$1 AND (user_id=$2 OR role_code=$3)`, queue, in.UserID, in.Role)
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `INSERT INTO queue_grants(queue_id,user_id,role_code,can_read,can_handle,can_transfer,can_approve,can_admin) VALUES($1,$2,$3,$4,$5,$6,$7,$8)`, queue, in.UserID, in.Role, in.Read, in.Handle, in.Transfer, in.Approve, in.Admin)
		return e
	})
}

// Check after the lock so a concurrent revocation cannot authorize stale grants.
func (s *Service) lockQueueAdministration(ctx context.Context, tx pgx.Tx, u *auth.User, id int64) error {
	var locked int64
	if e := tx.QueryRow(ctx, `SELECT id FROM queues WHERE id=$1 FOR UPDATE`, id).Scan(&locked); e != nil {
		return &Error{404, "cola no encontrada"}
	}
	if !s.allowed(ctx, tx, u, id, "admin") {
		return forbidden()
	}
	return nil
}
