package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/catalog"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/database"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
)

func main() {
	ctx := context.Background()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL es obligatorio")
	}
	adminPassword := os.Getenv("SEED_ADMIN_PASSWORD")
	consultaPassword := os.Getenv("SEED_CONSULTA_PASSWORD")
	if adminPassword == "" || consultaPassword == "" {
		log.Fatal("SEED_ADMIN_PASSWORD y SEED_CONSULTA_PASSWORD son obligatorios")
	}
	adminHash, err := auth.HashPassword(adminPassword)
	if err != nil {
		log.Fatal(err)
	}
	consultaHash, err := auth.HashPassword(consultaPassword)
	if err != nil {
		log.Fatal(err)
	}

	pool, err := database.Open(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := database.ApplyMigrations(ctx, pool, os.Getenv("MIGRATIONS_DIR")); err != nil {
		log.Fatal(err)
	}
	report, err := importer.Run("data/CatalogoServicios.xlsx", importer.DefaultSheet)
	if err != nil {
		log.Fatal(err)
	}
	if err := catalog.NewRepository(pool).PersistImport(ctx, report); err != nil {
		log.Fatal(err)
	}

	var companyID, areaID, departmentID, sectionID, positionID int64
	err = pool.QueryRow(ctx, `
		INSERT INTO companies (code, name) VALUES ('DEMO', 'Empresa demostración')
		ON CONFLICT (code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`).Scan(&companyID)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO areas (company_id, code, name) VALUES ($1, 'TI', 'Tecnología de la Información')
		ON CONFLICT (company_id, code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, companyID).Scan(&areaID)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO departments (area_id, code, name) VALUES ($1, 'SERV', 'Servicios TI')
		ON CONFLICT (area_id, code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, areaID).Scan(&departmentID)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO sections (department_id, code, name) VALUES ($1, 'CAT', 'Catálogo de servicios')
		ON CONFLICT (department_id, code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, departmentID).Scan(&sectionID)
	if err != nil {
		log.Fatal(err)
	}
	err = pool.QueryRow(ctx, `
		INSERT INTO positions (section_id, code, name) VALUES ($1, 'COORD', 'Coordinación de catálogo')
		ON CONFLICT (section_id, code) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, sectionID).Scan(&positionID)
	if err != nil {
		log.Fatal(err)
	}

	if err := upsertUser(ctx, pool, positionID, "Administrador demostración", "admin-demo", "admin-demo@example.local", adminHash, "admin"); err != nil {
		log.Fatal(err)
	}
	if err := upsertUser(ctx, pool, positionID, "Usuario consulta demostración", "consulta-demo", "consulta-demo@example.local", consultaHash, "consulta"); err != nil {
		log.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `
		INSERT INTO service_assignments (service_level2_id, section_id, responsible_user_id)
		SELECT s.id, $1, u.id
		FROM services_level2 s CROSS JOIN app_users u
		WHERE u.username = 'admin-demo'
		ORDER BY s.code LIMIT 3
		ON CONFLICT (service_level2_id) DO UPDATE SET section_id = EXCLUDED.section_id, responsible_user_id = EXCLUDED.responsible_user_id`, sectionID); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Datos de demostración creados: organización, usuarios y tres asignaciones")
}

func upsertUser(ctx context.Context, pool *pgxpool.Pool, positionID int64, fullName, username, email, passwordHash, role string) error {
	var id int64
	return pool.QueryRow(ctx, `
		INSERT INTO app_users (position_id, full_name, username, email, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (username) DO UPDATE SET
			position_id = EXCLUDED.position_id,
			full_name = EXCLUDED.full_name,
			email = EXCLUDED.email,
			password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			is_active = TRUE
		RETURNING id`, positionID, fullName, username, email, passwordHash, role).Scan(&id)
}
