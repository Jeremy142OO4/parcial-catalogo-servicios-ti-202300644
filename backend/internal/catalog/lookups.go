package catalog

import (
	"context"
	"fmt"
	"strings"
)

func (r *Repository) MaintainLookup(ctx context.Context, kind, original, name, method string) error {
	tables := map[string]string{"classes": "service_classes", "criticalities": "criticalities", "types": "service_types"}
	columns := map[string]string{"classes": "service_class_id", "criticalities": "criticality_id", "types": "service_type_id"}
	table, ok := tables[kind]
	if !ok {
		return fmt.Errorf("catálogo inválido")
	}
	name, original = strings.TrimSpace(name), strings.TrimSpace(original)
	originalLabels := map[string][]string{
		"classes":       {"A DEMANDA", "RECURRENTE"},
		"criticalities": {"Very Low", "Low", "Normal", "High", "Very High"},
		"types":         {"Back End", "Demostration", "End User Service", "Front End", "IT Management", "IT Operational", "Other", "Project", "Reporting", "Training", "Underpinning Contract"},
	}
	if method != "POST" {
		for _, label := range originalLabels[kind] {
			if original == label && (method == "DELETE" || name != original) {
				return fmt.Errorf("la etiqueta original del Excel está protegida")
			}
		}
	}
	if method != "DELETE" && name == "" {
		return fmt.Errorf("nombre obligatorio")
	}
	if method == "POST" {
		_, err := r.pool.Exec(ctx, fmt.Sprintf("INSERT INTO %s (name) VALUES ($1)", table), name)
		return err
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id int64
	if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT id FROM %s WHERE name=$1 FOR UPDATE", table), original).Scan(&id); err != nil {
		return fmt.Errorf("opción no encontrada")
	}
	if method == "DELETE" {
		var used bool
		if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM services_level2 WHERE %s=$1)", columns[kind]), id).Scan(&used); err != nil {
			return err
		}
		if used {
			return fmt.Errorf("no se puede eliminar una opción utilizada por servicios")
		}
		_, err = tx.Exec(ctx, fmt.Sprintf("DELETE FROM %s WHERE id=$1", table), id)
	} else {
		// Las etiquetas originales del Excel se conservan. Solo se editan opciones nuevas.
		var sourceUsed bool
		if err := tx.QueryRow(ctx, fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM services_level2 WHERE %s=$1 AND source_sheet IS NOT NULL)", columns[kind]), id).Scan(&sourceUsed); err != nil {
			return err
		}
		if sourceUsed {
			return fmt.Errorf("la etiqueta original del Excel está protegida")
		}
		_, err = tx.Exec(ctx, fmt.Sprintf("UPDATE %s SET name=$1 WHERE id=$2", table), name, id)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
