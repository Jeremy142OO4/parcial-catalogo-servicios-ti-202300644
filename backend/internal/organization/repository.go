package organization

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Unit struct {
	ID         int64  `json:"id"`
	Type       string `json:"type"`
	Code       string `json:"code"`
	Name       string `json:"name"`
	ParentID   *int64 `json:"parent_id,omitempty"`
	ParentType string `json:"parent_type,omitempty"`
	IsActive   bool   `json:"is_active"`
}

type Repository struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

func (r *Repository) ListUnits(ctx context.Context) ([]Unit, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, type, code, name, parent_id, parent_type, is_active
		FROM (
			SELECT c.id, 'company'::text AS type, c.code, c.name, NULL::bigint AS parent_id,
			       ''::text AS parent_type, c.is_active FROM companies c
			UNION ALL
			SELECT a.id, 'area', a.code, a.name, a.company_id, 'company', a.is_active FROM areas a
			UNION ALL
			SELECT d.id, 'department', d.code, d.name, d.area_id, 'area', d.is_active FROM departments d
			UNION ALL
			SELECT s.id, 'section', s.code, s.name, s.department_id, 'department', s.is_active FROM sections s
			UNION ALL
			SELECT p.id, 'position', p.code, p.name, p.section_id, 'section', p.is_active FROM positions p
		) units
		ORDER BY type, code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	units := make([]Unit, 0)
	for rows.Next() {
		var unit Unit
		if err := rows.Scan(&unit.ID, &unit.Type, &unit.Code, &unit.Name, &unit.ParentID, &unit.ParentType, &unit.IsActive); err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, rows.Err()
}

func (r *Repository) CreateUnit(ctx context.Context, unitType string, parentID *int64, code, name string) (Unit, error) {
	unitType = strings.ToLower(strings.TrimSpace(unitType))
	code, name = strings.TrimSpace(code), strings.TrimSpace(name)
	if code == "" || name == "" {
		return Unit{}, fmt.Errorf("código y nombre son obligatorios")
	}
	if unitType != "company" && parentID == nil {
		return Unit{}, fmt.Errorf("la unidad %s requiere un padre", unitType)
	}
	if parentID != nil && !parentIsActive(ctx, r.pool, unitType, *parentID) {
		return Unit{}, fmt.Errorf("el padre no existe o está inactivo")
	}
	var unit Unit
	var err error
	switch unitType {
	case "company":
		err = r.pool.QueryRow(ctx, `INSERT INTO companies (code, name) VALUES ($1, $2) RETURNING id, code, name, is_active`, code, name).
			Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	case "area":
		err = r.pool.QueryRow(ctx, `INSERT INTO areas (company_id, code, name) VALUES ($1, $2, $3) RETURNING id, code, name, is_active`, *parentID, code, name).
			Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	case "department":
		err = r.pool.QueryRow(ctx, `INSERT INTO departments (area_id, code, name) VALUES ($1, $2, $3) RETURNING id, code, name, is_active`, *parentID, code, name).
			Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	case "section":
		err = r.pool.QueryRow(ctx, `INSERT INTO sections (department_id, code, name) VALUES ($1, $2, $3) RETURNING id, code, name, is_active`, *parentID, code, name).
			Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	case "position":
		err = r.pool.QueryRow(ctx, `INSERT INTO positions (section_id, code, name) VALUES ($1, $2, $3) RETURNING id, code, name, is_active`, *parentID, code, name).
			Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	default:
		return Unit{}, fmt.Errorf("tipo de unidad inválido")
	}
	if err != nil {
		return Unit{}, fmt.Errorf("crear %s: %w", unitType, err)
	}
	unit.Type = unitType
	unit.ParentID = parentID
	unit.ParentType = parentType(unitType)
	return unit, nil
}

func parentIsActive(ctx context.Context, pool *pgxpool.Pool, childType string, parentID int64) bool {
	parentTable := map[string]string{"area": "companies", "department": "areas", "section": "departments", "position": "sections"}[childType]
	if parentTable == "" {
		return false
	}
	var active bool
	_ = pool.QueryRow(ctx, fmt.Sprintf(`SELECT is_active FROM %s WHERE id = $1`, parentTable), parentID).Scan(&active)
	return active
}

func parentType(unitType string) string {
	return map[string]string{"area": "company", "department": "area", "section": "department", "position": "section"}[unitType]
}

func (r *Repository) SetActive(ctx context.Context, unitType string, id int64, active bool) error {
	table := map[string]string{"company": "companies", "area": "areas", "department": "departments", "section": "sections", "position": "positions"}[unitType]
	if table == "" {
		return fmt.Errorf("tipo de unidad inválido")
	}
	if !active {
		children := map[string]string{"company": "areas", "area": "departments", "department": "sections", "section": "positions", "position": "app_users"}[unitType]
		column := map[string]string{"company": "company_id", "area": "area_id", "department": "department_id", "section": "section_id", "position": "position_id"}[unitType]
		var exists bool
		if err := r.pool.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE %s = $1 AND is_active = TRUE)`, children, column), id).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("no se puede desactivar: existen dependencias activas")
		}
	}
	commandTag, err := r.pool.Exec(ctx, fmt.Sprintf(`UPDATE %s SET is_active = $1 WHERE id = $2`, table), active, id)
	if err != nil {
		return err
	}
	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("registro no encontrado")
	}
	return nil
}

func (r *Repository) UpdateUnit(ctx context.Context, unitType string, id int64, code, name string) (Unit, error) {
	code, name = strings.TrimSpace(code), strings.TrimSpace(name)
	if code == "" || name == "" {
		return Unit{}, fmt.Errorf("código y nombre son obligatorios")
	}
	table := map[string]string{"company": "companies", "area": "areas", "department": "departments", "section": "sections", "position": "positions"}[unitType]
	if table == "" {
		return Unit{}, fmt.Errorf("tipo de unidad inválido")
	}
	var unit Unit
	err := r.pool.QueryRow(ctx, fmt.Sprintf(`
		UPDATE %s SET code = $1, name = $2
		WHERE id = $3
		RETURNING id, code, name, is_active`, table), code, name, id).
		Scan(&unit.ID, &unit.Code, &unit.Name, &unit.IsActive)
	if err != nil {
		return Unit{}, fmt.Errorf("actualizar unidad: %w", err)
	}
	unit.Type = unitType
	return unit, nil
}
