package desk_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/catalog"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
	"github.com/xuri/excelize/v2"
)

func TestFinalAcceptance(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("requires isolated PostgreSQL")
	}
	u, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(u.Path, "_desk_test") {
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
	var position, service int64
	must(p.QueryRow(ctx, `SELECT id,position_id FROM app_users WHERE username='admin-demo'`).Scan(&admin.ID, &position))
	admin.Role = "admin"
	admin.Roles = []string{"admin"}
	must(p.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY code LIMIT 1`).Scan(&service))
	a := auth.NewService(p, time.Hour)
	s := desk.NewService(p)
	prefix := fmt.Sprintf("FINAL-%d", time.Now().UnixNano())
	reader, e := a.CreateUser(ctx, position, "Consulta final", prefix+"read", prefix+"read@example.local", "TemporalFinal123!", "consulta")
	must(e)
	reader.Roles = []string{"consulta"}
	t.Run("P01_inactive_session", func(t *testing.T) {
		session, e := a.Login(ctx, reader.Username, "TemporalFinal123!")
		must(e)
		must(a.SetUserActive(ctx, reader.ID, false))
		if _, e = a.Authenticate(ctx, session.Token); e == nil {
			t.Fatal("inactive session accepted")
		}
		if _, e = a.Login(ctx, reader.Username, "TemporalFinal123!"); e == nil {
			t.Fatal("inactive login accepted")
		}
		must(a.SetUserActive(ctx, reader.ID, true))
	})
	t.Run("P02_reimport", func(t *testing.T) {
		report, e := importer.Run("data/CatalogoServicios.xlsx", importer.DefaultSheet)
		must(e)
		repo := catalog.NewRepository(p)
		must(repo.PersistImport(ctx, report))
		must(repo.PersistImport(ctx, report))
		var n1, n2 int
		must(p.QueryRow(ctx, `SELECT (SELECT count(*) FROM services_level1 WHERE source_sheet IS NOT NULL),(SELECT count(*) FROM services_level2 WHERE source_sheet IS NOT NULL)`).Scan(&n1, &n2))
		if n1 != 12 || n2 != 46 {
			t.Fatalf("counts %d/%d", n1, n2)
		}
		t.Log("two imports, source counts 12/46, importer unit tests verify SE.12 and unknowns")
	})
	t.Run("CMDB_original_rows_and_readonly_review", func(t *testing.T) {
		var beforeCI, beforeRelations int
		if e := p.QueryRow(ctx, `SELECT (SELECT count(*) FROM configuration_items),(SELECT count(*) FROM ci_relations)`).Scan(&beforeCI, &beforeRelations); e != nil {
			t.Fatal(e)
		}
		if e := s.ImportCMDBSource(ctx, "data/fuentes-service-desk/CMDB.xlsx"); e != nil {
			t.Fatal(e)
		}
		if e := s.ImportCMDBSource(ctx, "data/fuentes-service-desk/CMDB.xlsx"); e != nil {
			t.Fatal(e)
		}
		rows, e := s.CMDBSources(ctx, &admin)
		if e != nil {
			t.Fatal(e)
		}
		if len(rows) != 2 {
			t.Fatalf("source rows %d", len(rows))
		}
		book, e := excelize.OpenFile("data/fuentes-service-desk/CMDB.xlsx")
		if e != nil {
			t.Fatal(e)
		}
		defer book.Close()
		for _, row := range rows {
			raw := row["raw_values"].(map[string]any)
			if len(raw) != 20 {
				t.Fatal("original field lost")
			}
			for cell, value := range raw {
				expected, e := book.GetCellValue("CMDB", cell)
				if e != nil || value != expected {
					t.Fatalf("source mismatch at %s", cell)
				}
			}
		}
		var afterCI, afterRelations int
		if e := p.QueryRow(ctx, `SELECT (SELECT count(*) FROM configuration_items),(SELECT count(*) FROM ci_relations)`).Scan(&afterCI, &afterRelations); e != nil {
			t.Fatal(e)
		}
		if beforeCI != afterCI || beforeRelations != afterRelations {
			t.Fatal("source import created operational CI or relationships")
		}
		if _, e = s.CMDBSources(ctx, &reader); e == nil {
			t.Fatal("consulta accessed administrator source review")
		}
	})
	t.Run("P10_execution_without_approval_and_P15_empty_resolution", func(t *testing.T) {
		var changeQueue, eventQueue int64
		if e := p.QueryRow(ctx, `SELECT id FROM queues WHERE process_code='cambios' AND is_active ORDER BY id LIMIT 1`).Scan(&changeQueue); e != nil {
			t.Fatal(e)
		}
		if e := p.QueryRow(ctx, `SELECT id FROM queues WHERE process_code='eventos' AND is_active ORDER BY id LIMIT 1`).Scan(&eventQueue); e != nil {
			t.Fatal(e)
		}
		change, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: prefix + " approval", Description: "Fixture deliberado sin aprobación", Service: service, Process: "cambios", Queue: changeQueue, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Prueba sin CI", Details: map[string]any{"impact_assessment": "Bajo", "risk": "Bajo", "test_plan": "Verificar", "rollback_plan": "Restaurar", "planned_at": "2026-10-28T10:00:00Z"}})
		if e != nil {
			t.Fatal(e)
		}
		id := int64(change["id"].(float64))
		// Deliberately inconsistent test fixture, never a production transition.
		if _, e = p.Exec(ctx, `UPDATE tickets SET state='planificado' WHERE id=$1`, id); e != nil {
			t.Fatal(e)
		}
		if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: "ejecucion", Version: 1}); e == nil {
			t.Fatal("execution accepted without approval")
		}
		event, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: prefix + " resolution", Description: "Cierre controlado", Service: service, Process: "eventos", Queue: eventQueue, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Prueba sin CI", Details: map[string]any{"category": "E0", "evaluation": "Programado"}})
		if e != nil {
			t.Fatal(e)
		}
		id = int64(event["id"].(float64))
		for _, state := range []string{"clasificado", "evaluado"} {
			event, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: state, Version: int64(event["version"].(float64))})
			if e != nil {
				t.Fatal(e)
			}
		}
		if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: "resuelto", Version: int64(event["version"].(float64))}); e == nil {
			t.Fatal("empty resolution accepted")
		}
		event, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: "resuelto", Resolution: "Evento comprobado", Version: int64(event["version"].(float64))})
		if e != nil {
			t.Fatal(e)
		}
		if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "close", Version: int64(event["version"].(float64))}); e == nil {
			t.Fatal("closure without confirmation accepted")
		}
		event, e = s.Act(ctx, &admin, id, desk.Action{Action: "close", Reason: "Confirmación autorizada", Version: int64(event["version"].(float64))})
		if e != nil {
			t.Fatal(e)
		}
		var actor int64
		if e = p.QueryRow(ctx, `SELECT actor_id FROM ticket_history WHERE ticket_id=$1 AND action='close'`, id).Scan(&actor); e != nil || actor != admin.ID {
			t.Fatal("closure actor missing", e)
		}
	})
	t.Run("FAQ_publication_permissions_versions", func(t *testing.T) {
		in := desk.FAQInput{Slug: strings.ToLower(prefix), Title: prefix, Body: "Contenido de prueba", Active: true}
		article, e := s.SaveFAQ(ctx, &admin, 0, in)
		must(e)
		id := int64(article["id"].(float64))
		rows, e := s.FAQs(ctx, &reader, prefix, "", false, 10, 0)
		must(e)
		if len(rows) != 0 {
			t.Fatal("draft leaked")
		}
		if _, e = s.SaveFAQ(ctx, &reader, 0, in); e == nil {
			t.Fatal("consulta write")
		}
		if _, e = s.FAQs(ctx, &reader, "", "", true, 10, 0); e == nil {
			t.Fatal("consulta management")
		}
		in.Version = 1
		in.Published = true
		_, e = s.SaveFAQ(ctx, &admin, id, in)
		must(e)
		rows, e = s.FAQs(ctx, &reader, prefix, "", false, 10, 0)
		must(e)
		if len(rows) != 1 {
			t.Fatal("published missing")
		}
		if _, e = s.SaveFAQ(ctx, &admin, id, in); e == nil {
			t.Fatal("stale version")
		}
		in.Version = 2
		in.Active = false
		_, e = s.SaveFAQ(ctx, &admin, id, in)
		must(e)
		rows, e = s.FAQs(ctx, &reader, prefix, "", false, 10, 0)
		must(e)
		if len(rows) != 0 {
			t.Fatal("inactive leaked")
		}
		if _, e = p.Exec(ctx, `DELETE FROM faq_history WHERE article_id=$1`, id); e == nil {
			t.Fatal("mutable FAQ history")
		}
	})
	t.Run("P07_order_claim_audit", func(t *testing.T) {
		q, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix, Name: prefix, Process: "servicios", Level: 1, Active: true})
		must(e)
		qid := int64(q["id"].(float64))
		var ids []int64
		for i := 0; i < 3; i++ {
			v, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: prefix, Description: "Orden estable", Service: service, Process: "servicios", Queue: qid, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Prueba sin CI"})
			must(e)
			ids = append(ids, int64(v["id"].(float64)))
		}
		mustExec := func(sql string, args ...any) { _, e := p.Exec(ctx, sql, args...); must(e) }
		mustExec(`UPDATE tickets SET created_at='2026-10-28T10:00:00Z' WHERE id=ANY($1)`, ids)
		rows, e := s.ListTickets(ctx, &admin, map[string]string{"queue": fmt.Sprint(qid)}, 10, 0)
		must(e)
		if len(rows) != 3 {
			t.Fatalf("order filter count %d", len(rows))
		}
		for i, v := range rows {
			if int64(v["id"].(float64)) != ids[i] {
				t.Fatal("unstable order")
			}
		}
		var wg sync.WaitGroup
		start := make(chan struct{})
		out := make(chan error, 2)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				_, e := s.Act(ctx, &admin, ids[0], desk.Action{Action: "claim", Version: 1})
				out <- e
			}()
		}
		close(start)
		wg.Wait()
		close(out)
		successes := 0
		for e := range out {
			if e == nil {
				successes++
			}
		}
		var count int
		must(p.QueryRow(ctx, `SELECT count(*) FROM ticket_history WHERE ticket_id=$1 AND action='claim'`, ids[0]).Scan(&count))
		if successes != 1 || count != 1 {
			t.Fatalf("successes=%d audits=%d", successes, count)
		}
		filters := map[string]string{"q": prefix, "queue": fmt.Sprint(qid), "process": "servicios", "state": "solicitado", "priority": "2", "service": fmt.Sprint(service), "assignee": fmt.Sprint(admin.ID), "from": "2026-10-28T09:59:00Z", "to": "2026-10-28T10:01:00Z"}
		rows, e = s.ListTickets(ctx, &admin, filters, 10, 0)
		must(e)
		if len(rows) != 1 || int64(rows[0]["id"].(float64)) != ids[0] {
			t.Fatal("combined filters", rows)
		}
		rows, e = s.ListTickets(ctx, &admin, map[string]string{"queue": fmt.Sprint(qid)}, 1, 1)
		must(e)
		if len(rows) != 1 || int64(rows[0]["id"].(float64)) != ids[1] {
			t.Fatal("stable pagination")
		}
	})
	t.Run("P05_inactive_references_and_derived_level", func(t *testing.T) {
		var queue, ci, level1 int64
		q, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix + "-refs", Name: "Referencias inactivas", Process: "servicios", Level: 1, Active: true})
		must(e)
		queue = int64(q["id"].(float64))
		_, e = s.Generate(ctx, &admin, "integration", 202300644, 5000)
		must(e)
		must(p.QueryRow(ctx, `SELECT id FROM configuration_items WHERE is_active ORDER BY id LIMIT 1`).Scan(&ci))
		must(p.QueryRow(ctx, `SELECT service_level1_id FROM services_level2 WHERE id=$1`, service).Scan(&level1))
		in := desk.TicketInput{Title: prefix + " refs", Description: "Validación", Service: service, Process: "servicios", Queue: queue, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", CI: []int64{ci}}
		for _, ref := range []struct {
			table string
			id    int64
		}{{"services_level2", service}, {"configuration_items", ci}, {"queues", queue}} {
			_, e = p.Exec(ctx, "UPDATE "+ref.table+" SET is_active=false WHERE id=$1", ref.id)
			must(e)
			_, createErr := s.CreateTicket(ctx, &admin, in)
			_, restoreErr := p.Exec(ctx, "UPDATE "+ref.table+" SET is_active=true WHERE id=$1", ref.id)
			must(restoreErr)
			if createErr == nil {
				t.Fatalf("inactive %s accepted", ref.table)
			}
		}
		ticket, e := s.CreateTicket(ctx, &admin, in)
		must(e)
		ticket, e = s.Ticket(ctx, &admin, int64(ticket["id"].(float64)))
		must(e)
		if ticket["service_level1_id"] != float64(level1) {
			t.Fatal("level1 not derived")
		}
	})
	t.Run("P16_query_plan", func(t *testing.T) {
		_, e := s.Generate(ctx, &admin, "integration", 202300644, 5000)
		must(e)
		var count int
		must(p.QueryRow(ctx, `SELECT count(*) FROM configuration_items WHERE is_device AND is_active`).Scan(&count))
		if count < 5000 {
			t.Fatalf("plan requires 5000 devices, found %d", count)
		}
		t.Logf("dataset integration seed 202300644, active devices %d", count)
		rows, e := p.Query(ctx, `EXPLAIN (ANALYZE,BUFFERS) SELECT id,code,name FROM configuration_items WHERE is_active ORDER BY code,id LIMIT 100 OFFSET 100`)
		must(e)
		defer rows.Close()
		for rows.Next() {
			var line string
			must(rows.Scan(&line))
			t.Log(line)
		}
		must(rows.Err())
	})
	t.Run("P11_concurrent_events", func(t *testing.T) {
		var queue int64
		must(p.QueryRow(ctx, `SELECT id FROM queues WHERE process_code='eventos' AND is_active ORDER BY id LIMIT 1`).Scan(&queue))
		for _, category := range []string{"E2", "E3"} {
			ticket, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: prefix + category, Description: "Carrera E2/E3", Service: service, Process: "eventos", Queue: queue, Priority: 1, Impact: "alto", Urgency: "alto", Channel: "test", NoCI: "Diagnóstico pendiente", Details: map[string]any{"category": category, "evaluation": "Evaluación de prueba"}})
			must(e)
			id := int64(ticket["id"].(float64))
			for _, state := range []string{"clasificado", "evaluado"} {
				ticket, e = s.Act(ctx, &admin, id, desk.Action{Action: "transition", To: state, Version: int64(ticket["version"].(float64))})
				must(e)
			}
			version := int64(ticket["version"].(float64))
			start := make(chan struct{})
			out := make(chan error, 2)
			for i := 0; i < 2; i++ {
				go func() {
					<-start
					_, e := s.Act(ctx, &admin, id, desk.Action{Action: "process_event", Version: version})
					out <- e
				}()
			}
			close(start)
			wins := 0
			for i := 0; i < 2; i++ {
				if <-out == nil {
					wins++
				}
			}
			var links, notifications int
			must(p.QueryRow(ctx, `SELECT (SELECT count(*) FROM event_incidents WHERE event_ticket_id=$1),(SELECT count(*) FROM notifications WHERE dedup_key='event-'||$1::text)`, id).Scan(&links, &notifications))
			if wins != 1 || links != 1 || notifications != 1 {
				t.Fatalf("%s wins=%d links=%d notifications=%d", category, wins, links, notifications)
			}
		}
	})
	t.Run("P04_concurrent_class_cycle", func(t *testing.T) {
		first, e := s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-race-a", Name: "A", IsActive: true})
		must(e)
		second, e := s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-race-b", Name: "B", IsActive: true})
		must(e)
		one, two := int64(first["id"].(float64)), int64(second["id"].(float64))
		start := make(chan struct{})
		out := make(chan error, 2)
		go func() {
			<-start
			_, e := s.SaveClass(ctx, &admin, one, desk.ClassInput{Code: prefix + "-race-a", Name: "A", ParentID: &two, IsActive: true})
			out <- e
		}()
		go func() {
			<-start
			_, e := s.SaveClass(ctx, &admin, two, desk.ClassInput{Code: prefix + "-race-b", Name: "B", ParentID: &one, IsActive: true})
			out <- e
		}()
		close(start)
		wins := 0
		for i := 0; i < 2; i++ {
			if <-out == nil {
				wins++
			}
		}
		if wins != 1 {
			t.Fatalf("cycle race successes=%d", wins)
		}
		t.Log("cycle integrity verified, exact classroom taxonomy remains source-limited")
	})
	t.Run("P14_impact_service_and_depth", func(t *testing.T) {
		items, e := s.ListCI(ctx, "", 0, "", 100, 0)
		must(e)
		if len(items) < 34 {
			t.Fatal("requires generated CIs")
		}
		class, e := s.SaveClass(ctx, &admin, 0, desk.ClassInput{Code: prefix + "-impact", Name: "Fixture de impacto no dispositivo", IsActive: true})
		must(e)
		ids := make([]int64, 34)
		for i := range ids {
			ci, e := s.SaveCI(ctx, &admin, 0, desk.CIInput{Code: fmt.Sprintf("%s-impact-%d", prefix, i), Name: "Fixture de impacto", ClassID: int64(class["id"].(float64)), LocationID: int64(items[0]["location_id"].(float64)), OwnerID: admin.ID, Status: "activo", IsActive: true})
			must(e)
			ids[i] = int64(ci["id"].(float64))
		}
		for i := 0; i < 33; i++ {
			must(s.Relation(ctx, &admin, 0, ids[i], ids[i+1], "depende_de", false))
		}
		must(s.LinkService(ctx, &admin, ids[2], service))
		impact, e := s.Impact(ctx, ids[33])
		must(e)
		found, limit := false, false
		for _, row := range impact {
			if row["service_id"] == float64(service) {
				found = true
			}
			if row["depth_limit_reached"] == true {
				limit = true
			}
		}
		if !found || !limit {
			t.Fatalf("service=%v depth_flag=%v", found, limit)
		}
		if e = s.Relation(ctx, &admin, 0, ids[33], ids[0], "depende_de", false); e == nil {
			t.Fatal("indirect dependency cycle accepted")
		}
		must(s.Relation(ctx, &admin, 0, ids[0], ids[1], "conectado_a", false))
		must(s.Relation(ctx, &admin, 0, ids[1], ids[2], "conectado_a", false))
		must(s.Relation(ctx, &admin, 0, ids[2], ids[0], "conectado_a", false))
	})
	t.Run("P15_audit_failure_rolls_back", func(t *testing.T) {
		var queue int64
		must(p.QueryRow(ctx, `SELECT id FROM queues WHERE process_code='eventos' AND is_active ORDER BY id LIMIT 1`).Scan(&queue))
		ticket, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: prefix + " rollback", Description: "Atomicidad", Service: service, Process: "eventos", Queue: queue, Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Prueba sin CI", Details: map[string]any{"category": "E0"}})
		must(e)
		id := int64(ticket["id"].(float64))
		fn := fmt.Sprintf("final_audit_%d", id)
		_, e = p.Exec(ctx, fmt.Sprintf(`CREATE FUNCTION %s() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN IF NEW.ticket_id=%d THEN RAISE EXCEPTION 'synthetic audit failure'; END IF; RETURN NEW; END $$`, fn, id))
		must(e)
		defer p.Exec(context.Background(), "DROP FUNCTION "+fn+"() CASCADE")
		_, e = p.Exec(ctx, fmt.Sprintf(`CREATE TRIGGER %s BEFORE INSERT ON ticket_history FOR EACH ROW EXECUTE FUNCTION %s()`, fn, fn))
		must(e)
		if _, e = s.Act(ctx, &admin, id, desk.Action{Action: "claim", Version: 1}); e == nil {
			t.Fatal("write without audit succeeded")
		}
		current, e := s.Ticket(ctx, &admin, id)
		must(e)
		if current["version"] != float64(1) || current["assignee_id"] != nil {
			t.Fatal("write was not rolled back")
		}
		if _, e = p.Exec(ctx, `DELETE FROM ticket_history WHERE ticket_id=$1`, id); e == nil {
			t.Fatal("history deletion accepted")
		}
	})
}
