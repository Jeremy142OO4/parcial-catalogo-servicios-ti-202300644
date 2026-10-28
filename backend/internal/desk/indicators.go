package desk

import (
	"context"
	"encoding/json"
	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/auth"
	"sort"
	"time"
)

type Interval struct {
	Start time.Time  `json:"started_at"`
	End   *time.Time `json:"ended_at"`
}
type ObjectiveSegment struct {
	ID         int64      `json:"id"`
	Queue      int64      `json:"queue_id"`
	Cycle      int        `json:"cycle"`
	Start      time.Time  `json:"started_at"`
	End        *time.Time `json:"ended_at"`
	First      *float64   `json:"first_response_minutes"`
	Resolution *float64   `json:"resolution_minutes"`
	Source     string     `json:"source"`
}
type CycleSource struct {
	Cycle    int        `json:"cycle"`
	Start    time.Time  `json:"started_at"`
	Resolved *time.Time `json:"resolved_at"`
	Closed   *time.Time `json:"closed_at"`
}
type metricSource struct {
	Ticket struct {
		ID       int64      `json:"id"`
		Number   string     `json:"number"`
		Queue    int64      `json:"queue_id"`
		Process  string     `json:"process_code"`
		Priority int        `json:"priority"`
		State    string     `json:"state"`
		Created  time.Time  `json:"created_at"`
		Updated  time.Time  `json:"updated_at"`
		First    *time.Time `json:"first_attended_at"`
	} `json:"ticket"`
	Pauses   []Interval         `json:"pauses"`
	Segments []ObjectiveSegment `json:"segments"`
	Cycles   []CycleSource      `json:"cycles"`
	Reopens  int                `json:"reopens"`
}

// Clip and merge intervals so overlap is never subtracted twice.
func activeMinutes(start, end time.Time, pauses []Interval) (float64, float64) {
	if !end.After(start) {
		return 0, 0
	}
	intervals := []Interval{}
	for _, p := range pauses {
		a := p.Start
		b := end
		if p.End != nil && p.End.Before(b) {
			b = *p.End
		}
		if a.Before(start) {
			a = start
		}
		if b.After(a) {
			bb := b
			intervals = append(intervals, Interval{a, &bb})
		}
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i].Start.Before(intervals[j].Start) })
	paused := time.Duration(0)
	if len(intervals) > 0 {
		a, b := intervals[0].Start, *intervals[0].End
		for _, p := range intervals[1:] {
			if !p.Start.After(b) {
				if p.End.After(b) {
					b = *p.End
				}
			} else {
				paused += b.Sub(a)
				a, b = p.Start, *p.End
			}
		}
		paused += b.Sub(a)
	}
	return (end.Sub(start) - paused).Minutes(), paused.Minutes()
}
func objectiveStatus(minutes float64, goal *float64, completed bool) string {
	if goal == nil {
		return "sin_objetivo"
	}
	if minutes > *goal {
		return "incumplido"
	}
	if completed {
		return "cumplido"
	}
	return "pendiente"
}
func combinedStatus(statuses []string) string {
	if len(statuses) == 0 {
		return "sin_objetivo"
	}
	out := "cumplido"
	for _, s := range statuses {
		if s == "incumplido" {
			return s
		}
		if s == "sin_objetivo" {
			out = s
		} else if s == "pendiente" && out != "sin_objetivo" {
			out = s
		}
	}
	return out
}
func calculateIndicators(src metricSource, now time.Time) map[string]any {
	end := now
	if src.Ticket.First != nil {
		end = *src.Ticket.First
	} else if src.Ticket.State == "cerrado" || src.Ticket.State == "rechazado" {
		end = src.Ticket.Updated
	}
	firstActive, firstPaused := activeMinutes(src.Ticket.Created, end, src.Pauses)
	var firstMinutes *float64
	if src.Ticket.First != nil {
		firstMinutes = &firstActive
	}
	firstStates := []string{}
	segments := []map[string]any{}
	cycles := []map[string]any{}
	for _, seg := range src.Segments {
		finish := now
		if src.Ticket.State == "cerrado" || src.Ticket.State == "rechazado" {
			finish = src.Ticket.Updated
		}
		if seg.End != nil {
			finish = *seg.End
		}
		elapsed, paused := activeMinutes(seg.Start, finish, src.Pauses)
		firstState := "no_aplica"
		if seg.Start.Before(end) || seg.Start.Equal(end) {
			cut := finish
			if end.Before(cut) {
				cut = end
			}
			m, _ := activeMinutes(seg.Start, cut, src.Pauses)
			firstState = objectiveStatus(m, seg.First, src.Ticket.First != nil || seg.End != nil)
			firstStates = append(firstStates, firstState)
		}
		segments = append(segments, map[string]any{"id": seg.ID, "queue_id": seg.Queue, "cycle": seg.Cycle, "started_at": seg.Start, "ended_at": seg.End, "source": seg.Source, "first_response_minutes": seg.First, "resolution_minutes": seg.Resolution, "active_minutes": elapsed, "paused_minutes": paused, "first_status": firstState, "resolution_status": objectiveStatus(elapsed, seg.Resolution, seg.End != nil)})
	}
	for _, c := range src.Cycles {
		finish := now
		historicalTerminal := c.Resolved == nil && c.Closed == nil && (src.Ticket.State == "cerrado" || src.Ticket.State == "rechazado")
		if historicalTerminal {
			finish = src.Ticket.Updated
		}
		if c.Closed != nil {
			finish = *c.Closed
		}
		if c.Resolved != nil {
			finish = *c.Resolved
		}
		elapsed, paused := activeMinutes(c.Start, finish, src.Pauses)
		var resolution *float64
		if c.Resolved != nil {
			resolution = &elapsed
		}
		states := []string{}
		for _, seg := range segments {
			if seg["cycle"] == c.Cycle {
				states = append(states, seg["resolution_status"].(string))
			}
		}
		status := combinedStatus(states)
		if c.Resolved == nil && (c.Closed != nil || historicalTerminal) {
			status = "sin_resolucion"
		}
		cycles = append(cycles, map[string]any{"cycle": c.Cycle, "started_at": c.Start, "resolved_at": c.Resolved, "closed_at": c.Closed, "resolution_minutes": resolution, "active_minutes": elapsed, "paused_minutes": paused, "status": status})
	}
	age := any(nil)
	if src.Ticket.State != "cerrado" && src.Ticket.State != "rechazado" {
		age = now.Sub(src.Ticket.Created).Minutes()
	}
	return map[string]any{"ticket_id": src.Ticket.ID, "number": src.Ticket.Number, "queue_id": src.Ticket.Queue, "process_code": src.Ticket.Process, "priority": src.Ticket.Priority, "state": src.Ticket.State, "open_age_minutes": age, "first_attention_minutes": firstMinutes, "first_pending_active_minutes": firstActive, "first_paused_minutes": firstPaused, "first_status": combinedStatus(firstStates), "reopens": src.Reopens, "cycles": cycles, "objective_segments": segments, "calculated_at": now, "clock": "calendar_utc"}
}

