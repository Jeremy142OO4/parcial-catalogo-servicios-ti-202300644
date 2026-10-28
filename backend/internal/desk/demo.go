package desk

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
)

// PrepareDemo adds identified fixtures without resetting cases already prepared.
// The command restricts this operation to an explicitly enabled demonstration database.
func (s *Service) PrepareDemo(ctx context.Context, password string) (map[string]any, error) {
	if err := s.ImportCMDBSource(ctx, "data/fuentes-service-desk/CMDB.xlsx"); err != nil {
		return nil, err
	}
	conn, err := s.repo.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	if _, err = conn.Exec(ctx, `SELECT pg_advisory_lock(20261028,37)`); err != nil {
		return nil, err
	}
	defer conn.Exec(context.Background(), `SELECT pg_advisory_unlock(20261028,37)`)
	a := auth.NewService(s.repo.pool, time.Hour)
	var position, service int64
	if err = s.repo.pool.QueryRow(ctx, `SELECT id FROM positions ORDER BY id LIMIT 1`).Scan(&position); err != nil {
		return nil, err
	}
	if err = s.repo.pool.QueryRow(ctx, `SELECT id FROM services_level2 WHERE is_active ORDER BY code LIMIT 1`).Scan(&service); err != nil {
		return nil, err
	}
	users := map[string]*auth.User{}
	for _, item := range [][2]string{{"desk-admin", "admin"}, {"desk-solicitante", "solicitante"}, {"desk-agente1", "agente"}, {"desk-agente2", "agente"}, {"desk-aprobador", "aprobador"}, {"desk-consulta", "consulta"}} {
		var exists bool
		if err = s.repo.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM app_users WHERE username=$1)`, item[0]).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			if _, err = a.CreateUser(ctx, position, "Demostración "+item[0], item[0], item[0]+"@example.local", password, item[1]); err != nil {
				return nil, err
			}
		}
		session, e := a.Login(ctx, item[0], password)
		if e != nil {
			return nil, fmt.Errorf("cuenta %s existente: no se restablece su contraseña: %w", item[0], e)
		}
		u := session.User
		if len(u.Roles) != 1 || !u.HasRole(item[1]) {
			return nil, fmt.Errorf("roles inesperados en %s, no se modifican", item[0])
		}
		if e = a.Logout(ctx, session.Token); e != nil {
			return nil, e
		}
		users[item[0]] = &u
	}
	admin, agent, requester, approver := users["desk-admin"], users["desk-agente1"], users["desk-solicitante"], users["desk-aprobador"]
	queues, e := s.Queues(ctx, admin)
	if e != nil {
		return nil, e
	}
	qs := map[string]int64{}
	for _, q := range queues {
		process := q["process_code"].(string)
		if qs[process] != 0 {
			return nil, fmt.Errorf("demo requiere una cola por proceso, duplicado %s", process)
		}
		qs[process] = int64(q["id"].(float64))
	}
	if len(qs) != 7 {
		return nil, fmt.Errorf("demo requiere siete colas, obtuvo %d", len(qs))
	}
	for process, q := range qs {
		var used bool
		if err = s.repo.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tickets WHERE queue_id=$1)`, q).Scan(&used); err != nil {
			return nil, err
		}
		if used {
			continue
		}
		first, resolution := 15, 120
		if _, e = s.SaveQueue(ctx, admin, q, QueueInput{Code: process, Name: "Demostración · " + process, Process: process, Level: 1, First: &first, Resolution: &resolution, Active: true}); e != nil {
			return nil, e
		}
		for _, role := range []string{"agente", "aprobador"} {
			r := role
			if e = s.Grant(ctx, admin, q, GrantInput{Role: &r, Read: true, Handle: role == "agente", Transfer: role == "agente", Approve: role == "aprobador"}); e != nil {
				return nil, e
			}
		}
	}
	if _, e = s.Generate(ctx, admin, "demostracion-5000", 20261028, 5000); e != nil {
		return nil, e
	}
	targets := map[string][]string{
		"incidentes": {"registrado", "clasificado", "diagnostico", "atencion", "recuperado", "cerrado", "cerrado"},
		"nuevos":     {"solicitado", "analisis", "pendiente_aprobacion", "implementacion", "cerrado"},
		"cambios":    {"solicitado", "evaluacion", "pendiente_aprobacion", "ejecucion", "cerrado"},
		"eventos":    {"registrado", "evaluado", "respuesta", "resuelto", "cerrado"},
		"problemas":  {"registrado", "investigacion", "causa_identificada", "solucion_propuesta", "cerrado"},
		"accesos":    {"solicitado", "validacion", "pendiente_aprobacion", "provision_o_revocacion", "rechazado"},
		"servicios":  {"solicitado", "validacion", "asignado", "cumplimiento", "cerrado"},
	}
	for _, process := range []string{"incidentes", "nuevos", "cambios", "eventos", "problemas", "accesos", "servicios"} {
		for i, target := range targets[process] {
			key := fmt.Sprintf("%s-%02d", process, i+1)
			var id int64
			var prepared *time.Time
			e = s.repo.pool.QueryRow(ctx, `SELECT ticket_id,prepared_at FROM demo_tickets WHERE dataset_key='desk-v1' AND case_key=$1`, key).Scan(&id, &prepared)
			if e != nil && e != pgx.ErrNoRows {
				return nil, e
			}
			if prepared != nil {
				continue
			}
			details := map[string]any{"need": "Necesidad de demostración", "feasibility": "Evaluación sintética viable", "proposed_service_name": "Servicio de demostración", "impact_assessment": "Impacto documentado sintético", "risk": "Bajo", "test_plan": "Verificar respuesta", "rollback_plan": "Restaurar configuración", "planned_at": "2026-10-28T18:00:00Z", "execution_result": "Ejecución comprobada", "validation_result": "Validación comprobada", "evaluation": "Evaluación sintética", "root_cause": "Configuración de prueba", "known_error": true, "workaround": "Alternativa temporal", "definitive_solution": "Configuración corregida", "beneficiary_id": float64(requester.ID), "resource": "Portal de demostración", "permission": "Lectura", "operation": "grant", "provision_evidence": "Permiso verificado", "fulfillment_evidence": "Requisición entregada", "confirmation": "Solicitante confirma entrega"}
			if process == "incidentes" {
				details["category"] = fmt.Sprintf("I%d", i)
				details["security"] = i == 6
				details["interrupted"] = i != 0
				details["estimated_minutes"] = []float64{0, 60, 300, 1000, 5000, 43200, 60}[i]
				details["redundancy"] = i == 0
				details["solution"] = "Servicio recuperado en demostración"
			}
			if process == "eventos" {
				details["category"] = []string{"E0", "E1", "E2", "E3", "E0"}[i]
			}
			if id == 0 {
				e = s.transaction(ctx, requester, func(tx pgx.Tx) error {
					t, e := s.createTicket(ctx, tx, requester, TicketInput{Title: "[DEMO] " + key, Description: "Caso sintético reproducible para evaluación. No corresponde a una incidencia real.", Service: service, Process: process, Queue: qs[process], Priority: 2, Impact: "medio", Urgency: "medio", Channel: "portal", NoCI: "Caso funcional de demostración sin equipo afectado", Details: details})
					if e != nil {
						return e
					}
					id = int64(t["id"].(float64))
					_, e = tx.Exec(ctx, `INSERT INTO demo_tickets(dataset_key,case_key,ticket_id) VALUES('desk-v1',$1,$2)`, key, id)
					return e
				})
				if e != nil {
					return nil, fmt.Errorf("crear %s: %w", key, e)
				}
			}
			t, e := s.Ticket(ctx, admin, id)
			if e != nil {
				return nil, e
			}
			act := func(u *auth.User, in Action) error {
				in.Version = int64(t["version"].(float64))
				var e error
				t, e = s.Act(ctx, u, id, in)
				return e
			}
			if t["assignee_id"] == nil {
				if e = act(admin, Action{Action: "assign", Assignee: &agent.ID}); e != nil {
					return nil, e
				}
			}
			for t["state"] != target {
				state := t["state"].(string)
				in := Action{Action: "transition", Reason: "Evidencia sintética de preparación", Resolution: "Resultado de demostración validado"}
				actor := agent
				if state == "pendiente_aprobacion" {
					actor = approver
					in.Action = "approve"
					if target == "rechazado" {
						in.Action = "reject"
					}
				} else {
					route := Routes[process]
					found := false
					for j, st := range route {
						if st == state && j+1 < len(route) {
							in.To = route[j+1]
							found = true
							break
						}
					}
					if !found {
						return nil, fmt.Errorf("estado inesperado %s en %s", state, key)
					}
				}
				if process == "nuevos" && state == "implementacion" && t["result_service_id"] == nil {
					if e = act(agent, Action{Action: "link_service", Details: map[string]any{"service_id": float64(service)}}); e != nil {
						return nil, e
					}
				}
				if process == "eventos" && state == "evaluado" {
					if i == 2 || i == 3 {
						var incident int64
						if e = s.repo.pool.QueryRow(ctx, `SELECT ticket_id FROM demo_tickets WHERE case_key='incidentes-02' AND dataset_key='desk-v1'`).Scan(&incident); e != nil {
							return nil, e
						}
						if e = act(agent, Action{Action: "process_event", IncidentID: incident}); e != nil {
							return nil, e
						}
					} else {
						in.To = "resuelto"
					}
				}
				if e = act(actor, in); e != nil {
					return nil, fmt.Errorf("preparar %s desde %s: %w", key, state, e)
				}
			}
			if process == "incidentes" && i == 3 {
				if e = act(agent, Action{Action: "pause", Reason: "Espera de proveedor de demostración"}); e != nil {
					return nil, e
				}
			}
			if process == "incidentes" && i == 5 {
				if e = act(agent, Action{Action: "reopen", Reason: "Reaparición sintética para demostrar ciclo 2"}); e != nil {
					return nil, e
				}
			}
			if _, e = s.repo.pool.Exec(ctx, `UPDATE demo_tickets SET prepared_at=now() WHERE ticket_id=$1`, id); e != nil {
				return nil, e
			}
		}
	}
	for _, article := range [][3]string{{"crear-ticket", "¿Cómo crear un ticket?", "Desde Mi portal selecciona el proceso y un servicio. Describe el caso y vincula los CI afectados o justifica su ausencia."}, {"seguimiento", "¿Dónde consultar el avance?", "Mi portal muestra únicamente tus solicitudes. Abre el ticket para consultar comentarios públicos, estado e historial autorizado."}, {"colas", "¿Cómo atiende un agente?", "En la consola filtra por cola, reclama el ticket y registra la atención. Las transiciones válidas dependen del proceso."}, {"aprobaciones", "¿Quién aprueba un acceso?", "Un usuario con permiso de aprobación en la cola. El solicitante no puede aprobar su propio acceso, aunque sea administrador."}, {"adjuntos", "¿Cómo adjuntar evidencia?", "Adjunta archivos permitidos de hasta 10 MiB. Los adjuntos internos solo son visibles para personal autorizado de la cola."}, {"cmdb", "¿Para qué sirve la CMDB?", "Relaciona equipos y dependencias con tickets. La vista de impacto muestra dependientes y servicios relacionados, con límites de recorrido explícitos."}, {"indicadores", "¿Cómo se calculan los indicadores?", "Se miden por ciclo con pausas y objetivos configurados. Un caso abierto o sin objetivo conserva valores desconocidos, no un cero ficticio."}, {"consulta", "¿Qué permite el rol consulta?", "Consultar información autorizada sin crear ni modificar registros, incluso si se combina con otros roles."}} {
		var exists bool
		if e = s.repo.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM faq_articles WHERE slug=$1)`, article[0]).Scan(&exists); e != nil {
			return nil, e
		}
		if !exists {
			if _, e = s.SaveFAQ(ctx, admin, 0, FAQInput{Slug: article[0], Title: article[1], Body: article[2], Published: true, Active: true}); e != nil {
				return nil, e
			}
		}
	}
	return object(ctx, s.repo.pool, `SELECT jsonb_build_object('dataset','desk-v1','tickets',(SELECT count(*) FROM demo_tickets WHERE dataset_key='desk-v1' AND prepared_at IS NOT NULL),'queues',(SELECT count(*) FROM queues),'faq',(SELECT count(*) FROM faq_articles WHERE is_active AND is_published),'by_process',(SELECT jsonb_object_agg(process_code,n) FROM (SELECT t.process_code,count(*) n FROM demo_tickets d JOIN tickets t ON t.id=d.ticket_id WHERE d.dataset_key='desk-v1' GROUP BY t.process_code) x))`)
}
