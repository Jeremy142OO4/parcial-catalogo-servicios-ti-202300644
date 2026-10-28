package desk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/httpapi"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestServiceDeskIntegration(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("DESK_TEST_DATABASE_URL required, isolated database only")
	}
	parsed, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(parsed.Path, "_desk_test") {
		t.Fatal("test database name must end in _desk_test")
	}
	ctx := context.Background()
	p, e := pgxpool.New(ctx, dsn)
	if e != nil {
		t.Fatal(e)
	}
	defer p.Close()
	prefix := fmt.Sprintf("IT-%d", time.Now().UnixNano())
	s := desk.NewService(p)
	a := auth.NewService(p, time.Hour)
	var admin auth.User
	var position, service int64
	if e = p.QueryRow(ctx, `SELECT id,position_id FROM app_users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&admin.ID, &position); e != nil {
		t.Fatal("seed isolated database first", e)
	}
	admin.Role = "admin"
	admin.Roles = []string{"admin"}
	if e = p.QueryRow(ctx, `SELECT id FROM services_level2 ORDER BY id LIMIT 1`).Scan(&service); e != nil {
		t.Fatal(e)
	}
	must := func(e error) {
		t.Helper()
		if e != nil {
			t.Fatal(e)
		}
	}
	log := func(label string) { t.Log("PASS", label) }
	agent, e := a.CreateUser(ctx, position, "Agente sintético", prefix+"-agent", prefix+"-agent@example.local", "TemporalDemo123!", "agente")
	must(e)
	agent.Roles = []string{"agente"}
	agent2, e := a.CreateUser(ctx, position, "Agente 2 sintético", prefix+"-agent2", prefix+"-agent2@example.local", "TemporalDemo123!", "agente")
	must(e)
	agent2.Roles = []string{"agente"}
	requester, e := a.CreateUser(ctx, position, "Solicitante sintético", prefix+"-req", prefix+"-req@example.local", "TemporalDemo123!", "solicitante")
	must(e)
	requester.Roles = []string{"solicitante"}
	consulta, e := a.CreateUser(ctx, position, "Consulta sintético", prefix+"-read", prefix+"-read@example.local", "TemporalDemo123!", "consulta")
	must(e)
	must(a.SetRoles(ctx, consulta.ID, []string{"admin", "agente", "consulta"}))
	consulta.Roles = []string{"admin", "agente", "consulta"}
	must(a.SetRoles(ctx, requester.ID, []string{"consulta"}))
	_, e = s.CreateTicket(ctx, &requester, desk.TicketInput{})
	var denied *desk.Error
	if !errors.As(e, &denied) || denied.Status != 403 {
		t.Fatal("stale role snapshot did not return forbidden", e)
	}
	must(a.SetRoles(ctx, requester.ID, []string{"solicitante"}))
	queues, e := s.Queues(ctx, &admin)
	must(e)
	qs := map[string]int64{}
	for _, q := range queues {
		code := q["process_code"].(string)
		if qs[code] == 0 {
			isolated, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix + "-" + code, Name: "Cola aislada " + code, Process: code, Level: 1, Active: true})
			must(e)
			qs[code] = int64(isolated["id"].(float64))
		}
	}
	for _, q := range qs {
		r := "agente"
		must(s.Grant(ctx, &admin, q, desk.GrantInput{Role: &r, Read: true, Handle: true, Transfer: true}))
	}
	if _, e = s.SaveClass(ctx, &consulta, 0, desk.ClassInput{Code: prefix + "-denied", Name: "No", IsActive: true}); e == nil {
		t.Fatal("consulta writes")
	}
	log("P01 consulta con roles elevados sigue sin escritura")
	var n1, n2 int
	must(p.QueryRow(ctx, `SELECT count(*) FROM services_level1 WHERE source_sheet IS NOT NULL`).Scan(&n1))
	must(p.QueryRow(ctx, `SELECT count(*) FROM services_level2 WHERE source_sheet IS NOT NULL`).Scan(&n2))
	if n1 != 12 || n2 != 46 {
		t.Fatal("catalog regression", n1, n2)
	}
	log("P02 regresión de controles de origen 12/46")
	c1, e := s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-A", Name: "A", IsActive: true})
	must(e)
	id1 := int64(c1["id"].(float64))
	c2, e := s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-B", Name: "B", ParentID: &id1, IsActive: true})
	must(e)
	id2 := int64(c2["id"].(float64))
	if _, e = s.SaveClass(ctx, &admin, id1, desk.ClassInput{Code: prefix + "-A", Name: "A", ParentID: &id2, IsActive: true}); e == nil {
		t.Fatal("classification cycle")
	}
	invalid := int64(999999999)
	if _, e = s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-orphan", Name: "bad", ParentID: &invalid, IsActive: true}); e == nil {
		t.Fatal("orphan")
	}
	log("P04 padres y ciclos de clasificación rechazados")
	report, e := s.Generate(ctx, &admin, "integration", 202300644, 5000)
	must(e)
	if report["devices"] != 5000 {
		t.Fatal(report)
	}
	report, e = s.Generate(ctx, &admin, "integration", 202300644, 5000)
	must(e)
	if report["created"] != 0 {
		t.Fatal(report)
	}
	log("P03 5000 dispositivos y repetición idempotente")
	items, e := s.ListCI(ctx, "", 0, "", 100, 0)
	must(e)
	if len(items) != 100 {
		t.Fatal("pagination")
	}
	location := int64(items[0]["location_id"].(float64))
	ciInput := desk.CIInput{Code: prefix + "-CI1", Name: "CI aislado", ClassID: id1, LocationID: location, OwnerID: admin.ID, Status: "activo", IsActive: true}
	first, e := s.SaveCI(ctx, &admin, 0, ciInput)
	must(e)
	ci1 := int64(first["id"].(float64))
	ciInput.Code = prefix + "-CI2"
	second, e := s.SaveCI(ctx, &admin, 0, ciInput)
	must(e)
	ci2 := int64(second["id"].(float64))
	ciInput.Name = "CI editado"
	_, e = s.SaveCI(ctx, &admin, ci2, ciInput)
	must(e)
	ciInput.IsActive = false
	_, e = s.SaveCI(ctx, &admin, ci2, ciInput)
	must(e)
	ciInput.IsActive = true
	_, e = s.SaveCI(ctx, &admin, ci2, ciInput)
	must(e)
	log("CRUD CI: crear, consultar, editar y baja lógica/reactivación")
	must(s.Relation(ctx, &admin, 0, ci1, ci2, "depende_de", false))
	if e = s.Relation(ctx, &admin, 0, ci2, ci1, "depende_de", false); e == nil {
		t.Fatal("technical cycle")
	}
	impact, e := s.Impact(ctx, ci2)
	must(e)
	if len(impact) < 2 {
		t.Fatal(impact)
	}
	log("P14 dependencias y ciclos técnicos")
	start := time.Now()
	_, e = s.ListCI(ctx, "Sintético", 0, "activo", 100, 100)
	must(e)
	t.Log("P16 tiempo consulta paginada", time.Since(start), "entorno: PostgreSQL aislado en laptop")
	base := func(process string) desk.TicketInput {
		return desk.TicketInput{Title: prefix + " " + process, Description: "Prueba aislada", Requester: requester.ID, Service: service, Process: process, Queue: qs[process], Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Trámite sintético sin dispositivo afectado", Details: map[string]any{}}
	}
	create := func(process string, d map[string]any) map[string]any {
		t.Helper()
		in := base(process)
		in.Details = d
		out, e := s.CreateTicket(ctx, &admin, in)
		must(e)
		return out
	}
	act := func(ticket map[string]any, who *auth.User, input desk.Action) map[string]any {
		t.Helper()
		input.Version = int64(ticket["version"].(float64))
		out, e := s.Act(ctx, who, int64(ticket["id"].(float64)), input)
		must(e)
		return out
	}
	in := base("servicios")
	in.Service = 999999999
	if _, e = s.CreateTicket(ctx, &admin, in); e == nil {
		t.Fatal("invalid reference")
	}
	in = base("servicios")
	in.NoCI = ""
	if _, e = s.CreateTicket(ctx, &admin, in); e == nil {
		t.Fatal("missing no-CI reason")
	}
	log("P05 referencias y condición sin CI")
	ticket := create("servicios", map[string]any{})
	id := int64(ticket["id"].(float64))
	if _, e = s.Ticket(ctx, &agent2, id); e != nil {
		t.Fatal(e)
	}
	// Two goroutines race on the same version, only one must claim successfully.
	var wg sync.WaitGroup
	outcomes := make(chan error, 2)
	for _, user := range []*auth.User{&agent, &agent2} {
		wg.Add(1)
		go func(user *auth.User) {
			defer wg.Done()
			_, e := s.Act(ctx, user, id, desk.Action{Action: "claim", Version: 1})
			outcomes <- e
		}(user)
	}
	wg.Wait()
	close(outcomes)
	success := 0
	for e := range outcomes {
		if e == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("claims: %d", success)
	}
	log("P07 dos reclamaciones, un ganador")
	destination, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix + "-dest", Name: "Destino", Process: "servicios", Level: 2, Active: true})
	must(e)
	current, e := s.Ticket(ctx, &admin, id)
	must(e)
	_, e = s.Act(ctx, &admin, id, desk.Action{Action: "transfer", Queue: int64(destination["id"].(float64)), Reason: "Transferencia aislada", Version: int64(current["version"].(float64))})
	must(e)
	current, e = s.Ticket(ctx, &admin, id)
	must(e)
	if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "transfer", Queue: qs["cambios"], Reason: "Inválida", Version: int64(current["version"].(float64))}); e == nil {
		t.Fatal("cross process transfer")
	}
	log("P08 transferencia e incompatibilidad de proceso")
	must(s.Comment(ctx, &admin, id, "Nota secreta sintética", "internal"))
	external, e := s.Ticket(ctx, &requester, id)
	must(e)
	if len(external["comments"].([]map[string]any)) != 0 {
		t.Fatal("internal note leaked")
	}
	if e = s.Comment(ctx, &consulta, id, "write", "public"); e == nil {
		t.Fatal("consulta comment")
	}
	log("P06 ámbito y notas internas")
	completed := map[string]map[string]any{}
	for process, route := range desk.Routes {
		details := map[string]any{"need": "Necesidad", "feasibility": "Viable", "proposed_service_name": "Nuevo", "impact_assessment": "Bajo", "risk": "Bajo", "test_plan": "Prueba", "rollback_plan": "Revertir", "planned_at": "2026-10-28T10:00:00Z", "execution_result": "Ejecutado", "validation_result": "Correcto", "category": "E0", "evaluation": "Programado", "root_cause": "Causa", "workaround": "Temporal", "definitive_solution": "Definitiva", "beneficiary_id": float64(requester.ID), "resource": "Recurso", "permission": "Leer", "operation": "grant", "provision_evidence": "Concedido", "fulfillment_evidence": "Cumplido", "confirmation": "Confirmado"}
		if process == "incidentes" {
			details["category"] = "I0"
			details["security"] = false
			details["interrupted"] = false
			details["redundancy"] = true
			details["solution"] = "Sin interrupción"
		}
		current := create(process, details)
		tid := int64(current["id"].(float64))
		if _, e = s.Act(ctx, &admin, tid, desk.Action{Action: "transition", To: "cerrado", Version: 1}); e == nil {
			t.Fatal("invalid transition", process)
		}
		// Assign agent so service fulfillment can enter assigned state.
		current = act(current, &admin, desk.Action{Action: "assign", Assignee: &agent.ID})
		for _, next := range route[1:] {
			if next == "aprobado" {
				current = act(current, &admin, desk.Action{Action: "approve", Reason: "Evaluación autorizada"})
			} else {
				current = act(current, &admin, desk.Action{Action: "transition", To: next, Reason: "Confirmación de prueba", Resolution: "Solución comprobada"})
			}
			if process == "nuevos" && next == "implementacion" {
				current = act(current, &admin, desk.Action{Action: "link_service", Details: map[string]any{"service_id": float64(service)}})
			}
			if process == "cambios" && next == "validacion" {
				if _, e = s.Act(ctx, &admin, tid, desk.Action{Action: "details", Version: int64(current["version"].(float64)), Details: map[string]any{"risk": "cambiado"}}); e == nil {
					t.Fatal("approved scope modified")
				}
			}
		}
		if current["state"] != "cerrado" {
			t.Fatal(process)
		}
		log("P09 ruta completa " + process)
		completed[process] = current
		if process == "cambios" {
			current = act(current, &admin, desk.Action{Action: "reopen", Reason: "Reevaluación"})
			if int(current["cycle"].(float64)) != 2 {
				t.Fatal("reopen cycle")
			}
			log("P15 cierre y reapertura con ciclos")
		}
	}
	must(s.LinkTickets(ctx, &admin, int64(completed["incidentes"]["id"].(float64)), int64(completed["problemas"]["id"].(float64)), "incidente_problema"))
	must(s.LinkTickets(ctx, &admin, int64(completed["problemas"]["id"].(float64)), int64(completed["cambios"]["id"].(float64)), "problema_cambio"))
	log("P13 incidente vinculado a problema y cambio")
	access := base("accesos")
	access.Requester = admin.ID
	access.Details = map[string]any{"beneficiary_id": float64(requester.ID), "resource": "R", "permission": "P", "operation": "grant"}
	self, e := s.CreateTicket(ctx, &admin, access)
	must(e)
	self = act(self, &admin, desk.Action{Action: "transition", To: "validacion"})
	self = act(self, &admin, desk.Action{Action: "transition", To: "pendiente_aprobacion"})
	if _, e = s.Act(ctx, &admin, int64(self["id"].(float64)), desk.Action{Action: "approve", Version: int64(self["version"].(float64)), Reason: "self"}); e == nil {
		t.Fatal("self approval")
	}
	log("P10 autoaprobación admin rechazada")
	for _, category := range []string{"E0", "E1", "E2", "E3"} {
		ev := create("eventos", map[string]any{"category": category, "evaluation": "Evaluación"})
		ev = act(ev, &admin, desk.Action{Action: "transition", To: "clasificado"})
		ev = act(ev, &admin, desk.Action{Action: "transition", To: "evaluado"})
		if category == "E2" || category == "E3" {
			ev = act(ev, &admin, desk.Action{Action: "process_event"})
			ev = act(ev, &admin, desk.Action{Action: "process_event"})
			var count int
			must(p.QueryRow(ctx, `SELECT count(*) FROM event_incidents WHERE event_ticket_id=$1`, int64(ev["id"].(float64))).Scan(&count))
			if count != 1 {
				t.Fatal("duplicate event")
			}
		}
	}
	log("P11 E0–E3 y deduplicación")
	// API checks prove authorization is enforced at the HTTP boundary, not only UI.
	server := httptest.NewServer(httpapi.NewServer(p).Handler())
	defer server.Close()
	req := func(method, path string, body any, token string) int {
		t.Helper()
		payload, _ := json.Marshal(body)
		r, _ := http.NewRequest(method, server.URL+path, bytes.NewReader(payload))
		r.Header.Set("Authorization", "Bearer "+token)
		response, e := http.DefaultClient.Do(r)
		must(e)
		defer response.Body.Close()
		return response.StatusCode
	}
	if req("GET", "/api/desk/tickets", nil, "") != 401 {
		t.Fatal("missing auth")
	}
	session, e := a.Login(ctx, consulta.Username, "TemporalDemo123!")
	must(e)
	if req("POST", "/api/cmdb/classes", map[string]any{"code": "bad", "name": "bad"}, session.Token) != 403 {
		t.Fatal("HTTP consulta")
	}
	must(a.Logout(ctx, session.Token))
	if req("GET", "/api/desk/tickets", nil, session.Token) != 401 {
		t.Fatal("logout")
	}
	log("P01 HTTP sin sesión, consulta y logout")
}
