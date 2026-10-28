package desk

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"strconv"
	"strings"
)

type TicketInput struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Requester   int64          `json:"requester_id"`
	Service     int64          `json:"service_level2_id"`
	Process     string         `json:"process_code"`
	Queue       int64          `json:"queue_id"`
	Priority    int            `json:"priority"`
	Impact      string         `json:"impact"`
	Urgency     string         `json:"urgency"`
	Channel     string         `json:"channel"`
	NoCI        string         `json:"no_ci_reason"`
	CI          []int64        `json:"ci_ids"`
	Details     map[string]any `json:"details"`
}

func (s *Service) CreateTicket(ctx context.Context, u *auth.User, in TicketInput) (map[string]any, error) {
	var out map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error { var e error; out, e = s.createTicket(ctx, tx, u, in); return e })
	return out, err
}
func (s *Service) createTicket(ctx context.Context, tx pgx.Tx, u *auth.User, in TicketInput) (map[string]any, error) {
	route, ok := Routes[in.Process]
	if !ok {
		return nil, bad("proceso inválido")
	}
	if !u.HasRole("admin") && !u.HasRole("solicitante") && !s.allowed(ctx, tx, u, in.Queue, "handle") {
		return nil, forbidden()
	}
	if in.Requester == 0 {
		in.Requester = u.ID
	}
	if in.Requester != u.ID && !u.HasRole("admin") && !s.allowed(ctx, tx, u, in.Queue, "handle") {
		return nil, forbidden()
	}
	if strings.TrimSpace(in.Title) == "" || strings.TrimSpace(in.Description) == "" || strings.TrimSpace(in.Channel) == "" {
		return nil, bad("título, descripción y canal requeridos")
	}
	if len(in.CI) == 0 && strings.TrimSpace(in.NoCI) == "" {
		return nil, bad("justificar ticket sin CI")
	}
	var valid bool
	if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM queues WHERE id=$1 AND process_code=$2 AND is_active) AND EXISTS(SELECT 1 FROM services_level2 WHERE id=$3 AND is_active) AND EXISTS(SELECT 1 FROM app_users WHERE id=$4 AND is_active)`, in.Queue, in.Process, in.Service, in.Requester).Scan(&valid); e != nil {
		return nil, e
	}
	if !valid {
		return nil, bad("cola, servicio o solicitante inexistente/inactivo")
	}
	if in.Details == nil {
		in.Details = map[string]any{}
	}
	if in.Process == "incidentes" {
		if e := validateIncident(in.Details); e != nil {
			return nil, e
		}
		in.Priority = categoryPriority(in.Details["category"].(string))
	}
	if in.Process == "eventos" {
		c, _ := in.Details["category"].(string)
		if c != "E0" && c != "E1" && c != "E2" && c != "E3" {
			return nil, bad("categoría E0-E3 requerida")
		}
	}
	var n int64
	if e := tx.QueryRow(ctx, `SELECT nextval('ticket_number_seq')`).Scan(&n); e != nil {
		return nil, e
	}
	out, e := object(ctx, tx, `INSERT INTO tickets(number,title,description,requester_id,creator_id,service_level2_id,process_code,queue_id,state,priority,impact,urgency,channel,no_ci_reason,details) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15) RETURNING to_jsonb(tickets)`, fmt.Sprintf("SD-%08d", n), in.Title, in.Description, in.Requester, u.ID, in.Service, in.Process, in.Queue, route[0], in.Priority, in.Impact, in.Urgency, in.Channel, in.NoCI, jsonBytes(in.Details))
	if e != nil {
		return nil, e
	}
	id := int64(out["id"].(float64))
	for _, ci := range in.CI {
		var yes bool
		if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM configuration_items WHERE id=$1 AND is_active)`, ci).Scan(&yes); e != nil {
			return nil, e
		}
		if !yes {
			return nil, bad("CI inexistente/inactivo")
		}
		if _, e = tx.Exec(ctx, `INSERT INTO ticket_ci VALUES($1,$2)`, id, ci); e != nil {
			return nil, e
		}
	}
	if _, e = tx.Exec(ctx, `INSERT INTO ticket_cycles(ticket_id,cycle) VALUES($1,1)`, id); e != nil {
		return nil, e
	}
	if e = s.audit(ctx, tx, id, u.ID, "crear", nil, out, ""); e != nil {
		return nil, e
	}
	return out, nil
}
func (s *Service) readable(ctx context.Context, q queryer, u *auth.User, t map[string]any) bool {
	return int64(t["requester_id"].(float64)) == u.ID || s.allowed(ctx, q, u, int64(t["queue_id"].(float64)), "read")
}
func (s *Service) ListTickets(ctx context.Context, u *auth.User, filters map[string]string, limit, offset int) ([]map[string]any, error) {
	limit, offset = page(limit, offset)
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(t)||jsonb_build_object('service_level1_id',l.service_level1_id) FROM tickets t JOIN services_level2 l ON l.id=t.service_level2_id WHERE ($1 OR t.requester_id=$2 OR EXISTS(SELECT 1 FROM queue_grants g WHERE g.queue_id=t.queue_id AND g.can_read AND(g.user_id=$2 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$2)))) AND ($3='' OR t.number ILIKE '%'||$3||'%' OR t.title ILIKE '%'||$3||'%') AND($4='' OR t.process_code=$4) AND($5='' OR t.state=$5) AND($6='' OR t.queue_id::text=$6) AND($7='' OR t.priority::text=$7) AND($8='' OR t.assignee_id::text=$8) AND($9='' OR t.service_level2_id::text=$9) AND($10='' OR t.created_at>=nullif($10,'')::timestamptz) AND($11='' OR t.created_at<=nullif($11,'')::timestamptz) ORDER BY t.priority,t.created_at,t.id LIMIT $12 OFFSET $13`, u.HasRole("admin") || u.ReadOnly(), u.ID, filters["q"], filters["process"], filters["state"], filters["queue"], filters["priority"], filters["assignee"], filters["service"], filters["from"], filters["to"], limit, offset)
}
func (s *Service) Ticket(ctx context.Context, u *auth.User, id int64) (map[string]any, error) {
	t, e := object(ctx, s.repo.pool, `SELECT to_jsonb(t)||jsonb_build_object('service_level1_id',s.service_level1_id) FROM tickets t JOIN services_level2 s ON s.id=t.service_level2_id WHERE t.id=$1`, id)
	if e != nil {
		return nil, e
	}
	if !s.readable(ctx, s.repo.pool, u, t) {
		return nil, forbidden()
	}
	internal := !u.ReadOnly() && (u.HasRole("admin") || s.allowed(ctx, s.repo.pool, u, int64(t["queue_id"].(float64)), "handle"))
	comments, e := objects(ctx, s.repo.pool, `SELECT to_jsonb(c) FROM ticket_comments c WHERE ticket_id=$1 AND ($2 OR visibility='public') ORDER BY id`, id, internal)
	if e != nil {
		return nil, e
	}
	t["comments"] = comments
	history, e := objects(ctx, s.repo.pool, `SELECT CASE WHEN $2 THEN to_jsonb(h) ELSE jsonb_build_object('id',id,'action',action,'occurred_at',occurred_at,'actor_id',actor_id) END FROM ticket_history h WHERE ticket_id=$1 ORDER BY id`, id, internal)
	if e != nil {
		return nil, e
	}
	t["history"] = history
	cis, e := objects(ctx, s.repo.pool, `SELECT to_jsonb(c) FROM ticket_ci c WHERE ticket_id=$1`, id)
	if e != nil {
		return nil, e
	}
	t["ci"] = cis
	links, e := objects(ctx, s.repo.pool, `SELECT to_jsonb(l) FROM ticket_links l JOIN tickets a ON a.id=l.source_id JOIN tickets b ON b.id=l.target_id WHERE (l.source_id=$1 OR l.target_id=$1) AND ($2 OR (a.requester_id=$3 AND b.requester_id=$3))`, id, internal || u.ReadOnly(), u.ID)
	if e != nil {
		return nil, e
	}
	t["links"] = links
	return t, nil
}

type Action struct {
	Action     string         `json:"action"`
	To         string         `json:"to"`
	Reason     string         `json:"reason"`
	Resolution string         `json:"resolution"`
	Version    int64          `json:"version"`
	Queue      int64          `json:"queue_id"`
	Assignee   *int64         `json:"assignee_id"`
	Details    map[string]any `json:"details"`
	IncidentID int64          `json:"incident_id"`
}

func (s *Service) Act(ctx context.Context, u *auth.User, id int64, in Action) (map[string]any, error) {
	var out map[string]any
	err := s.transaction(ctx, u, func(tx pgx.Tx) error {
		t, e := object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1 FOR UPDATE`, id)
		if e != nil {
			return e
		}
		queue := int64(t["queue_id"].(float64))
		version := int64(t["version"].(float64))
		state := t["state"].(string)
		process := t["process_code"].(string)
		cycle := int(t["cycle"].(float64))
		requester := int64(t["requester_id"].(float64))
		if in.Version != version {
			return conflict("versión desactualizada")
		}
		if !s.readable(ctx, tx, u, t) {
			return forbidden()
		}
		permission := "handle"
		if in.Action == "approve" || in.Action == "reject" {
			permission = "approve"
		}
		if in.Action == "transfer" || in.Action == "escalate" {
			permission = "transfer"
		}
		own := requester == u.ID && (in.Action == "reopen" || in.Action == "close")
		if !own && !s.allowed(ctx, tx, u, queue, permission) {
			return forbidden()
		}
		var queueActive bool
		if e = tx.QueryRow(ctx, `SELECT is_active FROM queues WHERE id=$1 FOR SHARE`, queue).Scan(&queueActive); e != nil {
			return e
		}
		if !queueActive {
			return conflict("cola inactiva")
		}
		if (state == "cerrado" || state == "rechazado") && in.Action != "reopen" {
			return conflict("ticket terminal")
		}
		switch in.Action {
		case "claim", "assign":
			if state == "resuelto" {
				return conflict("ticket resuelto")
			}
			target := u.ID
			if in.Action == "claim" {
				if t["assignee_id"] != nil {
					return conflict("ticket ya reclamado")
				}
			} else {
				if in.Assignee == nil {
					return bad("agente requerido")
				}
				target = *in.Assignee
			}
			agent, e := s.loadUser(ctx, tx, target)
			if e != nil {
				return e
			}
			if !agent.HasRole("agente") && !agent.HasRole("admin") {
				return bad("responsable no es agente")
			}
			if !s.allowed(ctx, tx, agent, queue, "handle") {
				return forbidden()
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET assignee_id=$2 WHERE id=$1`, id, target)
			if e != nil {
				return e
			}
		case "transfer", "escalate":
			if strings.TrimSpace(in.Reason) == "" {
				return bad("motivo requerido")
			}
			var dest string
			var next, old int
			if e = tx.QueryRow(ctx, `SELECT process_code,support_level FROM queues WHERE id=$1 AND is_active FOR SHARE`, in.Queue).Scan(&dest, &next); e != nil {
				return bad("cola destino inexistente/inactiva")
			}
			if dest != process {
				return bad("proceso destino incompatible")
			}
			if !s.allowed(ctx, tx, u, in.Queue, "read") {
				return forbidden()
			}
			if in.Action == "escalate" {
				if process != "incidentes" || state != "atencion" {
					return bad("escalamiento requiere incidente en atención")
				}
				if e = tx.QueryRow(ctx, `SELECT support_level FROM queues WHERE id=$1`, queue).Scan(&old); e != nil {
					return e
				}
				if next <= old {
					return bad("nivel destino debe ser superior")
				}
				state = "escalado"
			}
			if in.Assignee != nil {
				agent, e := s.loadUser(ctx, tx, *in.Assignee)
				if e != nil {
					return e
				}
				if (!agent.HasRole("agente") && !agent.HasRole("admin")) || !s.allowed(ctx, tx, agent, in.Queue, "handle") {
					return bad("agente destino no autorizado")
				}
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET queue_id=$2,assignee_id=$3,state=$4 WHERE id=$1`, id, in.Queue, in.Assignee, state)
			if e != nil {
				return e
			}
		case "details":
			if state == "resuelto" {
				return conflict("reabrir para editar")
			}
			if in.Details == nil {
				return bad("details requerido")
			}
			var approved bool
			if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticket_approvals WHERE ticket_id=$1 AND cycle=$2 AND decision='approved')`, id, cycle).Scan(&approved); e != nil {
				return e
			}
			if approved {
				return conflict("alcance aprobado, usar reevaluate antes de modificar")
			}
			if state == "aprobado" || state == "planificado" || state == "ejecucion" || state == "implementacion" || state == "provision_o_revocacion" {
				return conflict("no editar alcance aprobado, reabrir evaluación")
			}
			if process == "incidentes" {
				if e = validateIncident(in.Details); e != nil {
					return e
				}
			}
			if process == "eventos" {
				if state != "registrado" && state != "clasificado" {
					return conflict("clasificación de evento ya evaluada")
				}
				c, _ := in.Details["category"].(string)
				if c != "E0" && c != "E1" && c != "E2" && c != "E3" {
					return bad("categoría E0-E3 requerida")
				}
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET details=$2 WHERE id=$1`, id, jsonBytes(in.Details))
			if e != nil {
				return e
			}
			if process == "incidentes" {
				if _, e = tx.Exec(ctx, `UPDATE tickets SET priority=$2 WHERE id=$1`, id, categoryPriority(in.Details["category"].(string))); e != nil {
					return e
				}
			}
		case "approve", "reject":
			if state != "pendiente_aprobacion" {
				return bad("ticket no pendiente de aprobación")
			}
			if process == "accesos" && requester == u.ID {
				return forbidden()
			}
			if strings.TrimSpace(in.Reason) == "" {
				return bad("motivo de decisión requerido")
			}
			decision := "approved"
			next := "aprobado"
			if in.Action == "reject" {
				decision = "rejected"
				next = "rechazado"
			}
			if _, e = tx.Exec(ctx, `INSERT INTO ticket_approvals(ticket_id,cycle,approver_id,decision,reason) VALUES($1,$2,$3,$4,$5)`, id, cycle, u.ID, decision, in.Reason); e != nil {
				return e
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET state=$2 WHERE id=$1`, id, next)
			if e != nil {
				return e
			}
		case "transition", "close":
			to := in.To
			if in.Action == "close" {
				to = "cerrado"
			}
			if to == "aprobado" || to == "rechazado" {
				return bad("usar acción de aprobación")
			}
			if !validTransition(process, state, to) {
				return bad("transición inválida")
			}
			if e = s.guard(ctx, tx, t, to); e != nil {
				return e
			}
			resolution := t["resolution"]
			if in.Resolution != "" {
				resolution = in.Resolution
			}
			if to == "resuelto" && (resolution == nil || strings.TrimSpace(fmt.Sprint(resolution)) == "") {
				return bad("resolución requerida")
			}
			if to == "cerrado" && (strings.TrimSpace(in.Reason) == "" || resolution == nil) {
				return bad("confirmación y resolución requeridas")
			}
			if to == "resuelto" || to == "cerrado" {
				var paused bool
				if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticket_pauses WHERE ticket_id=$1 AND ended_at IS NULL)`, id).Scan(&paused); e != nil {
					return e
				}
				if paused {
					return bad("reanudar antes de resolver/cerrar")
				}
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET state=$2,resolution=$3,first_attended_at=coalesce(first_attended_at,now()),resolved_at=CASE WHEN $2='resuelto' THEN now() ELSE resolved_at END,closed_at=CASE WHEN $2='cerrado' THEN now() ELSE closed_at END WHERE id=$1`, id, to, resolution)
			if e != nil {
				return e
			}
			if to == "resuelto" || to == "cerrado" {
				_, e = tx.Exec(ctx, `UPDATE ticket_cycles SET resolved_at=CASE WHEN $3='resuelto' THEN now() ELSE resolved_at END,closed_at=CASE WHEN $3='cerrado' THEN now() ELSE closed_at END,resolution=$4 WHERE ticket_id=$1 AND cycle=$2`, id, cycle, to, resolution)
				if e != nil {
					return e
				}
			}
		case "reevaluate":
			if (process != "nuevos" && process != "cambios" && process != "accesos") || strings.TrimSpace(in.Reason) == "" || state == "resuelto" {
				return bad("reevaluación requiere proceso de aprobación, motivo y ticket abierto")
			}
			next := "evaluacion"
			if process == "nuevos" {
				next = "analisis"
			}
			if process == "accesos" {
				next = "validacion"
			}
			if _, e = tx.Exec(ctx, `UPDATE ticket_cycles SET closed_at=now() WHERE ticket_id=$1 AND cycle=$2`, id, cycle); e != nil {
				return e
			}
			if _, e = tx.Exec(ctx, `UPDATE tickets SET state=$2,cycle=cycle+1 WHERE id=$1`, id, next); e != nil {
				return e
			}
			if _, e = tx.Exec(ctx, `INSERT INTO ticket_cycles(ticket_id,cycle) VALUES($1,$2)`, id, cycle+1); e != nil {
				return e
			}
		case "reopen":
			if state != "cerrado" && state != "resuelto" {
				return bad("solo reabrir resuelto/cerrado")
			}
			if strings.TrimSpace(in.Reason) == "" {
				return bad("motivo requerido")
			}
			next := Routes[process][0]
			_, e = tx.Exec(ctx, `UPDATE tickets SET state=$2,cycle=cycle+1,resolution=NULL,resolved_at=NULL,closed_at=NULL WHERE id=$1`, id, next)
			if e != nil {
				return e
			}
			_, e = tx.Exec(ctx, `INSERT INTO ticket_cycles(ticket_id,cycle) VALUES($1,$2)`, id, cycle+1)
			if e != nil {
				return e
			}
		case "pause":
			if strings.TrimSpace(in.Reason) == "" || state == "resuelto" {
				return bad("pausa requiere motivo y ticket atendible")
			}
			_, e = tx.Exec(ctx, `INSERT INTO ticket_pauses(ticket_id,reason) VALUES($1,$2)`, id, in.Reason)
			if e != nil {
				return e
			}
		case "resume":
			tag, e := tx.Exec(ctx, `UPDATE ticket_pauses SET ended_at=now() WHERE ticket_id=$1 AND ended_at IS NULL`, id)
			if e != nil {
				return e
			}
			if tag.RowsAffected() != 1 {
				return bad("no hay pausa abierta")
			}
		case "process_event":
			if process != "eventos" || state != "evaluado" {
				return bad("requiere evento evaluado")
			}
			if e = s.eventIncident(ctx, tx, u, t, in.IncidentID); e != nil {
				return e
			}
		case "link_service":
			if process != "nuevos" || state != "implementacion" {
				return bad("requiere nuevo servicio aprobado en implementación")
			}
			var approved bool
			if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticket_approvals WHERE ticket_id=$1 AND cycle=$2 AND decision='approved')`, id, cycle).Scan(&approved); e != nil {
				return e
			}
			if !approved {
				return bad("aprobación requerida")
			}
			var service int64
			if v, ok := in.Details["service_id"].(float64); ok {
				service = int64(v)
			}
			var valid bool
			if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM services_level2 WHERE id=$1 AND is_active)`, service).Scan(&valid); e != nil {
				return e
			}
			if !valid {
				return bad("servicio resultado inválido")
			}
			_, e = tx.Exec(ctx, `UPDATE tickets SET details=jsonb_set(details,'{result_service_id}',to_jsonb($2::bigint)) WHERE id=$1 AND NOT details ? 'result_service_id'`, id, service)
			if e != nil {
				return e
			}
		default:
			return bad("acción inválida")
		}
		_, e = tx.Exec(ctx, `UPDATE tickets SET version=version+1,updated_at=now() WHERE id=$1`, id)
		if e != nil {
			return e
		}
		out, e = object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1`, id)
		if e != nil {
			return e
		}
		return s.audit(ctx, tx, id, u.ID, in.Action, t, out, in.Reason)
	})
	return out, err
}
func (s *Service) loadUser(ctx context.Context, q queryer, id int64) (*auth.User, error) {
	var u auth.User
	u.ID = id
	e := q.QueryRow(ctx, `SELECT role,is_active,ARRAY(SELECT role_code FROM user_roles WHERE user_id=$1 ORDER BY role_code) FROM app_users WHERE id=$1 FOR SHARE`, id).Scan(&u.Role, &u.IsActive, &u.Roles)
	if e != nil || !u.IsActive {
		return nil, bad("usuario inexistente/inactivo")
	}
	return &u, nil
}
func (s *Service) guard(ctx context.Context, tx pgx.Tx, t map[string]any, to string) error {
	p := t["process_code"].(string)
	d := t["details"].(map[string]any)
	id := int64(t["id"].(float64))
	cycle := int(t["cycle"].(float64))
	if to == "ejecucion" || to == "implementacion" || to == "provision_o_revocacion" {
		var approved bool
		if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ticket_approvals WHERE ticket_id=$1 AND cycle=$2 AND decision='approved')`, id, cycle).Scan(&approved); e != nil {
			return e
		}
		if !approved {
			return bad("aprobación vigente requerida")
		}
	}
	switch p {
	case "nuevos":
		if to == "pendiente_aprobacion" {
			return required(d, "need", "feasibility", "proposed_service_name")
		}
		if to == "validacion" || to == "resuelto" {
			if d["result_service_id"] == nil {
				return bad("vincular servicio resultado")
			}
		}
	case "cambios":
		if to == "pendiente_aprobacion" || to == "planificado" || to == "ejecucion" {
			if e := required(d, "impact_assessment", "risk", "test_plan", "rollback_plan", "planned_at"); e != nil {
				return e
			}
		}
		if to == "validacion" {
			return required(d, "execution_result")
		}
		if to == "resuelto" {
			return required(d, "validation_result")
		}
	case "incidentes":
		if e := validateIncident(d); e != nil {
			return e
		}
		if to == "recuperado" || to == "resuelto" {
			return required(d, "solution")
		}
	case "eventos":
		if to == "evaluado" {
			return required(d, "evaluation")
		}
		if to == "resuelto" {
			c := d["category"]
			if c == "E2" || c == "E3" {
				var yes bool
				if e := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM event_incidents WHERE event_ticket_id=$1) AND EXISTS(SELECT 1 FROM notifications WHERE dedup_key=$2)`, id, "event-"+strconv.FormatInt(id, 10)).Scan(&yes); e != nil {
					return e
				}
				if !yes {
					return bad("incidente y notificación requeridos")
				}
			}
		}
	case "problemas":
		if to == "causa_identificada" || to == "resuelto" {
			if e := required(d, "root_cause"); e != nil {
				return e
			}
		}
		if v, _ := d["known_error"].(bool); v {
			if e := required(d, "root_cause", "workaround"); e != nil {
				return e
			}
		}
		if to == "solucion_aplicada" || to == "resuelto" {
			return required(d, "definitive_solution")
		}
	case "accesos":
		if to == "pendiente_aprobacion" {
			if e := required(d, "resource", "permission", "operation"); e != nil {
				return e
			}
			if d["operation"] != "grant" && d["operation"] != "revoke" {
				return bad("operación de acceso inválida")
			}
			beneficiary, ok := d["beneficiary_id"].(float64)
			if !ok {
				return bad("beneficiario requerido")
			}
			if _, e := s.loadUser(ctx, tx, int64(beneficiary)); e != nil {
				return e
			}
		}
		if to == "verificacion" || to == "resuelto" {
			return required(d, "provision_evidence")
		}
	case "servicios":
		if to == "asignado" && t["assignee_id"] == nil {
			return bad("asignar agente antes de atender")
		}
		if to == "confirmacion" {
			return required(d, "fulfillment_evidence")
		}
		if to == "resuelto" {
			return required(d, "confirmation")
		}
	}
	return nil
}
func (s *Service) eventIncident(ctx context.Context, tx pgx.Tx, u *auth.User, t map[string]any, existing int64) error {
	d := t["details"].(map[string]any)
	if d["category"] != "E2" && d["category"] != "E3" {
		return bad("E0/E1 no generan incidente")
	}
	id := int64(t["id"].(float64))
	var linked int64
	e := tx.QueryRow(ctx, `SELECT incident_ticket_id FROM event_incidents WHERE event_ticket_id=$1`, id).Scan(&linked)
	if e == nil {
		return nil
	}
	if e != pgx.ErrNoRows {
		return e
	}
	if existing > 0 {
		other, e := object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1 AND process_code='incidentes'`, existing)
		if e != nil {
			return e
		}
		if !s.readable(ctx, tx, u, other) {
			return forbidden()
		}
		linked = existing
	} else {
		var q int64
		if e = tx.QueryRow(ctx, `SELECT id FROM queues WHERE process_code='incidentes' AND is_active ORDER BY id LIMIT 1`).Scan(&q); e != nil {
			return bad("sin cola de incidentes")
		}
		if !s.allowed(ctx, tx, u, q, "handle") {
			return forbidden()
		}
		details := map[string]any{"category": "I1", "security": false, "interrupted": true, "estimated_minutes": float64(60), "redundancy": nil, "classification_reason": "Clasificación inicial por agente, requiere diagnóstico"}
		// Unknown durations are preserved, never inferred from the event category.
		details["estimated_minutes"] = nil
		if d["category"] == "E3" {
			details["category"] = "I6"
			details["security"] = true
		}
		incident, e := s.createTicket(ctx, tx, u, TicketInput{Title: "Incidente de " + t["number"].(string), Description: t["description"].(string), Requester: int64(t["requester_id"].(float64)), Service: int64(t["service_level2_id"].(float64)), Process: "incidentes", Queue: q, Priority: 1, Impact: t["impact"].(string), Urgency: t["urgency"].(string), Channel: "evento", NoCI: "CI pendiente de diagnóstico, ver evento origen", Details: details})
		if e != nil {
			return e
		}
		linked = int64(incident["id"].(float64))
	}
	if _, e = tx.Exec(ctx, `INSERT INTO event_incidents VALUES($1,$2)`, id, linked); e != nil {
		return e
	}
	_, e = tx.Exec(ctx, `INSERT INTO notifications(ticket_id,recipient_id,message,dedup_key) VALUES($1,$2,$3,$4) ON CONFLICT(dedup_key) DO NOTHING`, id, int64(t["requester_id"].(float64)), fmt.Sprintf("Evento vinculado al incidente %d", linked), fmt.Sprintf("event-%d", id))
	return e
}
func (s *Service) Comment(ctx context.Context, u *auth.User, id int64, body, visibility string) error {
	return s.transaction(ctx, u, func(tx pgx.Tx) error {
		t, e := object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1 FOR UPDATE`, id)
		if e != nil {
			return e
		}
		if !s.readable(ctx, tx, u, t) {
			return forbidden()
		}
		if visibility != "public" && visibility != "internal" {
			return bad("visibilidad inválida")
		}
		if visibility == "internal" && !s.allowed(ctx, tx, u, int64(t["queue_id"].(float64)), "handle") {
			return forbidden()
		}
		_, e = tx.Exec(ctx, `INSERT INTO ticket_comments(ticket_id,author_id,body,visibility) VALUES($1,$2,$3,$4)`, id, u.ID, body, visibility)
		if e != nil {
			return e
		}
		return s.audit(ctx, tx, id, u.ID, "comment", nil, map[string]any{"visibility": visibility}, "")
	})
}
func (s *Service) LinkTickets(ctx context.Context, u *auth.User, from, to int64, kind string) error {
	return s.transaction(ctx, u, func(tx pgx.Tx) error {
		a, e := object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1 FOR UPDATE`, from)
		if e != nil {
			return e
		}
		b, e := object(ctx, tx, `SELECT to_jsonb(t) FROM tickets t WHERE id=$1`, to)
		if e != nil {
			return e
		}
		if !s.allowed(ctx, tx, u, int64(a["queue_id"].(float64)), "handle") || !s.readable(ctx, tx, u, b) {
			return forbidden()
		}
		if kind == "incidente_problema" && (a["process_code"] != "incidentes" || b["process_code"] != "problemas") {
			return bad("vínculo incidente/problema inválido")
		}
		if kind == "problema_cambio" && (a["process_code"] != "problemas" || b["process_code"] != "cambios") {
			return bad("vínculo problema/cambio inválido")
		}
		_, e = tx.Exec(ctx, `INSERT INTO ticket_links VALUES($1,$2,$3)`, from, to, kind)
		if e != nil {
			return e
		}
		return s.audit(ctx, tx, from, u.ID, "link", nil, map[string]any{"target": to, "kind": kind}, "")
	})
}
func (s *Service) Notifications(ctx context.Context, u *auth.User) ([]map[string]any, error) {
	return objects(ctx, s.repo.pool, `SELECT to_jsonb(n) FROM notifications n WHERE recipient_id=$1 ORDER BY id`, u.ID)
}
