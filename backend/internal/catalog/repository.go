package catalog

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
)

// Repository persists the normalized result of an Excel import.
type Repository struct {
	pool *pgxpool.Pool
}

type ServiceRow struct {
	ID                  int64    `json:"id"`
	Level1ID            int64    `json:"level1_id"`
	Code                string   `json:"code"`
	Name                string   `json:"name"`
	Level1Code          string   `json:"level1_code"`
	Level1Name          string   `json:"level1_name"`
	ActiveValue         *string  `json:"active_value"`
	ServiceClass        *string  `json:"service_class"`
	Criticality         *string  `json:"criticality"`
	ServiceType         *string  `json:"service_type"`
	Description         *string  `json:"description"`
	Metric              *string  `json:"metric"`
	Minimum             *float64 `json:"minimum"`
	Maximum             *float64 `json:"maximum"`
	ReviewRequired      bool     `json:"review_required"`
	IsActive            bool     `json:"is_active"`
	SectionID           *int64   `json:"section_id,omitempty"`
	SectionName         *string  `json:"section_name,omitempty"`
	ResponsibleUserID   *int64   `json:"responsible_user_id,omitempty"`
	ResponsibleUserName *string  `json:"responsible_user_name,omitempty"`
}

type AssignmentInput struct {
	ServiceID         int64  `json:"service_id"`
	SectionID         int64  `json:"section_id"`
	ResponsibleUserID *int64 `json:"responsible_user_id"`
}

type ServiceFilters struct {
	Query        string
	Level1ID     *int64
	IsActive     *bool
	ServiceClass string
	Criticality  string
	ServiceType  string
}

type Level1Option struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type Lookups struct {
	ServiceClasses []string `json:"service_classes"`
	Criticalities  []string `json:"criticalities"`
	ServiceTypes   []string `json:"service_types"`
}

type Level1Input struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type Level2Input struct {
	Level1ID     int64    `json:"level1_id"`
	Code         string   `json:"code"`
	Name         string   `json:"name"`
	ActiveValue  *string  `json:"active_value"`
	ServiceClass *string  `json:"service_class"`
	Criticality  *string  `json:"criticality"`
	ServiceType  *string  `json:"service_type"`
	Description  *string  `json:"description"`
	Metric       *string  `json:"metric"`
	Minimum      *float64 `json:"minimum"`
	Maximum      *float64 `json:"maximum"`
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) ListServices(ctx context.Context, filters ServiceFilters) ([]ServiceRow, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	rows, err := r.pool.Query(ctx, `
		SELECT s.id, s.service_level1_id, s.code, s.name, l.code, l.canonical_name, s.active_value,
		       c.name, cr.name, st.name, s.description, s.metric, s.minimum, s.maximum,
		       s.review_required, s.is_active, a.section_id, sec.name, a.responsible_user_id, u.full_name
		FROM services_level2 s
		JOIN services_level1 l ON l.id = s.service_level1_id
		LEFT JOIN service_classes c ON c.id = s.service_class_id
		LEFT JOIN criticalities cr ON cr.id = s.criticality_id
		LEFT JOIN service_types st ON st.id = s.service_type_id
		LEFT JOIN service_assignments a ON a.service_level2_id = s.id
		LEFT JOIN sections sec ON sec.id = a.section_id
		LEFT JOIN app_users u ON u.id = a.responsible_user_id
		WHERE ($1 = '' OR s.code ILIKE '%' || $1 || '%' OR s.name ILIKE '%' || $1 || '%'
		       OR l.code ILIKE '%' || $1 || '%' OR l.canonical_name ILIKE '%' || $1 || '%')
		  AND ($2::bigint IS NULL OR s.service_level1_id = $2)
		  AND ($3::boolean IS NULL OR s.is_active = $3)
		  AND ($4 = '' OR c.name = $4)
		  AND ($5 = '' OR cr.name = $5)
		  AND ($6 = '' OR st.name = $6)
		ORDER BY s.code`, filters.Query, filters.Level1ID, filters.IsActive, filters.ServiceClass, filters.Criticality, filters.ServiceType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	services := make([]ServiceRow, 0)
	for rows.Next() {
		var service ServiceRow
		if err := rows.Scan(&service.ID, &service.Level1ID, &service.Code, &service.Name, &service.Level1Code,
			&service.Level1Name, &service.ActiveValue, &service.ServiceClass, &service.Criticality,
			&service.ServiceType, &service.Description, &service.Metric, &service.Minimum,
			&service.Maximum, &service.ReviewRequired, &service.IsActive, &service.SectionID, &service.SectionName,
			&service.ResponsibleUserID, &service.ResponsibleUserName); err != nil {
			return nil, err
		}
		services = append(services, service)
	}
	return services, rows.Err()
}

func (r *Repository) ListLevel1(ctx context.Context) ([]Level1Option, error) {
	rows, err := r.pool.Query(ctx, `SELECT id, code, canonical_name FROM services_level1 WHERE is_active = TRUE ORDER BY code`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]Level1Option, 0)
	for rows.Next() {
		var option Level1Option
		if err := rows.Scan(&option.ID, &option.Code, &option.Name); err != nil {
			return nil, err
		}
		result = append(result, option)
	}
	return result, rows.Err()
}

