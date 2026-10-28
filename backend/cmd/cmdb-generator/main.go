package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/database"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"log"
	"os"
	"time"
)

func main() {
	dataset := flag.String("dataset", "cmdb-demo", "conjunto sintético")
	seed := flag.Int64("seed", 202300644, "semilla")
	count := flag.Int("count", 5000, "cantidad de dispositivos")
	source := flag.String("source", "data/fuentes-service-desk/CMDB.xlsx", "Excel original para trazabilidad")
	flag.Parse()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	p, e := database.Open(ctx, os.Getenv("DATABASE_URL"))
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	if e = database.ApplyMigrations(ctx, p, os.Getenv("MIGRATIONS_DIR")); e != nil {
		log.Fatal(e)
	}
	var user auth.User
	if e = p.QueryRow(ctx, `SELECT id,role FROM app_users WHERE role='admin' AND is_active AND NOT EXISTS(SELECT 1 FROM user_roles WHERE user_id=app_users.id AND role_code='consulta') ORDER BY id LIMIT 1`).Scan(&user.ID, &user.Role); e != nil {
		log.Fatal("crear cuenta administradora e importar catálogo antes de generar")
	}
	user.Roles = []string{"admin"}
	s := desk.NewService(p)
	if e = s.ImportCMDBSource(ctx, *source); e != nil {
		log.Fatal(e)
	}
	report, e := s.Generate(ctx, &user, *dataset, *seed, *count)
	if e != nil {
		log.Fatal(e)
	}
	json.NewEncoder(os.Stdout).Encode(report)
}
