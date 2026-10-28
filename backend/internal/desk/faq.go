package desk

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"regexp"
	"strings"
)

type FAQInput struct {
	Slug      string  `json:"slug"`
	Title     string  `json:"title"`
	Body      string  `json:"body"`
	Process   *string `json:"process_code"`
	Service   *int64  `json:"service_level2_id"`
	Published bool    `json:"is_published"`
	Active    bool    `json:"is_active"`
	Version   int64   `json:"version"`
}

func (s *Service) FAQs(ctx context.Context, u *auth.User, search, process string, manage bool, limit, offset int) ([]map[string]any, error) {
	if manage {
		if e := requireAdmin(u); e != nil {
			return nil, e
		}
	}
	limit, offset = page(limit, offset)
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(f) FROM faq_articles f WHERE ($1 OR is_active AND is_published) AND ($2='' OR title ILIKE '%'||$2||'%' OR body ILIKE '%'||$2||'%') AND ($3='' OR process_code=$3) ORDER BY title,id LIMIT $4 OFFSET $5`, manage, search, process, limit, offset)
}
func (s *Service) SaveFAQ(ctx context.Context, u *auth.User, id int64, in FAQInput) (map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Body = strings.TrimSpace(in.Body)
	if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,99}$`).MatchString(in.Slug) || len(in.Title) == 0 || len(in.Title) > 200 || len(in.Body) == 0 || len(in.Body) > 20000 {
		return nil, bad("slug, título o contenido inválidos")
	}
	var out map[string]any
	e := s.transaction(ctx, u, func(tx pgx.Tx) error {
		var before map[string]any
		if id > 0 {
			var e error
			before, e = object(ctx, tx, `SELECT to_jsonb(f) FROM faq_articles f WHERE id=$1 FOR UPDATE`, id)
			if e != nil {
				return e
			}
			if int64(before["version"].(float64)) != in.Version {
				return conflict("versión de artículo desactualizada")
			}
		}
		if in.Service != nil {
			var active bool
			if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM services_level2 WHERE id=$1 AND is_active)`, *in.Service).Scan(&active); e != nil {
				return e
			}
			if !active {
				return bad("servicio inexistente o inactivo")
			}
		}
		var e error
		if id == 0 {
			out, e = object(ctx, tx, `INSERT INTO faq_articles(slug,title,body,process_code,service_level2_id,is_published,is_active,created_by,updated_by) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$8) RETURNING to_jsonb(faq_articles)`, in.Slug, in.Title, in.Body, in.Process, in.Service, in.Published, in.Active, u.ID)
		} else {
			out, e = object(ctx, tx, `UPDATE faq_articles SET slug=$2,title=$3,body=$4,process_code=$5,service_level2_id=$6,is_published=$7,is_active=$8,updated_by=$9,updated_at=now(),version=version+1 WHERE id=$1 RETURNING to_jsonb(faq_articles)`, id, in.Slug, in.Title, in.Body, in.Process, in.Service, in.Published, in.Active, u.ID)
		}
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `INSERT INTO faq_history(article_id,actor_id,before_value,after_value) VALUES($1,$2,$3,$4)`, int64(out["id"].(float64)), u.ID, jsonBytes(before), jsonBytes(out))
		return e
	})
	return out, e
}