func (r *Repository) ListLookups(ctx context.Context) (Lookups, error) {
	result := Lookups{}
	var err error
	if result.ServiceClasses, err = listNames(ctx, r.pool, "service_classes"); err != nil {
		return result, err
	}
	if result.Criticalities, err = listNames(ctx, r.pool, "criticalities"); err != nil {
		return result, err
	}
	if result.ServiceTypes, err = listNames(ctx, r.pool, "service_types"); err != nil {
		return result, err
	}
	return result, nil
}

func listNames(ctx context.Context, pool *pgxpool.Pool, table string) ([]string, error) {
	if table != "service_classes" && table != "criticalities" && table != "service_types" {
		return nil, fmt.Errorf("invalid lookup table")
	}
	rows, err := pool.Query(ctx, fmt.Sprintf(`SELECT name FROM %s ORDER BY name`, table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]string, 0)
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}
		result = append(result, value)
	}
	return result, rows.Err()
}

func (r *Repository) CreateLevel1(ctx context.Context, input Level1Input) (Level1Option, error) {
	input.Code, input.Name = strings.TrimSpace(input.Code), strings.TrimSpace(input.Name)
	if input.Code == "" || input.Name == "" {
		return Level1Option{}, fmt.Errorf("código y nombre son obligatorios")
	}
	var result Level1Option
	err := r.pool.QueryRow(ctx, `
		INSERT INTO services_level1 (code, canonical_name)
		VALUES ($1, $2)
		RETURNING id, code, canonical_name`, input.Code, input.Name).
		Scan(&result.ID, &result.Code, &result.Name)
	if err != nil {
		return Level1Option{}, fmt.Errorf("crear servicio nivel 1: %w", err)
	}
	return result, nil
}

func (r *Repository) UpdateLevel1(ctx context.Context, id int64, input Level1Input) (Level1Option, error) {
	input.Code, input.Name = strings.TrimSpace(input.Code), strings.TrimSpace(input.Name)
	if input.Code == "" || input.Name == "" {
		return Level1Option{}, fmt.Errorf("código y nombre son obligatorios")
	}
	var result Level1Option
	err := r.pool.QueryRow(ctx, `
		UPDATE services_level1 SET code = $1, canonical_name = $2
		WHERE id = $3
		RETURNING id, code, canonical_name`, input.Code, input.Name, id).
		Scan(&result.ID, &result.Code, &result.Name)
	if err != nil {
		return Level1Option{}, fmt.Errorf("actualizar servicio nivel 1: %w", err)
	}
	return result, nil
}