const metricSQL = `SELECT jsonb_build_object('ticket',to_jsonb(t),'pauses',coalesce((SELECT jsonb_agg(to_jsonb(p)) FROM ticket_pauses p WHERE p.ticket_id=t.id),'[]'),'segments',coalesce((SELECT jsonb_agg(to_jsonb(s) ORDER BY s.id) FROM ticket_objective_segments s WHERE s.ticket_id=t.id),'[]'),'cycles',coalesce((SELECT jsonb_agg(to_jsonb(c) ORDER BY c.cycle) FROM ticket_cycles c WHERE c.ticket_id=t.id),'[]'),'reopens',(SELECT count(*) FROM ticket_history h WHERE h.ticket_id=t.id AND action='reopen')) FROM tickets t WHERE ($1::bigint=0 OR t.id=$1) AND ($2 OR t.requester_id=$3 OR EXISTS(SELECT 1 FROM queue_grants g WHERE g.queue_id=t.queue_id AND g.can_read AND (g.user_id=$3 OR g.role_code IN(SELECT role_code FROM user_roles WHERE user_id=$3)))) AND ($4='' OR t.process_code=$4) AND ($5='' OR t.queue_id::text=$5) ORDER BY t.id`

func (s *Service) Indicators(ctx context.Context, u *auth.User, id int64, process, queue string) (map[string]any, error) {
	rows, e := objects(ctx, s.repo.pool, metricSQL, id, u.HasRole("admin") || u.ReadOnly(), u.ID, process, queue)
	if e != nil {
		return nil, e
	}
	if id != 0 && len(rows) == 0 {
		return nil, forbidden()
	}
	items := []map[string]any{}
	first, resolution := map[string]int{}, map[string]int{}
	byQueue, byProcess, byPriority, byState := map[string]int{}, map[string]int{}, map[int]int{}, map[string]int{}
	now := time.Now().UTC()
	for _, r := range rows {
		var src metricSource
		b, e := json.Marshal(r)
		if e != nil {
			return nil, e
		}
		if e = json.Unmarshal(b, &src); e != nil {
			return nil, e
		}
		v := calculateIndicators(src, now)
		items = append(items, v)
		first[v["first_status"].(string)]++
		cs := v["cycles"].([]map[string]any)
		if len(cs) > 0 {
			resolution[cs[len(cs)-1]["status"].(string)]++
		}
		byQueue[jsonNumber(src.Ticket.Queue)]++
		byProcess[src.Ticket.Process]++
		byPriority[src.Ticket.Priority]++
		byState[src.Ticket.State]++
	}
	return map[string]any{"items": items, "total": len(items), "first_attention": first, "resolution": resolution, "by_queue": byQueue, "by_process": byProcess, "by_priority": byPriority, "by_state": byState, "calculated_at": now, "clock": "calendar_utc"}, nil
}
func jsonNumber(id int64) string { b, _ := json.Marshal(id); return string(b) }
