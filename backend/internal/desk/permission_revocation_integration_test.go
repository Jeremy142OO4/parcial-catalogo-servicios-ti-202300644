package desk_test

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
)

func TestConcurrentPermissionRevocation(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated PostgreSQL required")
	}
	parsed, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(parsed.Path, "_desk_test") {
		t.Fatal("isolated database required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
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
	must(p.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY id LIMIT 1`).Scan(&service))
	a := auth.NewService(p, time.Hour)
	s := desk.NewService(p)
	prefix := fmt.Sprintf("REVOKE-%d", time.Now().UnixNano())
	agent, e := a.CreateUser(ctx, position, "Agente revocación", prefix, prefix+"@example.local", "TemporaryRevoke123!", "agente")
	must(e)
	agent.Roles = []string{"agente"}
	namedPool := func(name string) *pgxpool.Pool {
		c, e := pgxpool.ParseConfig(dsn)
		must(e)
		c.ConnConfig.RuntimeParams["application_name"] = name
		result, e := pgxpool.NewWithConfig(ctx, c)
		must(e)
		return result
	}
	waitForQueueLock := func(name string) {
		t.Helper()
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			var blocked bool
			must(p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock' AND query LIKE '%FROM queues%FOR%')`, name).Scan(&blocked))
			if blocked {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
		t.Fatal("operation did not wait on queue lock", name)
	}
	for _, test := range []struct {
		permission string
		byRole     bool
	}{{"handle", false}, {"approve", false}, {"transfer", false}, {"handle", true}, {"approve", true}, {"transfer", true}} {
		permission := test.permission
		subject := "user"
		if test.byRole {
			subject = "role"
		}
		t.Run(permission+"_"+subject, func(t *testing.T) {
			process := "eventos"
			if permission == "approve" {
				process = "cambios"
			}
			qin := desk.QueueInput{Code: prefix + permission + subject, Name: "Revocación " + permission, Process: process, Level: 1, Active: true}
			q, e := s.SaveQueue(ctx, &admin, 0, qin)
			must(e)
			qid := int64(q["id"].(float64))
			grant := desk.GrantInput{UserID: &agent.ID, Read: true, Handle: permission == "handle", Approve: permission == "approve", Transfer: permission == "transfer"}
			if test.byRole {
				role := "agente"
				grant.UserID = nil
				grant.Role = &role
			}
			must(s.Grant(ctx, &admin, qid, grant))
			details := map[string]any{}
			if process == "eventos" {
				details["category"] = "E0"
			}
			ticket, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: "Revocación concurrente", Description: "Fixture aislado", Service: service, Process: process, Queue: qid, Priority: 2, Impact: "bajo", Urgency: "bajo", Channel: "prueba", NoCI: "Prueba de permisos sin CI", Details: details})
			must(e)
			id := int64(ticket["id"].(float64))
			version := int64(ticket["version"].(float64))
			action := desk.Action{Action: "pause", Reason: "Prueba de exclusión", Version: version}
			if permission == "approve" {
				// Controlled workflow fixture, no self-approval: requester is the administrator.
				_, e = p.Exec(ctx, `UPDATE tickets SET state='pendiente_aprobacion' WHERE id=$1`, id)
				must(e)
				action.Action = "approve"
			}
			if permission == "transfer" {
				qin.Code += "-dest"
				dest, e := s.SaveQueue(ctx, &admin, 0, qin)
				must(e)
				action.Action = "transfer"
				action.Queue = int64(dest["id"].(float64))
				must(s.Grant(ctx, &admin, action.Queue, desk.GrantInput{UserID: &agent.ID, Read: true}))
			}
			var before string
			must(p.QueryRow(ctx, `SELECT row_to_json(t)::text FROM tickets t WHERE id=$1`, id).Scan(&before))
			var beforeHistory int
			must(p.QueryRow(ctx, `SELECT count(*) FROM ticket_history WHERE ticket_id=$1`, id).Scan(&beforeHistory))
			blocker, e := p.Begin(ctx)
			must(e)
			defer blocker.Rollback(context.Background())
			_, e = blocker.Exec(ctx, `SELECT id FROM queues WHERE id=$1 FOR UPDATE`, qid)
			must(e)
			grantName := prefix + permission + subject + "-grant"
			actName := prefix + permission + subject + "-act"
			grantPool := namedPool(grantName)
			defer grantPool.Close()
			actPool := namedPool(actName)
			defer actPool.Close()
			grantResult := make(chan error, 1)
			actResult := make(chan error, 1)
			go func() {
				grantResult <- desk.NewService(grantPool).Grant(ctx, &admin, qid, desk.GrantInput{UserID: grant.UserID, Role: grant.Role, Read: true})
			}()
			waitForQueueLock(grantName)
			go func() { _, e := desk.NewService(actPool).Act(ctx, &agent, id, action); actResult <- e }()
			waitForQueueLock(actName)
			must(blocker.Commit(ctx))
			must(<-grantResult)
			e = <-actResult
			var denied *desk.Error
			if !errors.As(e, &denied) || denied.Status != 403 {
				t.Fatalf("revoked %s action accepted or wrong error: %v", permission, e)
			}
			var after string
			var afterHistory int
			must(p.QueryRow(ctx, `SELECT row_to_json(t)::text FROM tickets t WHERE id=$1`, id).Scan(&after))
			must(p.QueryRow(ctx, `SELECT count(*) FROM ticket_history WHERE ticket_id=$1`, id).Scan(&afterHistory))
			if before != after || beforeHistory != afterHistory {
				t.Fatal("denied action changed ticket or history")
			}
			var pauses, approvals int
			must(p.QueryRow(ctx, `SELECT (SELECT count(*) FROM ticket_pauses WHERE ticket_id=$1),(SELECT count(*) FROM ticket_approvals WHERE ticket_id=$1)`, id).Scan(&pauses, &approvals))
			if pauses != 0 || approvals != 0 {
				t.Fatal("denied action left operational evidence")
			}
			// The same operation succeeds after the precise permission is restored.
			must(s.Grant(ctx, &admin, qid, grant))
			_, e = s.Act(ctx, &agent, id, action)
			must(e)
			t.Log("PASS concurrent revocation: 403, unchanged ticket/history and success after restoring only", permission, "via", subject)
		})
	}
}
