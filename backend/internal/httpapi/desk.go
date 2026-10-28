package httpapi

import (
	"errors"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/desk"
	"net/http"
	"strconv"
)

func deskResult(w http.ResponseWriter, value any, e error) {
	if e != nil {
		var domain *desk.Error
		if errors.As(e, &domain) {
			writeError(w, domain.Status, domain.Message)
		} else {
			writeDomainError(w, e)
		}
		return
	}
	if value == nil {
		w.WriteHeader(204)
	} else {
		writeJSON(w, 200, value)
	}
}
func idOf(r *http.Request) int64             { id, _ := pathID(r); return id }
func number(r *http.Request, key string) int { n, _ := strconv.Atoi(r.URL.Query().Get(key)); return n }
func (s *Server) deskRoutes(mux *http.ServeMux) {
	mux.HandleFunc("PUT /api/users/{id}/roles", s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in struct {
			Roles []string `json:"roles"`
		}
		if decodeJSON(w, r, &in) {
			deskResult(w, nil, s.auth.SetRoles(r.Context(), idOf(r), in.Roles))
		}
	}))
	mux.HandleFunc("GET /api/desk/processes", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Processes(r.Context())
		deskResult(w, v, e)
	}))
	mux.HandleFunc("GET /api/desk/queues", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Queues(r.Context(), u)
		deskResult(w, v, e)
	}))
	for _, pattern := range []string{"POST /api/desk/queues", "PUT /api/desk/queues/{id}"} {
		mux.HandleFunc(pattern, s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
			var in desk.QueueInput
			if decodeJSON(w, r, &in) {
				v, e := s.desk.SaveQueue(r.Context(), u, idOf(r), in)
				deskResult(w, v, e)
			}
		}))
	}
	mux.HandleFunc("PUT /api/desk/queues/{id}/grants", s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in desk.GrantInput
		if decodeJSON(w, r, &in) {
			deskResult(w, nil, s.desk.Grant(r.Context(), u, idOf(r), in))
		}
	}))
	mux.HandleFunc("GET /api/cmdb/classes", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Classes(r.Context())
		deskResult(w, v, e)
	}))
	for _, pattern := range []string{"POST /api/cmdb/classes", "PUT /api/cmdb/classes/{id}"} {
		mux.HandleFunc(pattern, s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
			var in desk.ClassInput
			if decodeJSON(w, r, &in) {
				v, e := s.desk.SaveClass(r.Context(), u, idOf(r), in)
				deskResult(w, v, e)
			}
		}))
	}
	mux.HandleFunc("GET /api/cmdb/items", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		q := r.URL.Query()
		class, _ := strconv.ParseInt(q.Get("class_id"), 10, 64)
		v, e := s.desk.ListCI(r.Context(), q.Get("q"), class, q.Get("status"), number(r, "limit"), number(r, "offset"))
		deskResult(w, v, e)
	}))
	for _, pattern := range []string{"POST /api/cmdb/items", "PUT /api/cmdb/items/{id}"} {
		mux.HandleFunc(pattern, s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
			var in desk.CIInput
			if decodeJSON(w, r, &in) {
				v, e := s.desk.SaveCI(r.Context(), u, idOf(r), in)
				deskResult(w, v, e)
			}
		}))
	}
	mux.HandleFunc("GET /api/cmdb/items/{id}", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.CIDetail(r.Context(), idOf(r))
		deskResult(w, v, e)
	}))
	mux.HandleFunc("GET /api/cmdb/items/{id}/impact", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Impact(r.Context(), idOf(r))
		deskResult(w, v, e)
	}))
	mux.HandleFunc("PUT /api/cmdb/items/{id}/service", s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in struct {
			Service int64 `json:"service_id"`
		}
		if decodeJSON(w, r, &in) {
			deskResult(w, nil, s.desk.LinkService(r.Context(), u, idOf(r), in.Service))
		}
	}))
	for _, pattern := range []string{"POST /api/cmdb/relations", "PUT /api/cmdb/relations/{id}", "DELETE /api/cmdb/relations/{id}"} {
		mux.HandleFunc(pattern, s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
			var in struct {
				From int64  `json:"source_ci_id"`
				To   int64  `json:"target_ci_id"`
				Kind string `json:"kind"`
			}
			if r.Method == "DELETE" || decodeJSON(w, r, &in) {
				deskResult(w, nil, s.desk.Relation(r.Context(), u, idOf(r), in.From, in.To, in.Kind, r.Method == "DELETE"))
			}
		}))
	}
	mux.HandleFunc("GET /api/cmdb/locations", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Locations(r.Context())
		deskResult(w, v, e)
	}))
	for _, pattern := range []string{"POST /api/cmdb/locations", "PUT /api/cmdb/locations/{id}"} {
		mux.HandleFunc(pattern, s.requireAdmin(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
			var in struct {
				Code   string `json:"code"`
				Name   string `json:"name"`
				Active bool   `json:"is_active"`
			}
			if decodeJSON(w, r, &in) {
				v, e := s.desk.SaveLocation(r.Context(), u, idOf(r), in.Code, in.Name, in.Active)
				deskResult(w, v, e)
			}
		}))
	}
	mux.HandleFunc("GET /api/desk/tickets", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		filters := map[string]string{}
		for _, key := range []string{"q", "process", "state", "queue", "priority", "assignee", "service", "from", "to"} {
			filters[key] = r.URL.Query().Get(key)
		}
		v, e := s.desk.ListTickets(r.Context(), u, filters, number(r, "limit"), number(r, "offset"))
		deskResult(w, v, e)
	}))
	mux.HandleFunc("POST /api/desk/tickets", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in desk.TicketInput
		if decodeJSON(w, r, &in) {
			v, e := s.desk.CreateTicket(r.Context(), u, in)
			deskResult(w, v, e)
		}
	}))
	mux.HandleFunc("GET /api/desk/tickets/{id}", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Ticket(r.Context(), u, idOf(r))
		deskResult(w, v, e)
	}))
	mux.HandleFunc("POST /api/desk/tickets/{id}/actions", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in desk.Action
		if decodeJSON(w, r, &in) {
			v, e := s.desk.Act(r.Context(), u, idOf(r), in)
			deskResult(w, v, e)
		}
	}))
	mux.HandleFunc("POST /api/desk/tickets/{id}/comments", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in struct {
			Body       string `json:"body"`
			Visibility string `json:"visibility"`
		}
		if decodeJSON(w, r, &in) {
			deskResult(w, nil, s.desk.Comment(r.Context(), u, idOf(r), in.Body, in.Visibility))
		}
	}))
	mux.HandleFunc("POST /api/desk/tickets/{id}/links", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		var in struct {
			Target int64  `json:"target_id"`
			Kind   string `json:"kind"`
		}
		if decodeJSON(w, r, &in) {
			deskResult(w, nil, s.desk.LinkTickets(r.Context(), u, idOf(r), in.Target, in.Kind))
		}
	}))
	mux.HandleFunc("GET /api/desk/notifications", s.requireAuth(func(w http.ResponseWriter, r *http.Request, u *auth.User) {
		v, e := s.desk.Notifications(r.Context(), u)
		deskResult(w, v, e)
	}))
}