func (r *Repository) SetLevel1Active(ctx context.Context, id int64, active bool) error {
	if !active {
		var hasChildren bool
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM services_level2 WHERE service_level1_id = $1 AND is_active = TRUE)`, id).Scan(&hasChildren); err != nil {
			return err
		}
		if hasChildren {
			return fmt.Errorf("no se puede desactivar: existen servicios nivel 2 activos")
		}
	}
	result, err := r.pool.Exec(ctx, `UPDATE services_level1 SET is_active = $1 WHERE id = $2`, active, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("servicio nivel 1 no encontrado")
	}
	return nil
}

func validateLevel2Input(input Level2Input) error {
	input.Code, input.Name = strings.TrimSpace(input.Code), strings.TrimSpace(input.Name)
	if input.Level1ID <= 0 || input.Code == "" || input.Name == "" {
		return fmt.Errorf("nivel 1, código y nombre son obligatorios")
	}
	if input.Minimum != nil && input.Maximum != nil && *input.Minimum > *input.Maximum {
		return fmt.Errorf("el mínimo no puede ser mayor que el máximo")
	}
	return nil
}

func (r *Repository) CreateLevel2(ctx context.Context, input Level2Input) error {
	if err := validateLevel2Input(input); err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var parentActive bool
	if err := tx.QueryRow(ctx, `SELECT is_active FROM services_level1 WHERE id = $1`, input.Level1ID).Scan(&parentActive); err != nil || !parentActive {
		return fmt.Errorf("el servicio nivel 1 no existe o está inactivo")
	}
	classID, err := lookupID(ctx, tx, "service_classes", input.ServiceClass)
	if err != nil {
		return err
	}
	criticalityID, err := lookupID(ctx, tx, "criticalities", input.Criticality)
	if err != nil {
		return err
	}
	typeID, err := lookupID(ctx, tx, "service_types", input.ServiceType)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO services_level2
			(service_level1_id, code, name, active_value, service_class_id, criticality_id, service_type_id,
			 description, metric, minimum, maximum, review_required, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, FALSE, TRUE)`, input.Level1ID,
		input.Code, input.Name, input.ActiveValue, classID, criticalityID, typeID, input.Description,
		input.Metric, input.Minimum, input.Maximum); err != nil {
		return fmt.Errorf("crear servicio nivel 2: %w", err)
	}
	return tx.Commit(ctx)
}

func (r *Repository) UpdateLevel2(ctx context.Context, id int64, input Level2Input) error {
	if err := validateLevel2Input(input); err != nil {
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var parentActive bool
	if err := tx.QueryRow(ctx, `SELECT is_active FROM services_level1 WHERE id = $1`, input.Level1ID).Scan(&parentActive); err != nil || !parentActive {
		return fmt.Errorf("el servicio nivel 1 no existe o está inactivo")
	}
	classID, err := lookupID(ctx, tx, "service_classes", input.ServiceClass)
	if err != nil {
		return err
	}
	criticalityID, err := lookupID(ctx, tx, "criticalities", input.Criticality)
	if err != nil {
		return err
	}
	typeID, err := lookupID(ctx, tx, "service_types", input.ServiceType)
	if err != nil {
		return err
	}
	result, err := tx.Exec(ctx, `
		UPDATE services_level2 SET
			service_level1_id = $1, code = $2, name = $3, active_value = $4,
			service_class_id = $5, criticality_id = $6, service_type_id = $7,
			description = $8, metric = $9, minimum = $10, maximum = $11
		WHERE id = $12`, input.Level1ID, input.Code, input.Name, input.ActiveValue, classID,
		criticalityID, typeID, input.Description, input.Metric, input.Minimum, input.Maximum, id)
	if err != nil {
		return fmt.Errorf("actualizar servicio nivel 2: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("servicio nivel 2 no encontrado")
	}
	return tx.Commit(ctx)
}

func (r *Repository) SetLevel2Active(ctx context.Context, id int64, active bool) error {
	if !active {
		var assigned bool
		if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM service_assignments WHERE service_level2_id = $1)`, id).Scan(&assigned); err != nil {
			return err
		}
		if assigned {
			return fmt.Errorf("no se puede desactivar: el servicio tiene una asignación")
		}
	}
	result, err := r.pool.Exec(ctx, `UPDATE services_level2 SET is_active = $1 WHERE id = $2`, active, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("servicio nivel 2 no encontrado")
	}
	return nil
}

func (r *Repository) Assign(ctx context.Context, input AssignmentInput) error {
	var sectionActive bool
	if err := r.pool.QueryRow(ctx, `SELECT is_active FROM sections WHERE id = $1`, input.SectionID).Scan(&sectionActive); err != nil {
		return fmt.Errorf("sección no encontrada")
	}
	if !sectionActive {
		return fmt.Errorf("la sección está inactiva")
	}
	if input.ResponsibleUserID != nil {
		var valid bool
		if err := r.pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM app_users u
				JOIN positions p ON p.id = u.position_id
				WHERE u.id = $1 AND p.section_id = $2 AND u.is_active = TRUE
			)`, *input.ResponsibleUserID, input.SectionID).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return fmt.Errorf("el responsable no pertenece a la sección indicada o está inactivo")
		}
	}
	_, err := r.pool.Exec(ctx, `
		INSERT INTO service_assignments (service_level2_id, section_id, responsible_user_id)
		VALUES ($1, $2, $3)
		ON CONFLICT (service_level2_id) DO UPDATE SET
			section_id = EXCLUDED.section_id,
			responsible_user_id = EXCLUDED.responsible_user_id`, input.ServiceID, input.SectionID, input.ResponsibleUserID)
	return err
}

