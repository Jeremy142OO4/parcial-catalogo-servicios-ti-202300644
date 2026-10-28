package desk_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
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

func TestProtectedAttachmentsIntegration(t *testing.T) {
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("isolated database required")
	}
	parsed, e := url.Parse(dsn)
	if e != nil || !strings.HasSuffix(parsed.Path, "_desk_test") {
		t.Fatal("isolated database required")
	}
	if os.Getenv("ATTACHMENTS_DIR") == "" {
		t.Setenv("ATTACHMENTS_DIR", t.TempDir())
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
	s := desk.NewService(p)
	a := auth.NewService(p, time.Hour)
	var admin auth.User
	var position, service int64
	must(p.QueryRow(ctx, `SELECT id,position_id FROM app_users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&admin.ID, &position))
	admin.Roles = []string{"admin"}
	must(p.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY id LIMIT 1`).Scan(&service))
	prefix := fmt.Sprintf("ATT-%d", time.Now().UnixNano())
	newUser := func(role string) auth.User {
		u, e := a.CreateUser(ctx, position, "Usuario prueba adjuntos", prefix+role, prefix+role+"@example.local", "TemporalDemo123!", role)
		must(e)
		u.Roles = []string{role}
		return u
	}
	requester := newUser("solicitante")
	stranger := newUser("agente")
	consulta := newUser("consulta")
	queue, e := s.SaveQueue(ctx, &admin, 0, desk.QueueInput{Code: prefix, Name: "Prueba adjuntos", Process: "eventos", Level: 1, Active: true})
	must(e)
	ticket, e := s.CreateTicket(ctx, &admin, desk.TicketInput{Title: "Adjuntos sintéticos", Description: "Prueba aislada", Requester: requester.ID, Service: service, Process: "eventos", Queue: int64(queue["id"].(float64)), Priority: 2, Impact: "medio", Urgency: "medio", Channel: "test", NoCI: "Prueba sin dispositivo", Details: map[string]any{"category": "E0"}})
	must(e)
	id := int64(ticket["id"].(float64))
	base := fmt.Sprintf("/api/desk/tickets/%d/attachments", id)
	server := httptest.NewServer(httpapi.NewServer(p).Handler())
	defer server.Close()
	login := func(u auth.User) string {
		session, e := a.Login(ctx, u.Username, "TemporalDemo123!")
		must(e)
		return session.Token
	}
	reqToken, otherToken, readToken := login(requester), login(stranger), login(consulta)
	call := func(method, path, token, name, visibility string, body []byte) *http.Response {
		var input io.Reader
		ct := ""
		if method == "POST" {
			buf := new(bytes.Buffer)
			writer := multipart.NewWriter(buf)
			part, e := writer.CreateFormFile("file", name)
			must(e)
			_, e = part.Write(body)
			must(e)
			must(writer.WriteField("visibility", visibility))
			must(writer.Close())
			input = buf
			ct = writer.FormDataContentType()
		}
		r, e := http.NewRequest(method, server.URL+path, input)
		must(e)
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		r.Header.Set("Content-Type", ct)
		res, e := http.DefaultClient.Do(r)
		must(e)
		return res
	}
	status := func(res *http.Response, want int) {
		t.Helper()
		defer res.Body.Close()
		if res.StatusCode != want {
			b, _ := io.ReadAll(res.Body)
			t.Fatalf("want %d got %d: %s", want, res.StatusCode, b)
		}
	}
	status(call("POST", base, "", "nota.txt", "public", []byte("prueba")), 401)
	status(call("POST", base, otherToken, "nota.txt", "public", []byte("prueba")), 403)
	status(call("POST", base, readToken, "nota.txt", "public", []byte("prueba")), 403)
	status(call("POST", base, reqToken, "nota.txt", "internal", []byte("prueba")), 403)
	t.Log("A01 sin sesión, tercero, consulta y nota interna sin permiso rechazados")
	entries := func() int {
		items, e := os.ReadDir(os.Getenv("ATTACHMENTS_DIR"))
		if os.IsNotExist(e) {
			return 0
		}
		must(e)
		return len(items)
	}
	before := entries()
	for _, tc := range []struct {
		name string
		data []byte
		code int
	}{
		{"falso.pdf", []byte("texto"), 415}, {"html.txt", []byte("<!DOCTYPE html><html>no</html>"), 415},
		{"../nota.txt", []byte("texto"), 400}, {"vacio.txt", nil, 400},
		{"grande.txt", bytes.Repeat([]byte("a"), int(desk.MaxAttachmentSize)+1), 413},
		{"enorme.txt", bytes.Repeat([]byte("a"), 12<<20), 413},
	} {
		status(call("POST", base, reqToken, tc.name, "public", tc.data), tc.code)
	}
	if entries() != before {
		t.Fatal("invalid upload left file behind")
	}
	t.Log("A02 tipos falsos, rutas, vacíos y límites rechazados sin archivos residuales")
	var uploaded desk.Attachment
	res := call("POST", base, reqToken, "evidencia.txt", "public", []byte("evidencia de adjunto protegido"))
	if res.StatusCode != 201 {
		status(res, 201)
	}
	must(json.NewDecoder(res.Body).Decode(&uploaded))
	res.Body.Close()
	download := fmt.Sprintf("%s/%d/download", base, uploaded.ID)
	res = call("GET", download, reqToken, "", "", nil)
	data, e := io.ReadAll(res.Body)
	must(e)
	res.Body.Close()
	if res.StatusCode != 200 || string(data) != "evidencia de adjunto protegido" || res.Header.Get("X-Content-Type-Options") != "nosniff" || !strings.HasPrefix(res.Header.Get("Content-Disposition"), "attachment;") {
		t.Fatal("download mismatch")
	}
	status(call("GET", download, otherToken, "", "", nil), 403)
	status(call("GET", fmt.Sprintf("/api/desk/tickets/%d/attachments/%d/download", id+999999, uploaded.ID), reqToken, "", "", nil), 404)
	status(call("GET", download, readToken, "", "", nil), 200)
	qid := int64(queue["id"].(float64))
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{UserID: &stranger.ID, Read: true, Handle: true}))
	res = call("POST", base, otherToken, "interno.txt", "internal", []byte("contenido interno"))
	if res.StatusCode != 201 {
		status(res, 201)
	}
	var internal desk.Attachment
	must(json.NewDecoder(res.Body).Decode(&internal))
	res.Body.Close()
	status(call("GET", fmt.Sprintf("%s/%d/download", base, internal.ID), reqToken, "", "", nil), 403)
	res = call("GET", base, reqToken, "", "", nil)
	data, e = io.ReadAll(res.Body)
	must(e)
	res.Body.Close()
	if bytes.Contains(data, []byte("storage_key")) || bytes.Contains(data, []byte("interno.txt")) {
		t.Fatal("private metadata leaked")
	}
	must(s.Grant(ctx, &admin, qid, desk.GrantInput{UserID: &stranger.ID, Read: false, Handle: false}))
	status(call("GET", download, otherToken, "", "", nil), 403)
	t.Log("A03 descarga fiel, headers seguros, aislamiento por ticket, visibilidad interna y revocación comprobados")
	res = call("POST", base, reqToken, "limite.txt", "public", bytes.Repeat([]byte("a"), int(desk.MaxAttachmentSize)))
	status(res, 201)
	t.Log("A04 tamaño exacto de 10 MiB aceptado")
	t.Logf("Adjunto persistente de prueba ticket_id=%d attachment_id=%d", id, uploaded.ID)
}

