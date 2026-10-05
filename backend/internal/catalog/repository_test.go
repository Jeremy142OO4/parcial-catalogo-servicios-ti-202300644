package catalog

import "testing"

func TestLevel2Validation(t *testing.T) {
	valid := Level2Input{Level1ID: 1, Code: "SE.12.1", Name: "Servicio"}
	if err := validateLevel2Input(valid); err != nil {
		t.Fatal(err)
	}
	invalidActive := "X"
	invalid := valid
	invalid.ActiveValue = &invalidActive
	if validateLevel2Input(invalid) == nil {
		t.Fatal("activo inválido aceptado")
	}
	low, high := 10.0, 2.0
	invalid = valid
	invalid.Minimum, invalid.Maximum = &low, &high
	if validateLevel2Input(invalid) == nil {
		t.Fatal("rango invertido aceptado")
	}
	invalid.Maximum = nil
	if err := validateLevel2Input(invalid); err != nil {
		t.Fatal("un máximo ausente no debe transformarse en cero", err)
	}
}
