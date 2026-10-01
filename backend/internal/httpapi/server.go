package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/catalog"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/organization"
)

type Server struct {
	pool    *pgxpool.Pool
	auth    *auth.Service
	org     *organization.Repository
	catalog *catalog.Repository
}

func NewServer(pool *pgxpool.Pool) *Server {
	return &Server{
		pool:    pool,
		auth:    auth.NewService(pool, 8*time.Hour),
		org:     organization.NewRepository(pool),
		catalog: catalog.NewRepository(pool),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.health)
	mux.HandleFunc("GET /api/ready", s.ready)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/logout", s.requireAuth(s.logout))
	mux.HandleFunc("GET /api/me", s.requireAuth(s.me))
	mux.HandleFunc("GET /api/organization/units", s.requireAuth(s.listUnits))
	mux.HandleFunc("POST /api/organization/units", s.requireAdmin(s.createUnit))
	mux.HandleFunc("PATCH /api/organization/units/{type}/{id}", s.requireAdmin(s.updateUnit))
	mux.HandleFunc("PATCH /api/organization/units/{type}/{id}/active", s.requireAdmin(s.setUnitActive))
	mux.HandleFunc("GET /api/users", s.requireAuth(s.listUsers))
	mux.HandleFunc("POST /api/users", s.requireAdmin(s.createUser))
	mux.HandleFunc("PATCH /api/users/{id}", s.requireAdmin(s.updateUser))
	mux.HandleFunc("PATCH /api/users/{id}/active", s.requireAdmin(s.setUserActive))
	mux.HandleFunc("GET /api/catalog/services", s.requireAuth(s.listServices))
	mux.HandleFunc("POST /api/catalog/assignments", s.requireAdmin(s.assignService))
	return cors(mux)
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		writeError(w, http.StatusServiceUnavailable, "base de datos no disponible")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

type loginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	var input loginRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	session, err := s.auth.Login(r.Context(), input.Identifier, input.Password)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "credenciales inválidas")
		return
	}
	writeJSON(w, http.StatusOK, session)
}

func (s *Server) logout(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	if err := s.auth.Logout(r.Context(), bearerToken(r)); err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo cerrar la sesión")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) me(w http.ResponseWriter, _ *http.Request, user *auth.User) {
	writeJSON(w, http.StatusOK, user)
}

type unitRequest struct {
	Type     string `json:"type"`
	ParentID *int64 `json:"parent_id"`
	Code     string `json:"code"`
	Name     string `json:"name"`
}

func (s *Server) listUnits(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	units, err := s.org.ListUnits(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudo consultar la organización")
		return
	}
	writeJSON(w, http.StatusOK, units)
}

func (s *Server) createUnit(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	var input unitRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	unit, err := s.org.CreateUnit(r.Context(), input.Type, input.ParentID, input.Code, input.Name)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, unit)
}

func (s *Server) setUnitActive(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "identificador inválido")
		return
	}
	var input struct {
		IsActive bool `json:"is_active"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.org.SetActive(r.Context(), r.PathValue("type"), id, input.IsActive); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateUnit(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "identificador inválido")
		return
	}
	var input struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	unit, err := s.org.UpdateUnit(r.Context(), r.PathValue("type"), id, input.Code, input.Name)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, unit)
}

func (s *Server) listUsers(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	users, err := s.auth.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudieron consultar los usuarios")
		return
	}
	writeJSON(w, http.StatusOK, users)
}

type userRequest struct {
	PositionID int64  `json:"position_id"`
	FullName   string `json:"full_name"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
	Role       string `json:"role"`
}

func (s *Server) createUser(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	var input userRequest
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.Role != "admin" && input.Role != "consulta" {
		writeError(w, http.StatusBadRequest, "el rol debe ser admin o consulta")
		return
	}
	user, err := s.auth.CreateUser(r.Context(), input.PositionID, input.FullName, input.Username, input.Email, input.Password, input.Role)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, user)
}

func (s *Server) setUserActive(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "identificador inválido")
		return
	}
	var input struct {
		IsActive bool `json:"is_active"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	if err := s.auth.SetUserActive(r.Context(), id, input.IsActive); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) updateUser(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "identificador inválido")
		return
	}
	var input struct {
		PositionID *int64  `json:"position_id"`
		FullName   *string `json:"full_name"`
		Email      *string `json:"email"`
		Role       *string `json:"role"`
	}
	if !decodeJSON(w, r, &input) {
		return
	}
	user, err := s.auth.UpdateUser(r.Context(), id, input.PositionID, input.FullName, input.Email, input.Role)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (s *Server) listServices(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	services, err := s.catalog.ListServices(r.Context(), r.URL.Query().Get("q"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "no se pudieron consultar los servicios")
		return
	}
	writeJSON(w, http.StatusOK, services)
}

func (s *Server) assignService(w http.ResponseWriter, r *http.Request, _ *auth.User) {
	var input catalog.AssignmentInput
	if !decodeJSON(w, r, &input) {
		return
	}
	if input.ServiceID <= 0 || input.SectionID <= 0 {
		writeError(w, http.StatusBadRequest, "servicio y sección son obligatorios")
		return
	}
	if err := s.catalog.Assign(r.Context(), input); err != nil {
		writeDomainError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type protectedHandler func(http.ResponseWriter, *http.Request, *auth.User)

func (s *Server) requireAuth(next protectedHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := s.auth.Authenticate(r.Context(), bearerToken(r))
		if err != nil {
			writeError(w, http.StatusUnauthorized, "sesión requerida")
			return
		}
		next(w, r, user)
	}
}

func (s *Server) requireAdmin(next protectedHandler) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request, user *auth.User) {
		if user.Role != "admin" {
			writeError(w, http.StatusForbidden, "se requiere rol administrador")
			return
		}
		next(w, r, user)
	})
}

func bearerToken(r *http.Request) string {
	value := strings.TrimSpace(r.Header.Get("Authorization"))
	if strings.HasPrefix(strings.ToLower(value), "bearer ") {
		return strings.TrimSpace(value[7:])
	}
	return ""
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target any) bool {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		writeError(w, http.StatusBadRequest, "JSON inválido")
		return false
	}
	return true
}

func writeDomainError(w http.ResponseWriter, err error) {
	status := http.StatusBadRequest
	message := err.Error()
	if strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint") {
		status = http.StatusConflict
		message = "ya existe un registro con esos datos"
	}
	writeError(w, status, message)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
