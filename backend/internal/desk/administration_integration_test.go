package desk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/httpapi"
)

func TestAdministrationRegressions(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated PostgreSQL required")
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
	var position, service int64
	must(p.QueryRow(ctx, `SELECT id,position_id FROM app_users WHERE username='admin-demo'`).Scan(&admin.ID, &position))
	admin.Role = "admin"
	admin.Roles = []string{"admin"}
	must(p.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY id LIMIT 1`).Scan(&service))
	s := desk.NewService(p)
	a := auth.NewService(p, time.Hour)
	prefix := fmt.Sprintf("AUD-%d", time.Now().UnixNano())
	qin := desk.QueueInput{Code: prefix, Name: "Cola auditada", Process: "eventos", Level: 1, Active: true}
	queue, e := s.SaveQueue(ctx, &admin, 0, qin)
	must(e)
	qid := int64(queue["id"].(float64))
	other, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix + "-other", Name: "Otra cola", Process: "eventos", Level: 1, Active: true})
	must(e)
	otherID := int64(other["id"].(float64))
	agent, e := a.CreateUser(ctx, position, "Administrador delegado", prefix, prefix+"@example.local", "TemporaryAudit123!", "agente")
	must(e)
	agent.Roles = []string{"agente"}
	reader, e := a.CreateUser(ctx, position, "Consulta auditada", prefix+"-read", prefix+"-read@example.local", "TemporaryAudit123!", "consulta")
	must(e)
	reader.Roles = []string{"consulta"}
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{UserID: &agent.ID, Admin: true}))
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{UserID: &reader.ID, Admin: true}))

	server := httptest.NewServer(httpapi.NewServer(p).Handler())
	defer server.Close()
	session, e := a.Login(ctx, agent.Username, "TemporaryAudit123!")
	must(e)
	call := func(token, method, path string, body any, want int) {
		t.Helper()
		payload, e := json.Marshal(body)
		must(e)
		r, e := http.NewRequest(method, server.URL+path, bytes.NewReader(payload))
		must(e)
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		response, e := http.DefaultClient.Do(r)
		must(e)
		defer response.Body.Close()
		if response.StatusCode != want {
			t.Fatalf("%s %s: status %d, want %d", method, path, response.StatusCode, want)
		}
	}
	path := fmt.Sprintf("/api/desk/queues/%d", qid)
	call(session.Token, "GET", path+"/grants", nil, 200)
	qin.Name = "Edición delegada"
	call(session.Token, "PUT", path, qin, 200)
	role := "aprobador"
	call(session.Token, "PUT", path+"/grants", desk.GrantInput{Role: &role, Read: true, Approve: true}, 204)
	call(session.Token, "PUT", fmt.Sprintf("/api/desk/queues/%d", otherID), qin, 403)
	call(session.Token, "GET", fmt.Sprintf("/api/desk/queues/%d/grants", otherID), nil, 403)
	call(session.Token, "POST", "/api/desk/queues", qin, 403)
	call(session.Token, "PUT", fmt.Sprintf("/api/users/%d/roles", agent.ID), map[string]any{"roles": []string{"admin"}}, 403)
	call(session.Token, "GET", "/api/desk/admin-history", nil, 403)
	readSession, e := a.Login(ctx, reader.Username, "TemporaryAudit123!")
	must(e)
	call(readSession.Token, "PUT", path, qin, 403)
	call(readSession.Token, "PUT", path+"/grants", desk.GrantInput{Role: &role, Admin: true}, 403)
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{UserID: &agent.ID}))
	call(session.Token, "PUT", path, qin, 403)
	agentRole := "agente"
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{Role: &agentRole, Admin: true}))
	call(session.Token, "PUT", path, qin, 200)
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{Role: &agentRole}))
	call(session.Token, "PUT", path, qin, 403)
	var delegatedChanges int
	must(p.QueryRow(ctx, `SELECT count(*) FROM administration_history WHERE entity_table='queues' AND actor_id=$1`, agent.ID).Scan(&delegatedChanges))
	if delegatedChanges == 0 {
		t.Fatal("delegated actor missing from audit")
	}
	t.Log("PASS delegated queue edit/grants, cross-queue and global denials, consulta denial, live revocation")

	var class int64
	var code, name string
	var parent *int64
	must(p.QueryRow(ctx, `SELECT id,code,name,parent_id FROM ci_classes WHERE technical_profile='servidor'`).Scan(&class, &code, &name, &parent))
	original := desk.ClassInput{Code: code, Name: name, ParentID: parent, IsDevice: true, IsActive: true}
	renamed := original
	renamed.Code = prefix + "-server"
	_, e = s.SaveClass(ctx, &admin, class, renamed)
	must(e)
	defer func() { _, e := s.SaveClass(ctx, &admin, class, original); must(e) }()
	location, e := s.SaveLocation(ctx, &admin, 0, prefix, "Laboratorio auditoría", true)
	must(e)
	input := desk.CIInput{Code: prefix + "-ci", Name: "Servidor auditado", ClassID: class, LocationID: int64(location["id"].(float64)), OwnerID: admin.ID, Status: "activo", Technical: map[string]any{"ram_gb": float64(16)}, IsDevice: true, IsActive: true}
	ci, e := s.SaveCI(ctx, &admin, 0, input)
	must(e)
	ciID := int64(ci["id"].(float64))
	input.Name = "Servidor editado"
	_, e = s.SaveCI(ctx, &admin, ciID, input)
	must(e)
	report, e := s.Generate(ctx, &admin, "audit-regression", 202300644, 5000)
	must(e)
	if report["devices"] != 5000 {
		t.Fatal(report)
	}
	report, e = s.Generate(ctx, &admin, "audit-regression", 202300644, 5000)
	must(e)
	if report["created"] != 0 {
		t.Fatal("generation duplicated", report)
	}
	if _, e = p.Exec(ctx, `UPDATE ci_classes SET technical_profile='usuario' WHERE id=$1`, class); e == nil {
		t.Fatal("mutable technical profile")
	}
	t.Log("PASS renamed class: CI create/update and repeated 5000-device generation")

	input.Code = prefix + "-target"
	target, e := s.SaveCI(ctx, &admin, 0, input)
	must(e)
	targetID := int64(target["id"].(float64))
	must(s.Relation(ctx, &admin, 0, ciID, targetID, "depende_de", false))
	var relation int64
	must(p.QueryRow(ctx, `SELECT id FROM ci_relations WHERE source_ci_id=$1 AND target_ci_id=$2`, ciID, targetID).Scan(&relation))
	must(s.Relation(ctx, &admin, relation, ciID, targetID, "conectado_a", false))
	must(s.Relation(ctx, &admin, relation, 0, 0, "", true))
	must(s.LinkService(ctx, &admin, ciID, service))
	must(a.SetRolesBy(ctx, admin.ID, agent.ID, []string{"agente", "aprobador"}))
	for _, entity := range []string{"ci_classes", "queues", "queue_grants", "locations", "ci_relations", "ci_services", "user_roles"} {
		var count int
		must(p.QueryRow(ctx, `SELECT count(*) FROM administration_history WHERE entity_table=$1 AND actor_id=$2 AND origin='application'`, entity, admin.ID).Scan(&count))
		if count == 0 {
			t.Fatal("missing audit", entity)
		}
	}
	var removed bool
	must(p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM administration_history WHERE entity_table='ci_relations' AND action='DELETE' AND (before_value->>'id')::bigint=$1 AND after_value IS NULL AND actor_id=$2)`, relation, admin.ID).Scan(&removed))
	if !removed {
		t.Fatal("deleted relationship evidence missing")
	}
	var linkedHistory int
	must(p.QueryRow(ctx, `SELECT count(*) FROM ci_changes WHERE ci_id=$1 AND description LIKE 'ci_relations:%' AND actor_id=$2`, ciID, admin.ID).Scan(&linkedHistory))
	if linkedHistory != 3 {
		t.Fatal("relation history not visible on CI", linkedHistory)
	}
	history, e := s.AdministrationHistory(ctx, &admin, 100, 0)
	must(e)
	if len(history) == 0 {
		t.Fatal("empty history API")
	}
	if _, e = s.AdministrationHistory(ctx, &reader, 10, 0); e == nil {
		t.Fatal("consulta reads admin history")
	}
	for _, sql := range []string{`UPDATE administration_history SET action='alterado'`, `DELETE FROM administration_history`, `TRUNCATE administration_history`} {
		if _, e = p.Exec(ctx, sql); e == nil {
			t.Fatal("audit tampering permitted", sql)
		}
	}
	t.Log("PASS administrative snapshots, actor/date, relation deletion on both CI histories, immutable audit")
}
