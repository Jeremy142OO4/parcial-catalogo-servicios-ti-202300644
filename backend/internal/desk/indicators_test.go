package desk

import (
	"testing"
	"time"
)

func TestIndicatorIntervalsAndObjectives(t *testing.T) {
	start := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	at := func(m int) time.Time { return start.Add(time.Duration(m) * time.Minute) }
	end1, end2 := at(60), at(75)
	active, paused := activeMinutes(start, at(120), []Interval{{Start: at(30), End: &end1}, {Start: at(45), End: &end2}})
	if active != 75 || paused != 45 {
		t.Fatalf("overlap active=%v paused=%v", active, paused)
	}
	active, paused = activeMinutes(start, at(60), []Interval{{Start: at(50)}})
	if active != 50 || paused != 10 {
		t.Fatal("open pause", active, paused)
	}
	goal := 90.0
	if objectiveStatus(90, &goal, true) != "cumplido" || objectiveStatus(90.1, &goal, true) != "incumplido" || objectiveStatus(0, nil, true) != "sin_objetivo" || objectiveStatus(50, &goal, false) != "pendiente" {
		t.Fatal("boundary/unknown status")
	}
	t.Log("P17 unión de pausas solapadas, pausa abierta, objetivo desconocido y frontera inclusiva")
}

func TestIndicatorControlledCycles(t *testing.T) {
	start := time.Date(2026, 10, 28, 0, 0, 0, 0, time.UTC)
	at := func(m int) *time.Time { v := start.Add(time.Duration(m) * time.Minute); return &v }
	src := metricSource{}
	src.Ticket.ID = 1
	src.Ticket.Created = start
	src.Ticket.First = at(60)
	src.Ticket.State = "resuelto"
	src.Pauses = []Interval{{Start: *at(30), End: at(60)}}
	src.Cycles = []CycleSource{{Cycle: 1, Start: start, Resolved: at(120)}}
	first, resolution := 30.0, 90.0
	src.Segments = []ObjectiveSegment{{ID: 1, Cycle: 1, Start: start, End: at(120), First: &first, Resolution: &resolution}}
	m := calculateIndicators(src, *at(180))
	if *m["first_attention_minutes"].(*float64) != 30 {
		t.Fatal("first attention", m)
	}
	cs := m["cycles"].([]map[string]any)
	if *cs[0]["resolution_minutes"].(*float64) != 90 || cs[0]["status"] != "cumplido" {
		t.Fatal("120-30=90", cs)
	}
	src.Cycles = append(src.Cycles, CycleSource{Cycle: 2, Start: *at(120)})
	src.Segments = append(src.Segments, ObjectiveSegment{ID: 2, Cycle: 2, Start: *at(120)})
	src.Pauses = append(src.Pauses, Interval{Start: *at(170)})
	src.Reopens = 1
	src.Ticket.State = "registrado"
	m = calculateIndicators(src, *at(180))
	cs = m["cycles"].([]map[string]any)
	if cs[1]["resolution_minutes"].(*float64) != nil || cs[1]["active_minutes"] != float64(50) || cs[1]["paused_minutes"] != float64(10) || m["reopens"] != 1 {
		t.Fatal("open/reopened cycle", m)
	}
	t.Log("P17 resolución 120-30=90, abierto 60-10=50 sin resolución ficticia, ciclo anterior conservado")
}

func TestApprovedEvidenceDoesNotChangeScope(t *testing.T){
 before:=map[string]any{"risk":"alto","execution_result":"pendiente"}
 if !evidenceOnly("cambios",before,map[string]any{"risk":"alto","execution_result":"prueba ejecutada","validation_result":"aceptado"}){t.Fatal("evidence update blocked")}
 if evidenceOnly("cambios",before,map[string]any{"risk":"bajo","execution_result":"ejecutado"}){t.Fatal("approved risk changed")}
 if evidenceOnly("cambios",before,map[string]any{"execution_result":"ejecutado"}){t.Fatal("approved scope removed")}
}
