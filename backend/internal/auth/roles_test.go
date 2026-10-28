package auth

import "testing"

func TestConsultaCannotBeElevated(t *testing.T) {
	u := User{Role: "admin", Roles: []string{"admin", "agente", "consulta"}}
	if !u.ReadOnly() {
		t.Fatal("consulta elevated")
	}
	if (&User{Role: "consulta"}).ReadOnly() != true {
		t.Fatal("legacy role")
	}
}
