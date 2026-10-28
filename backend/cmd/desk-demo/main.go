package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/database"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
)

func main() {
	ctx := context.Background()
	if os.Getenv("DESK_DEMO_ALLOW_WRITE") != "1" {
		log.Fatal("habilitar explícitamente DESK_DEMO_ALLOW_WRITE=1")
	}
	p, e := database.Open(ctx, os.Getenv("DATABASE_URL"))
	if e != nil {
		log.Fatal(e)
	}
	defer p.Close()
	var name string
	if e = p.QueryRow(ctx, "SELECT current_database()").Scan(&name); e != nil {
		log.Fatal(e)
	}
	if !strings.HasSuffix(name, "_desk_demo") && !strings.HasSuffix(name, "_desk_test") {
		log.Fatal("solo se permite una base aislada _desk_demo o _desk_test")
	}
	password := os.Getenv("DESK_DEMO_PASSWORD")
	if len(password) < 12 {
		log.Fatal("DESK_DEMO_PASSWORD requiere al menos 12 caracteres")
	}
	if e = database.ApplyMigrations(ctx, p, os.Getenv("MIGRATIONS_DIR")); e != nil {
		log.Fatal(e)
	}
	result, e := desk.NewService(p).PrepareDemo(ctx, password)
	if e != nil {
		log.Fatal(e)
	}
	if e = json.NewEncoder(os.Stdout).Encode(result); e != nil {
		log.Fatal(e)
	}
}
