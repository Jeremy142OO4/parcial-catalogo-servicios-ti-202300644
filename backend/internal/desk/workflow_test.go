package desk

import "testing"

func TestIncidentBoundaries(t *testing.T) {
	for _, tc := range []struct {
		minutes  float64
		category string
		review   bool
	}{{1, "I1", false}, {180, "I1", false}, {181, "I2", false}, {480, "I2", false}, {481, "I3", false}, {2880, "I3", false}, {2881, "I4", false}, {10080, "I4", false}, {10081, "I5", true}, {43199, "I5", true}, {43200, "I5", false}} {
		v, r, e := IncidentRule(false, true, &tc.minutes)
		if e != nil || v != tc.category || r != tc.review {
			t.Fatalf("%+v: %s %v %v", tc, v, r, e)
		}
	}
	negative := -1.0
	if _, _, e := IncidentRule(false, true, &negative); e == nil {
		t.Fatal("negative accepted")
	}
	if c, _, _ := IncidentRule(true, true, nil); c != "I6" {
		t.Fatal(c)
	}
	if c, _, _ := IncidentRule(false, false, nil); c != "I0" {
		t.Fatal(c)
	}
	if c, r, _ := IncidentRule(false, true, nil); c != "" || !r {
		t.Fatal("unknown duration lost")
	}
	zero := 0.0
	if c, r, _ := IncidentRule(false, true, &zero); c != "" || !r {
		t.Fatal("zero duration must require review")
	}
	details := map[string]any{"category": "I1", "security": false, "interrupted": true, "estimated_minutes": float64(60), "redundancy": true}
	if validateIncident(details) == nil {
		t.Fatal("contradictory redundancy without justification")
	}
	details["redundancy_reason"] = "Respaldo presente pero falló en este caso"
	if e := validateIncident(details); e != nil {
		t.Fatal(e)
	}
	details["redundancy"] = nil
	if e := validateIncident(details); e != nil {
		t.Fatal("unknown redundancy must remain unknown", e)
	}
}
func TestEveryRouteAndInvalidJump(t *testing.T) {
	if len(Routes) != 7 {
		t.Fatal("seven processes required")
	}
	for process, states := range Routes {
		for i := 0; i < len(states)-1; i++ {
			if !validTransition(process, states[i], states[i+1]) {
				t.Fatal(process, states[i])
			}
		}
		if validTransition(process, states[0], "cerrado") {
			t.Fatal("invalid jump")
		}
	}
}
func TestDeviceCodes(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 5000; i++ {
		a := DeviceCode("test", 42, i)
		if a != DeviceCode("test", 42, i) || seen[a] {
			t.Fatal("not deterministic or duplicate")
		}
		seen[a] = true
	}
	if DeviceCode("test", 42, 0) == DeviceCode("test", 43, 0) {
		t.Fatal("seed ignored")
	}
}
