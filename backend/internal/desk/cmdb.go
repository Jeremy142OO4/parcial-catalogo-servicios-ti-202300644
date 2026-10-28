package desk

import (
	"context"
	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"net"
	"strings"
)

type ClassInput struct {
	Code     string `json:"code"`
	Name     string `json:"name"`
	ParentID *int64 `json:"parent_id"`
	IsDevice bool   `json:"is_device_class"`
	IsActive bool   `json:"is_active"`
}

func (s *Service) Classes(ctx context.Context) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(c) FROM ci_classes c ORDER BY id`)
}
func (s *Service) SaveClass(ctx context.Context, u *auth.User, id int64, in ClassInput) (map[string]any, error) {
	if err := requireAdmin(u); err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.Name) == "" {
		return nil, bad("código y nombre requeridos")
	}
	var out map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(202300645)`); e != nil {
			return e
		}
		if id > 0 {
			var used bool
			if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM configuration_items WHERE class_id=$1 AND is_active)`, id).Scan(&used); e != nil {
				return e
			}
			if used {
				var oldDevice bool
				if e := tx.QueryRow(ctx, `SELECT is_device_class FROM ci_classes WHERE id=$1`, id).Scan(&oldDevice); e != nil {
					return e
				}
				if !in.IsActive || oldDevice != in.IsDevice {
					return conflict("clase con CI activos")
				}
			}
		}
		var e error
		if id == 0 {
			out, e = object(ctx, tx, `INSERT INTO ci_classes(code,name,parent_id,is_device_class,is_active) VALUES($1,$2,$3,$4,$5) RETURNING to_jsonb(ci_classes)`, in.Code, in.Name, in.ParentID, in.IsDevice, in.IsActive)
		} else {
			out, e = object(ctx, tx, `UPDATE ci_classes SET code=$2,name=$3,parent_id=$4,is_device_class=$5,is_active=$6 WHERE id=$1 RETURNING to_jsonb(ci_classes)`, id, in.Code, in.Name, in.ParentID, in.IsDevice, in.IsActive)
		}
		return e
	})
	return out, err
}

type CIInput struct {
	Code          string         `json:"code"`
	Name          string         `json:"name"`
	ClassID       int64          `json:"class_id"`
	LocationID    int64          `json:"location_id"`
	OwnerID       int64          `json:"owner_user_id"`
	Description   *string        `json:"description"`
	Serial        *string        `json:"serial_number"`
	Version       *string        `json:"version"`
	AcquiredOn    *string        `json:"acquired_on"`
	Status        string         `json:"status"`
	Security      *string        `json:"security_level"`
	Compliance    *string        `json:"compliance"`
	Configuration *string        `json:"configuration_status"`
	License       *string        `json:"license_number"`
	Expires       *string        `json:"license_expires_on"`
	Technical     map[string]any `json:"technical_attributes"`
	IsDevice      bool           `json:"is_device"`
	IsActive      bool           `json:"is_active"`
}

func validateTechnical(in CIInput) error {
	if strings.TrimSpace(in.Code) == "" || strings.TrimSpace(in.Name) == "" {
		return bad("código y nombre requeridos")
	}
	if in.Status != "activo" && in.Status != "mantenimiento" && in.Status != "retirado" && in.Status != "desconocido" {
		return bad("estado inválido")
	}
	for k, v := range in.Technical {
		switch k {
		case "ip":
			x, ok := v.(string)
			if !ok || net.ParseIP(x) == nil {
				return bad("IP inválida")
			}
		case "ram_gb", "storage_gb", "ports", "power_watts":
			x, ok := v.(float64)
			if !ok || x <= 0 {
				return bad("atributo técnico numérico inválido")
			}
		default:
			return bad("atributo técnico no permitido")
		}
	}
	return nil
}
func (s *Service) SaveCI(ctx context.Context, u *auth.User, id int64, in CIInput) (map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	if e := validateTechnical(in); e != nil {
		return nil, e
	}
	if in.Technical == nil {
		in.Technical = map[string]any{}
	}
	var out map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error {
		var device bool
		var classCode string
		if e := tx.QueryRow(ctx, `SELECT is_device_class,coalesce(technical_profile,'') FROM ci_classes WHERE id=$1 AND is_active FOR SHARE`, in.ClassID).Scan(&device, &classCode); e != nil {
			return bad("clase inexistente o inactiva")
		}
		if in.IsDevice != device {
			return bad("indicador de dispositivo incompatible con clase")
		}
		allowed := map[string][]string{"servidor": {"ip", "ram_gb", "storage_gb"}, "usuario": {"ip", "ram_gb", "storage_gb"}, "red": {"ip", "ports"}, "impresion": {"ip"}, "almacenamiento": {"ip", "storage_gb"}, "energia": {"power_watts"}}
		for k := range in.Technical {
			ok := false
			for _, key := range allowed[classCode] {
				if k == key {
					ok = true
				}
			}
			if !ok {
				return bad("atributo incompatible con clase")
			}
		}
		var ref bool
		if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM locations WHERE id=$1 AND is_active) AND EXISTS(SELECT 1 FROM app_users WHERE id=$2 AND is_active)`, in.LocationID, in.OwnerID).Scan(&ref); e != nil {
			return e
		}
		if !ref {
			return bad("ubicación o responsable inexistente/inactivo")
		}
		var before map[string]any
		if id > 0 {
			var e error
			before, e = object(ctx, tx, `SELECT to_jsonb(c) FROM configuration_items c WHERE id=$1 FOR UPDATE`, id)
			if e != nil {
				return e
			}
		}
		args := []any{in.Code, in.Name, in.ClassID, in.LocationID, in.OwnerID, in.Description, in.Serial, in.Version, in.AcquiredOn, in.Status, in.Security, in.Compliance, in.Configuration, in.License, in.Expires, jsonBytes(in.Technical), in.IsDevice, in.IsActive}
		var e error
		if id == 0 {
			out, e = object(ctx, tx, `INSERT INTO configuration_items(code,name,class_id,location_id,owner_user_id,description,serial_number,version,acquired_on,status,security_level,compliance,configuration_status,license_number,license_expires_on,technical_attributes,is_device,is_active) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18) RETURNING to_jsonb(configuration_items)`, args...)
		} else {
			args = append(args, id)
			out, e = object(ctx, tx, `UPDATE configuration_items SET code=$1,name=$2,class_id=$3,location_id=$4,owner_user_id=$5,description=$6,serial_number=$7,version=$8,acquired_on=$9,status=$10,security_level=$11,compliance=$12,configuration_status=$13,license_number=$14,license_expires_on=$15,technical_attributes=$16,is_device=$17,is_active=$18,updated_at=now() WHERE id=$19 RETURNING to_jsonb(configuration_items)`, args...)
		}
		if e != nil {
			return e
		}
		_, e = tx.Exec(ctx, `INSERT INTO ci_changes(ci_id,actor_id,description,before_value,after_value) VALUES($1,$2,'guardar',$3,$4)`, int64(out["id"].(float64)), u.ID, jsonBytes(before), jsonBytes(out))
		return e
	})
	return out, err
}
func (s *Service) ListCI(ctx context.Context, search string, classID int64, status string, limit, offset int) ([]map[string]any, error) {
	limit, offset = page(limit, offset)
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(c) FROM configuration_items c WHERE ($1='' OR code ILIKE '%'||$1||'%' OR name ILIKE '%'||$1||'%') AND ($2::bigint=0 OR class_id=$2) AND ($3='' OR status=$3) ORDER BY id LIMIT $4 OFFSET $5`, search, classID, status, limit, offset)
}
func (s *Service) CIDetail(ctx context.Context, id int64) (map[string]any, error) {
	return object(ctx, s.repo.pool, `SELECT to_jsonb(c)||jsonb_build_object('classification',(WITH RECURSIVE p AS(SELECT id,name,parent_id FROM ci_classes WHERE id=c.class_id UNION SELECT x.id,x.name,x.parent_id FROM ci_classes x JOIN p ON x.id=p.parent_id)SELECT coalesce(jsonb_agg(to_jsonb(p)),'[]') FROM p),'relations',(SELECT coalesce(jsonb_agg(to_jsonb(r)),'[]') FROM ci_relations r WHERE source_ci_id=c.id OR target_ci_id=c.id),'services',(SELECT coalesce(jsonb_agg(service_level2_id),'[]') FROM ci_services WHERE ci_id=c.id),'history',(SELECT coalesce(jsonb_agg(to_jsonb(h)),'[]') FROM ci_changes h WHERE ci_id=c.id)) FROM configuration_items c WHERE id=$1`, id)
}
func (s *Service) Relation(ctx context.Context, u *auth.User, id, from, to int64, kind string, remove bool) error {
	if e := requireAdmin(u); e != nil {
		return e
	}
	return s.transaction(ctx, u, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(202300646)`); e != nil {
			return e
		}
		if remove {
			tag, e := tx.Exec(ctx, `DELETE FROM ci_relations WHERE id=$1`, id)
			if e == nil && tag.RowsAffected() == 0 {
				return &Error{404, "relación no encontrada"}
			}
			return e
		}
		if from == to {
			return bad("autorrelación no permitida")
		}
		var cycles bool
		if e := tx.QueryRow(ctx, `SELECT allows_cycles FROM ci_relation_types WHERE code=$1`, kind).Scan(&cycles); e != nil {
			return bad("tipo de relación inválido")
		}
		var active bool
		if e := tx.QueryRow(ctx, `SELECT count(*)=2 FROM configuration_items WHERE id IN($1,$2) AND is_active`, from, to).Scan(&active); e != nil {
			return e
		}
		if !active {
			return bad("CI inexistente o inactivo")
		}
		if !cycles {
			var loop bool
			if e := tx.QueryRow(ctx, `WITH RECURSIVE reach(id) AS(SELECT $2::bigint UNION SELECT r.target_ci_id FROM ci_relations r JOIN reach p ON r.source_ci_id=p.id JOIN ci_relation_types t ON t.code=r.kind WHERE NOT t.allows_cycles) SELECT EXISTS(SELECT 1 FROM reach WHERE id=$1)`, from, to).Scan(&loop); e != nil {
				return e
			}
			if loop {
				return conflict("ciclo de dependencias")
			}
		}
		if kind == "conectado_a" && from > to {
			from, to = to, from
		}
		if id > 0 {
			tag, e := tx.Exec(ctx, `UPDATE ci_relations SET source_ci_id=$2,target_ci_id=$3,kind=$4 WHERE id=$1`, id, from, to, kind)
			if e == nil && tag.RowsAffected() == 0 {
				return &Error{404, "relación no encontrada"}
			}
			return e
		}
		_, e := tx.Exec(ctx, `INSERT INTO ci_relations(source_ci_id,target_ci_id,kind) VALUES($1,$2,$3)`, from, to, kind)
		return e
	})
}
func (s *Service) Impact(ctx context.Context, id int64) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `WITH RECURSIVE reach(id,path) AS(SELECT $1::bigint,ARRAY[$1::bigint] UNION ALL SELECT r.source_ci_id,p.path||r.source_ci_id FROM reach p JOIN ci_relations r ON r.target_ci_id=p.id JOIN ci_relation_types t ON t.code=r.kind WHERE t.propagates_impact AND NOT r.source_ci_id=ANY(p.path) AND cardinality(p.path)<32) SELECT jsonb_build_object('ci_id',r.id,'path',r.path,'service_id',s.service_level2_id,'potential_only',true,'depth_limit_reached',cardinality(r.path)=32) FROM reach r LEFT JOIN ci_services s ON s.ci_id=r.id LIMIT 1000`, id)
}
func (s *Service) LinkService(ctx context.Context, u *auth.User, ci, service int64) error {
	if e := requireAdmin(u); e != nil {
		return e
	}
	return s.transaction(ctx, u, func(tx pgx.Tx) error {
		var yes bool
		if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM configuration_items WHERE id=$1 AND is_active) AND EXISTS(SELECT 1 FROM services_level2 WHERE id=$2 AND is_active)`, ci, service).Scan(&yes); e != nil {
			return e
		}
		if !yes {
			return bad("referencia inexistente o inactiva")
		}
		_, e := tx.Exec(ctx, `INSERT INTO ci_services VALUES($1,$2) ON CONFLICT DO NOTHING`, ci, service)
		return e
	})
}
func (s *Service) Locations(ctx context.Context) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(l) FROM locations l ORDER BY id`)
}
func (s *Service) SaveLocation(ctx context.Context, u *auth.User, id int64, code, name string, active bool) (map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	if strings.TrimSpace(code) == "" || strings.TrimSpace(name) == "" {
		return nil, bad("código y nombre requeridos")
	}
	var out map[string]any
	e := s.transaction(ctx, u, func(tx pgx.Tx) error {
		var e error
		if id == 0 {
			out, e = object(ctx, tx, `INSERT INTO locations(code,name,is_active) VALUES($1,$2,$3) RETURNING to_jsonb(locations)`, code, name, active)
		} else {
			out, e = object(ctx, tx, `UPDATE locations SET code=$2,name=$3,is_active=$4 WHERE id=$1 RETURNING to_jsonb(locations)`, id, code, name, active)
		}
		return e
	})
	return out, e
}
