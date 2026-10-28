package desk_test

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestIndicatorsIntegration(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated database required")
	}
	parsed, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(parsed.Path, "_desk_test") {
		t.Fatal("isolated database required")
	}
	ctx := context.Background()
	p, e := pgxpool.New(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer p.Close()
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	var admin auth.User
	var service int64
	admin.Roles = []string{"admin"}
	must(p.QueryRow(ctx, `SELECT id FROM app_users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&admin.ID))
	must(p.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY id LIMIT 1`).Scan(&service))
	s := desk.NewService(p)
	first, resolution := 30, 90
	qin := desk.QueueInput{Code: fmt.Sprintf("MET-%d", time.Now().UnixNano()), Name: "Objetivos de prueba", Process: "eventos", Level: 1, First: &first, Resolution: &resolution, Active: true}
	q, e := s.SaveQueue(ctx, &admin, 0, qin)
	must(e)
	qid := int64(q["id"].(float64))
	ticket, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: "Fixture controlado P17", Description: "Datos temporales deliberados para cálculo", Requester: admin.ID, Service: service, Process: "eventos", Queue: qid, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Fixture sin CI", Details: map[string]any{"category": "E0"}})
	must(e)
	id := int64(ticket["id"].(float64))
	assigned, e := s.Act(ctx, &admin, id, desk.Action{Action: "assign", Version: 1, Assignee: &admin.ID})
	must(e)
	var unattended *time.Time
	must(p.QueryRow(ctx, `SELECT first_attended_at FROM tickets WHERE id=$1`, id).Scan(&unattended))
	if unattended != nil {
		t.Fatal("assignment incorrectly counted as human attention")
	}
	nextFirst, nextResolution := 60, 200
	qin.First = &nextFirst
	qin.Resolution = &nextResolution
	_, e = s.SaveQueue(ctx, &admin, qid, qin)
	must(e)
	anchor := time.Now().UTC().Add(-time.Hour * 3).Truncate(time.Second)
	// Synthetic controlled fixture, not production actions or invented measurements.
	_, e = p.Exec(ctx, `UPDATE tickets SET created_at=$2,first_attended_at=$2::timestamptz+interval '60 minutes',resolved_at=$2::timestamptz+interval '120 minutes',state='resuelto',resolution='Fixture' WHERE id=$1`, id, anchor)
	must(e)
	_, e = p.Exec(ctx, `UPDATE ticket_cycles SET started_at=$2,resolved_at=$2::timestamptz+interval '120 minutes',resolution='Fixture' WHERE ticket_id=$1`, id, anchor)
	must(e)
	_, e = p.Exec(ctx, `UPDATE ticket_objective_segments SET started_at=$2,ended_at=$2::timestamptz+interval '120 minutes' WHERE ticket_id=$1`, id, anchor)
	must(e)
	_, e = p.Exec(ctx, `INSERT INTO ticket_pauses(ticket_id,started_at,ended_at,reason) VALUES($1,$2::timestamptz+interval '30 minutes',$2::timestamptz+interval '60 minutes','Fixture de pausa')`, id, anchor)
	must(e)
	report, e := s.Indicators(ctx, &admin, id, "", "")
	must(e)
	metric := report["items"].([]map[string]any)[0]
	cycle := metric["cycles"].([]map[string]any)[0]
	seg := metric["objective_segments"].([]map[string]any)[0]
	if *metric["first_attention_minutes"].(*float64) != 30 || *cycle["resolution_minutes"].(*float64) != 90 || cycle["status"] != "cumplido" || *seg["resolution_minutes"].(*float64) != 90 {
		t.Fatal("controlled metrics or historical snapshot changed", metric)
	}
	reopened, e := s.Act(ctx, &admin, id, desk.Action{Action: "reopen", Version: int64(assigned["version"].(float64)), Reason: "Fixture de reapertura"})
	must(e)
	report, e = s.Indicators(ctx, &admin, id, "", "")
	must(e)
	metric = report["items"].([]map[string]any)[0]
	if metric["reopens"] != 1 || len(metric["cycles"].([]map[string]any)) != 2 {
		t.Fatal("reopening history", metric)
	}
	segs := metric["objective_segments"].([]map[string]any)
	if *segs[1]["resolution_minutes"].(*float64) != 200 {
		t.Fatal("new cycle snapshot missing")
	}
	_, e = s.Act(ctx, &admin, id, desk.Action{Action: "pause", Version: int64(reopened["version"].(float64)), Reason: "Esperar información"})
	must(e)
	version := int64(reopened["version"].(float64)) + 1
	if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: "clasificado", Version: version}); e == nil {
		t.Fatal("transition during pause accepted")
	}
	_, e = s.Act(ctx, &admin, id, desk.Action{Action: "resume", Version: version})
	must(e)
	qin.Code += "-DEST"
	qin.First = nil
	qin.Resolution = nil
	dest, e := s.SaveQueue(ctx, &admin, 0, qin)
	must(e)
	_, e = s.Act(ctx, &admin, id, desk.Action{Action: "transfer", Version: version + 1, Queue: int64(dest["id"].(float64)), Reason: "Fixture de transferencia"})
	must(e)
	report, e = s.Indicators(ctx, &admin, id, "", "")
	must(e)
	metric = report["items"].([]map[string]any)[0]
	segs = metric["objective_segments"].([]map[string]any)
	if len(segs) != 3 || segs[2]["resolution_minutes"].(*float64) != nil || metric["reopens"] != 1 {
		t.Fatal("transfer goals", metric)
	}
	own, e := s.OwnTickets(ctx, &admin, "Fixture controlado P17", "", 20, 0)
	must(e)
	if len(own) == 0 {
		t.Fatal("portal missing own ticket")
	}
	var position int64
	must(p.QueryRow(ctx, `SELECT position_id FROM app_users WHERE id=$1`, admin.ID).Scan(&position))
	prefix := fmt.Sprintf("MET-REQ-%d", time.Now().UnixNano())
	outsider, e := auth.NewService(p, time.Hour).CreateUser(ctx, position, "Solicitante aislado P17", prefix, prefix+"@example.local", "TemporalDemo123!", "solicitante")
	must(e)
	outsider.Roles = []string{"solicitante"}
	if _, e = s.Indicators(ctx, &outsider, id, "", ""); e == nil {
		t.Fatal("metrics exposed another requester's ticket")
	}
	report, e = s.Indicators(ctx, &outsider, 0, "", "")
	must(e)
	if report["total"] != 0 {
		t.Fatal("dashboard ACL leaked tickets")
	}
	if _, e = s.SaveQueue(ctx, &admin, int64(dest["id"].(float64)), desk.QueueInput{Code: qin.Code, Name: qin.Name, Process: "eventos", Level: 1, Active: false}); e == nil {
		t.Fatal("deactivated queue with pending ticket")
	}
	t.Log("P17 PostgreSQL: primera atención 30, resolución 90, objetivo histórico 90, ciclo nuevo 200, transferencia sin objetivo y pausa controlada")
}
