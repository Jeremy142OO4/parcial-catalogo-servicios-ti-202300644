package importer

import (
	"path/filepath"
	"testing"
)

func TestCatalogoOriginalConservaLosControlesDeImportacion(t *testing.T) {
	path := filepath.Join("..", "..", "..", "data", "CatalogoServicios.xlsx")
	report, err := Run(path, DefaultSheet)
	if err != nil {
		t.Fatalf("importación inesperadamente fallida: %v", err)
	}
	if report.Counts.Level1 != ExpectedLevel1 {
		t.Fatalf("nivel 1: got %d, want %d", report.Counts.Level1, ExpectedLevel1)
	}
	if report.Counts.Level2 != ExpectedLevel2 {
		t.Fatalf("nivel 2: got %d, want %d", report.Counts.Level2, ExpectedLevel2)
	}
	if report.ServicesLevel1[11].Code != "SE.12" || report.ServicesLevel1[11].Name != "Suministrar Analitica" {
		t.Fatalf("SE.12 canonical value was not taken from the first source occurrence")
	}
	if len(report.ServicesLevel1[11].SourceNames) != 2 || report.ServicesLevel1[11].SourceNames[1].Value != "Mantener Tableros de Control" {
		t.Fatalf("SE.12 source evidence must preserve both original names")
	}
	for _, service := range report.ServicesLevel2 {
		if service.Code == "SE.12.3" {
			if service.ParentCode != "SE.12" {
				t.Fatalf("SE.12.3 parent: got %s, want SE.12", service.ParentCode)
			}
			if !service.ReviewRequired || service.Active != nil || service.ServiceClass != nil || service.Criticality != nil || service.ServiceType != nil || service.Metric != nil || service.Minimum != nil || service.Maximum != nil {
				t.Fatalf("SE.12.3 must preserve incomplete attributes as unknown and require review")
			}
		}
	}
}

func TestParentCodeFromLevel2Code(t *testing.T) {
	for input, want := range map[string]string{
		"SE.12.1": "SE.12",
		"SE.12.2": "SE.12",
		"SE.12.3": "SE.12",
	} {
		if got := parentCodeFromLevel2(input); got != want {
			t.Fatalf("parentCodeFromLevel2(%q) = %q, want %q", input, got, want)
		}
	}
}
