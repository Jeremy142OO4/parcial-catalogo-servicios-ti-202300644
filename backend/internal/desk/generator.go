package desk

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/xuri/excelize/v2"
	"os"
	"regexp"
)

func DeviceCode(dataset string, seed int64, ordinal int) string {
	sum := sha256.Sum256([]byte(fmt.Sprintf("v1/%s/%d/%d", dataset, seed, ordinal)))
	return "CI-" + hex.EncodeToString(sum[:12])
}
func (s *Service) Generate(ctx context.Context, u *auth.User, dataset string, seed int64, count int) (map[string]any, error) {
	if e := requireAdmin(u); e != nil {
		return nil, e
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9_-]{1,40}$`).MatchString(dataset) || count < 5000 || count > 100000 {
		return nil, bad("dataset inválido o cantidad fuera de 5000–100000")
	}
	var summary map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error {
		if _, e := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(202300647)`); e != nil {
			return e
		}
		var location int64
		if e := tx.QueryRow(ctx, `INSERT INTO locations(code,name) VALUES('SYNTH-LAB','Laboratorio sintético') ON CONFLICT(code) DO UPDATE SET name=EXCLUDED.name RETURNING id`).Scan(&location); e != nil {
			return e
		}
		var service int64
		if e := tx.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY id LIMIT 1`).Scan(&service); e != nil {
			return bad("importar catálogo antes de generar")
		}
		rows, e := tx.Query(ctx, `SELECT id,code FROM ci_classes WHERE code IN('servidor','usuario','red','impresion','almacenamiento','energia') AND is_device_class AND is_active ORDER BY code`)
		if e != nil {
			return e
		}
		type class struct {
			id   int64
			code string
		}
		classes := []class{}
		for rows.Next() {
			var c class
			if e = rows.Scan(&c.id, &c.code); e != nil {
				rows.Close()
				return e
			}
			classes = append(classes, c)
		}
		e = rows.Err()
		rows.Close()
		if e != nil {
			return e
		}
		if len(classes) != 6 {
			return bad("faltan clases sintéticas activas")
		}
		var oldCount int
		if e = tx.QueryRow(ctx, `SELECT count(*) FROM configuration_items WHERE dataset_key=$1 AND seed=$2`, dataset, seed).Scan(&oldCount); e != nil {
			return e
		}
		if oldCount > count {
			return conflict("el conjunto ya tiene más elementos, no se borra al reducir cantidad")
		}
		batch := &pgx.Batch{}
		for i := 0; i < count; i++ {
			c := classes[i%len(classes)]
			attrs := map[string]any{}
			switch c.code {
			case "servidor", "usuario":
				attrs["ram_gb"] = 16
				attrs["storage_gb"] = 512
			case "red":
				attrs["ports"] = 24
			case "almacenamiento":
				attrs["storage_gb"] = 4096
			case "energia":
				attrs["power_watts"] = 1000
			}
			batch.Queue(`INSERT INTO configuration_items(code,name,class_id,location_id,owner_user_id,description,serial_number,status,technical_attributes,is_device,dataset_key,seed) VALUES($1,$2,$3,$4,$5,'Dispositivo sintético, no dato personal',$6,'activo',$7,TRUE,$8,$9) ON CONFLICT(code) DO NOTHING`, DeviceCode(dataset, seed, i), fmt.Sprintf("Sintético %s %05d", c.code, i+1), c.id, location, u.ID, fmt.Sprintf("SYN-%d-%d", seed, i), jsonBytes(attrs), dataset, seed)
		}
		if e = tx.SendBatch(ctx, batch).Close(); e != nil {
			return e
		}
		if _, e = tx.Exec(ctx, `INSERT INTO ci_services SELECT id,$3 FROM configuration_items WHERE dataset_key=$1 AND seed=$2 ON CONFLICT DO NOTHING`, dataset, seed, service); e != nil {
			return e
		}
		counts, e := objects(ctx, tx, `SELECT jsonb_build_object('class',c.code,'devices',count(*)) FROM configuration_items i JOIN ci_classes c ON c.id=i.class_id WHERE dataset_key=$1 AND seed=$2 AND i.is_device GROUP BY c.code ORDER BY c.code`, dataset, seed)
		if e != nil {
			return e
		}
		levels, e := objects(ctx, tx, `WITH RECURSIVE ancestry AS(SELECT i.id AS ci_id,c.id AS class_id,c.parent_id,c.code,0 AS depth FROM configuration_items i JOIN ci_classes c ON c.id=i.class_id WHERE i.dataset_key=$1 AND i.seed=$2 AND i.is_device UNION ALL SELECT a.ci_id,c.id,c.parent_id,c.code,a.depth+1 FROM ancestry a JOIN ci_classes c ON c.id=a.parent_id) SELECT jsonb_build_object('category',code,'distance_from_leaf',depth,'devices',count(DISTINCT ci_id)) FROM ancestry GROUP BY code,depth ORDER BY depth,code`, dataset, seed)
		if e != nil {
			return e
		}
		var total int
		if e = tx.QueryRow(ctx, `SELECT count(DISTINCT id) FROM configuration_items WHERE dataset_key=$1 AND seed=$2 AND is_device`, dataset, seed).Scan(&total); e != nil {
			return e
		}
		if total != count {
			return conflict("el conteo del conjunto no coincide, revisar colisiones de códigos o clasificación")
		}
		summary = map[string]any{"dataset": dataset, "seed": seed, "generator_version": "v1", "devices": total, "created": total - oldCount, "existing": oldCount, "by_class": counts, "by_ancestor": levels}
		_, e = tx.Exec(ctx, `INSERT INTO generation_runs(dataset_key,seed,generator_version,requested_count,actual_count,summary) VALUES($1,$2,'v1',$3,$4,$5)`, dataset, seed, count, total, jsonBytes(summary))
		return e
	})
	return summary, err
}
func (s *Service) ImportCMDBSource(ctx context.Context, path string) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	sum := sha256.Sum256(b)
	w, e := excelize.OpenFile(path)
	if e != nil {
		return e
	}
	defer w.Close()
	tx, e := s.repo.pool.Begin(ctx)
	if e != nil {
		return e
	}
	defer tx.Rollback(ctx)
	for row := 5; row <= 6; row++ {
		raw := map[string]string{}
		for col := 2; col <= 21; col++ {
			cell, _ := excelize.CoordinatesToCellName(col, row)
			value, e := w.GetCellValue("CMDB", cell)
			if e != nil {
				return e
			}
			raw[cell] = value
		}
		_, e = tx.Exec(ctx, `INSERT INTO source_records(source_file,sha256,sheet,source_row,raw_values,observations) VALUES('CMDB.xlsx',$1,'CMDB',$2,$3,$4) ON CONFLICT(sha256,sheet,source_row) DO NOTHING`, hex.EncodeToString(sum[:]), row, jsonBytes(raw), jsonBytes([]string{"Referencias y responsables requieren resolución, no se crean CI ficticios", "Espacios y encabezados originales conservados"}))
		if e != nil {
			return e
		}
	}
	return tx.Commit(ctx)
}
