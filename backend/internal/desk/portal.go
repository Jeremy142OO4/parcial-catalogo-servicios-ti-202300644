package desk

import (
	"context"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
)

func (s *Service) IntakeQueues(ctx context.Context) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT jsonb_build_object('id',id,'code',code,'name',name,'process_code',process_code) FROM queues WHERE is_active ORDER BY name`)
}
func (s *Service) OwnTickets(ctx context.Context, u *auth.User, q, state string, limit, offset int) ([]map[string]any, error) {
	limit, offset = page(limit, offset)
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(t) FROM tickets t WHERE requester_id=$1 AND ($2='' OR number ILIKE '%'||$2||'%' OR title ILIKE '%'||$2||'%') AND ($3='' OR state=$3) ORDER BY created_at DESC,id DESC LIMIT $4 OFFSET $5`, u.ID, q, state, limit, offset)
}
func (s *Service) QueueGrants(ctx context.Context, u *auth.User, id int64) ([]map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(g) FROM queue_grants g WHERE queue_id=$1 ORDER BY id`, id)
}

func (s *Service) CMDBSources(ctx context.Context, u *auth.User) ([]map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(s) FROM source_records s ORDER BY source_file,sheet,source_row,id`)
}
func (s *Service) CITickets(ctx context.Context, u *auth.User, id int64) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT jsonb_build_object('id',t.id,'number',t.number,'title',t.title,'state',t.state) FROM ticket_ci c JOIN tickets t ON t.id=c.ticket_id WHERE c.ci_id=$1 AND ($2 OR t.requester_id=$3 OR EXISTS(SELECT 1 FROM queue_grants g WHERE g.queue_id=t.queue_id AND g.can_read AND(g.user_id=$3 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$3)))) ORDER BY t.id DESC`, id, u.HasRole("admin") || u.ReadOnly(), u.ID)
}
