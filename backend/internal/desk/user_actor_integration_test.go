package desk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/httpapi"
)

func TestLegacyUserActorAndPermissions(t *testing.T) {
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
	var root, position int64
	must(p.QueryRow(ctx, `SELECT id,position_id FROM app_users WHERE username='admin-demo'`).Scan(&root, &position))
	a := auth.NewService(p, time.Hour)
	prefix := fmt.Sprintf("USER-AUD-%d", time.Now().UnixNano())
	actor, e := a.CreateUserBy(ctx, root, position, "Administrador prueba", prefix, prefix+"@example.local", "TemporaryUsers123!", "admin")
	must(e)
	session, e := a.Login(ctx, actor.Username, "TemporaryUsers123!")
	must(e)
	server := httptest.NewServer(httpapi.NewServer(p).Handler())
	defer server.Close()
	call := func(token, method, path string, body any, want int) []byte {
		t.Helper()
		raw, e := json.Marshal(body)
		must(e)
		r, e := http.NewRequest(method, server.URL+path, bytes.NewReader(raw))
		must(e)
		r.Header.Set("Authorization", "Bearer "+token)
		r.Header.Set("Content-Type", "application/json")
		response, e := http.DefaultClient.Do(r)
		must(e)
		defer response.Body.Close()
		data, e := io.ReadAll(response.Body)
		must(e)
		if response.StatusCode != want {
			t.Fatalf("%s %s status=%d want=%d response=%s", method, path, response.StatusCode, want, data)
		}
		return data
	}
	body := map[string]any{"position_id": position, "full_name": "Usuario auditado", "username": prefix + "-target", "email": prefix + "-target@example.local", "password": "TemporaryUsers123!", "role": "agente", "actor_id": root}
	data := call(session.Token, "POST", "/api/users", body, 201)
	var target auth.User
	must(json.Unmarshal(data, &target))
	path := fmt.Sprintf("/api/users/%d", target.ID)
	call(session.Token, "PATCH", path, map[string]any{"role": "aprobador", "actor_id": root}, 200)
	call(session.Token, "PATCH", path, map[string]any{"full_name": "Nombre actualizado"}, 200)
	call(session.Token, "PATCH", path+"/active", map[string]any{"is_active": false}, 204)
	call(session.Token, "PATCH", path+"/active", map[string]any{"is_active": true}, 204)
	var userEvents, roleEvents int
	must(p.QueryRow(ctx, `SELECT count(*) FROM administration_history WHERE entity_table='app_users' AND actor_id=$1 AND origin='application' AND coalesce(after_value,before_value)->>'id'=$2`, actor.ID, fmt.Sprint(target.ID)).Scan(&userEvents))
	must(p.QueryRow(ctx, `SELECT count(*) FROM administration_history WHERE entity_table='user_roles' AND actor_id=$1 AND origin='application' AND coalesce(after_value,before_value)->>'user_id'=$2`, actor.ID, fmt.Sprint(target.ID)).Scan(&roleEvents))
	if userEvents != 5 || roleEvents != 3 {
		t.Fatal("missing real HTTP actor", userEvents, roleEvents)
	}
	var unsafe bool
	must(p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM administration_history WHERE entity_table='app_users' AND (before_value ? 'password_hash' OR after_value ? 'password_hash'))`).Scan(&unsafe))
	if unsafe {
		t.Fatal("password hash leaked into history")
	}
	before := userEvents
	call(session.Token, "PATCH", path, map[string]any{"role": "invalid"}, 400)
	must(p.QueryRow(ctx, `SELECT count(*) FROM administration_history WHERE entity_table='app_users' AND actor_id=$1 AND coalesce(after_value,before_value)->>'id'=$2`, actor.ID, fmt.Sprint(target.ID)).Scan(&userEvents))
	if userEvents != before {
		t.Fatal("rejected update changed audit")
	}
	for _, roles := range [][]string{{"agente"}, {"admin", "agente", "consulta"}} {
		must(a.SetRolesBy(ctx, root, target.ID, roles))
		login, e := a.Login(ctx, target.Username, "TemporaryUsers123!")
		must(e)
		call(login.Token, "POST", "/api/users", body, 403)
		call(login.Token, "PATCH", path, map[string]any{"role": "admin"}, 403)
		call(login.Token, "PATCH", path+"/active", map[string]any{"is_active": false}, 403)
		if _, e = a.UpdateUserBy(ctx, target.ID, target.ID, nil, nil, nil, nil); e == nil {
			t.Fatal("service authorized forbidden actor", roles)
		}
	}
	// Controlled revocation holds the user row while creation is already waiting.
	config, e := pgxpool.ParseConfig(dsn)
	must(e)
	config.ConnConfig.RuntimeParams["application_name"] = prefix + "-waiting"
	waitPool, e := pgxpool.NewWithConfig(ctx, config)
	must(e)
	defer waitPool.Close()
	blocker, e := p.Begin(ctx)
	must(e)
	defer blocker.Rollback(context.Background())
	_, e = blocker.Exec(ctx, `UPDATE app_users SET role='agente' WHERE id=$1`, actor.ID)
	must(e)
	result := make(chan error, 1)
	waitCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	go func() {
		_, e := auth.NewService(waitPool, time.Hour).CreateUserBy(waitCtx, actor.ID, position, "No", prefix+"-denied", prefix+"-denied@example.local", "TemporaryUsers123!", "admin")
		result <- e
	}()
	blocked := false
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		must(p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pg_stat_activity WHERE application_name=$1 AND wait_event_type='Lock')`, prefix+"-waiting").Scan(&blocked))
		if blocked {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !blocked {
		t.Fatal("administrator mutation did not wait on revoked user row")
	}
	must(blocker.Commit(ctx))
	if e = <-result; e == nil {
		t.Fatal("concurrently revoked administrator created user")
	}
	var created bool
	must(p.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app_users WHERE username=$1)`, prefix+"-denied").Scan(&created))
	if created {
		t.Fatal("revoked creation persisted")
	}
	call(session.Token, "PATCH", path, map[string]any{"role": "admin"}, 403)
	t.Log("PASS legacy HTTP create/update/activation: authenticated actor, no password hash, rejected updates atomic")
	t.Log("PASS agent, mixed consulta/admin and revoked administrator cannot mutate users")
}