func (r *Repository) PersistImport(ctx context.Context, report *importer.Report) error {
	if report == nil {
		return fmt.Errorf("import report is required")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin catalog import: %w", err)
	}
	defer tx.Rollback(ctx)

	var runID int64
	err = tx.QueryRow(ctx, `
		INSERT INTO import_runs
			(source_file, status, created_count, updated_count, omitted_count, observed_count)
		VALUES ($1, 'running', $2, $3, $4, $5)
		RETURNING id`, report.SourceFile, report.Counts.Created, report.Counts.Updated,
		report.Counts.Omitted, report.Counts.Observed).Scan(&runID)
	if err != nil {
		return fmt.Errorf("create import run: %w", err)
	}

	level1IDs := make(map[string]int64, len(report.ServicesLevel1))
	for _, service := range report.ServicesLevel1 {
		id, err := upsertLevel1(ctx, tx, service)
		if err != nil {
			return fmt.Errorf("persist level 1 %s: %w", service.Code, err)
		}
		level1IDs[service.Code] = id
		if err := persistSourceNames(ctx, tx, id, service); err != nil {
			return fmt.Errorf("persist source names for %s: %w", service.Code, err)
		}
	}

	for _, service := range report.ServicesLevel2 {
		parentID, ok := level1IDs[service.ParentCode]
		if !ok {
			return fmt.Errorf("level 1 parent %s not found for %s", service.ParentCode, service.Code)
		}
		if err := upsertLevel2(ctx, tx, parentID, service); err != nil {
			return fmt.Errorf("persist level 2 %s: %w", service.Code, err)
		}
	}

	for _, observation := range report.Observations {
		if err := persistObservation(ctx, tx, runID, observation); err != nil {
			return fmt.Errorf("persist observation %s: %w", observation.Code, err)
		}
	}

	if _, err := tx.Exec(ctx, `
		UPDATE import_runs
		SET status = 'succeeded', finished_at = NOW()
		WHERE id = $1`, runID); err != nil {
		return fmt.Errorf("finish import run: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit catalog import: %w", err)
	}
	return nil
}

func upsertLevel1(ctx context.Context, tx pgx.Tx, service importer.ServiceLevel1) (int64, error) {
	var id int64
	err := tx.QueryRow(ctx, `
		INSERT INTO services_level1
			(code, canonical_name, review_required, source_sheet, source_row, source_range)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (code) DO UPDATE SET
			canonical_name = EXCLUDED.canonical_name,
			review_required = EXCLUDED.review_required,
			source_sheet = EXCLUDED.source_sheet,
			source_row = EXCLUDED.source_row,
			source_range = EXCLUDED.source_range
		RETURNING id`, service.Code, service.Name, len(service.SourceNames) > 1,
		service.Source.Sheet, service.Source.Row, service.Source.Range).Scan(&id)
	return id, err
}

