package catalog

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
)

// Repository persists the normalized result of an Excel import.
type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
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
