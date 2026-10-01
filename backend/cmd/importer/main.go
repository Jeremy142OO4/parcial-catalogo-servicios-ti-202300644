package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/catalog"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/database"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
)

func main() {
	input := flag.String("input", "data/CatalogoServicios.xlsx", "ruta del Excel original")
	reportPath := flag.String("report", "outputs/import-report.json", "ruta del reporte JSON")
	sheet := flag.String("sheet", importer.DefaultSheet, "hoja que se importará")
	databaseURL := flag.String("database-url", "", "URL de PostgreSQL; si se indica, persiste el resultado")
	migrationsDir := flag.String("migrations-dir", "backend/migrations", "directorio de migraciones")
	flag.Parse()

	report, err := importer.Run(*input, *sheet)
	if report != nil {
		if writeErr := importer.WriteJSON(report, *reportPath); writeErr != nil {
			log.Fatalf("no se pudo escribir el reporte: %v", writeErr)
		}
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "importación fallida: %v\n", err)
		os.Exit(1)
	}
	if *databaseURL != "" {
		ctx := context.Background()
		pool, dbErr := database.Open(ctx, *databaseURL)
		if dbErr != nil {
			log.Fatalf("no se pudo abrir PostgreSQL: %v", dbErr)
		}
		defer pool.Close()
		if dbErr = database.ApplyMigrations(ctx, pool, *migrationsDir); dbErr != nil {
			log.Fatalf("no se pudieron aplicar las migraciones: %v", dbErr)
		}
		if dbErr = catalog.NewService(catalog.NewRepository(pool)).PersistImport(ctx, report); dbErr != nil {
			log.Fatalf("no se pudo persistir la importación: %v", dbErr)
		}
		fmt.Println("Persistencia PostgreSQL: OK")
	}

	fmt.Printf("Importación validada: %d nivel 1, %d nivel 2, %d observaciones, %d filas omitidas\n", report.Counts.Level1, report.Counts.Level2, report.Counts.Observed, report.Counts.Omitted)
}