func persistSourceNames(ctx context.Context, tx pgx.Tx, serviceID int64, service importer.ServiceLevel1) error {
	for index, source := range service.SourceNames {
		if _, err := tx.Exec(ctx, `
			INSERT INTO services_level1_source_names
				(service_level1_id, original_name, source_sheet, source_row, source_range, is_canonical)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (service_level1_id, original_name, source_sheet, source_row, source_range)
			DO UPDATE SET is_canonical = EXCLUDED.is_canonical`, serviceID, source.Value,
			source.Ref.Sheet, source.Ref.Row, source.Ref.Range, index == 0); err != nil {
			return err
		}
	}
	return nil
}

func upsertLevel2(ctx context.Context, tx pgx.Tx, parentID int64, service importer.ServiceLevel2) error {
	classID, err := lookupID(ctx, tx, "service_classes", service.ServiceClass)
	if err != nil {
		return err
	}
	criticalityID, err := lookupID(ctx, tx, "criticalities", service.Criticality)
	if err != nil {
		return err
	}
	typeID, err := lookupID(ctx, tx, "service_types", service.ServiceType)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO services_level2
			(service_level1_id, code, name, active_value, service_class_id, criticality_id,
			 service_type_id, description, metric, minimum, maximum, review_required,
			 source_sheet, source_row, source_range)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (code) DO UPDATE SET
			service_level1_id = EXCLUDED.service_level1_id,
			name = EXCLUDED.name,
			active_value = EXCLUDED.active_value,
			service_class_id = EXCLUDED.service_class_id,
			criticality_id = EXCLUDED.criticality_id,
			service_type_id = EXCLUDED.service_type_id,
			description = EXCLUDED.description,
			metric = EXCLUDED.metric,
			minimum = EXCLUDED.minimum,
			maximum = EXCLUDED.maximum,
			review_required = EXCLUDED.review_required,
			source_sheet = EXCLUDED.source_sheet,
			source_row = EXCLUDED.source_row,
			source_range = EXCLUDED.source_range`, parentID, service.Code, service.Name,
		service.Active, classID, criticalityID, typeID, service.Description, service.Metric,
		service.Minimum, service.Maximum, service.ReviewRequired, service.Source.Sheet,
		service.Source.Row, service.Source.Range)
	return err
}

func lookupID(ctx context.Context, tx pgx.Tx, table string, value *string) (*int64, error) {
	if value == nil {
		return nil, nil
	}
	if table != "service_classes" && table != "criticalities" && table != "service_types" {
		return nil, fmt.Errorf("invalid lookup table %s", table)
	}
	var id int64
	query := fmt.Sprintf(`
		INSERT INTO %s (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, table)
	if err := tx.QueryRow(ctx, query, *value).Scan(&id); err != nil {
		return nil, err
	}
	return &id, nil
}

func persistObservation(ctx context.Context, tx pgx.Tx, runID int64, observation importer.Observation) error {
	var sheet *string
	var row *int
	var sourceRange *string
	if len(observation.SourceRefs) > 0 {
		sheetValue := observation.SourceRefs[0].Sheet
		rowValue := observation.SourceRefs[0].Row
		rangeValue := observation.SourceRefs[0].Range
		sheet, row, sourceRange = &sheetValue, &rowValue, &rangeValue
	}
	details, err := json.Marshal(observation.Details)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO import_observations
			(import_run_id, code, severity, message, sheet_name, row_start, row_end, source_range, details)
		VALUES ($1, $2, $3, $4, $5, $6, $6, $7, $8)`, runID, observation.Code,
		observation.Severity, observation.Message, sheet, row, sourceRange, details)
	return err
}