// Run explicitly before and after restarting the isolated API and database.
func TestAttachmentPersistenceCheck(t *testing.T) {
	password := os.Getenv("ATTACHMENT_PERSISTENCE_PASSWORD")
	dsn := os.Getenv("DESK_TEST_DATABASE_URL")
	if password == "" || dsn == "" {
		t.Skip("explicit persistence check only")
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
	var username string
	if e = p.QueryRow(ctx, `SELECT username FROM app_users WHERE role='admin' ORDER BY id LIMIT 1`).Scan(&username); e != nil {
		t.Fatal(e)
	}
	session, e := auth.NewService(p, time.Hour).Login(ctx, username, password)
	if e != nil {
		t.Fatal(e)
	}
	var ticket, id int64
	if e = p.QueryRow(ctx, `SELECT ticket_id,id FROM ticket_attachments WHERE original_name='evidencia.txt' ORDER BY id DESC LIMIT 1`).Scan(&ticket, &id); e != nil {
		t.Fatal(e)
	}
	r, e := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/desk/tickets/%d/attachments/%d/download", ticket, id), nil)
	if e != nil {
		t.Fatal(e)
	}
	r.Header.Set("Authorization", "Bearer "+session.Token)
	res, e := http.DefaultClient.Do(r)
	if e != nil {
		t.Fatal(e)
	}
	defer res.Body.Close()
	data, e := io.ReadAll(res.Body)
	if e != nil || res.StatusCode != 200 || string(data) != "evidencia de adjunto protegido" {
		t.Fatalf("persistence download status=%d error=%v", res.StatusCode, e)
	}
	t.Logf("PERSISTENCIA OK ticket_id=%d attachment_id=%d bytes=%d", ticket, id, len(data))
}
