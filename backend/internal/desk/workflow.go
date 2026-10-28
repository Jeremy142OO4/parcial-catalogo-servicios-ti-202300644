package desk

import "strings"

var Routes = map[string][]string{
	"nuevos":     {"solicitado", "analisis", "pendiente_aprobacion", "aprobado", "implementacion", "validacion", "resuelto", "cerrado"},
	"cambios":    {"solicitado", "evaluacion", "pendiente_aprobacion", "aprobado", "planificado", "ejecucion", "validacion", "resuelto", "cerrado"},
	"incidentes": {"registrado", "clasificado", "diagnostico", "atencion", "recuperado", "resuelto", "cerrado"},
	"eventos":    {"registrado", "clasificado", "evaluado", "respuesta", "resuelto", "cerrado"},
	"problemas":  {"registrado", "investigacion", "causa_identificada", "solucion_propuesta", "solucion_aplicada", "resuelto", "cerrado"},
	"accesos":    {"solicitado", "validacion", "pendiente_aprobacion", "aprobado", "provision_o_revocacion", "verificacion", "resuelto", "cerrado"},
	"servicios":  {"solicitado", "validacion", "asignado", "cumplimiento", "confirmacion", "resuelto", "cerrado"},
}

func validTransition(process, from, to string) bool {
	for i, s := range Routes[process] {
		if s == from && i+1 < len(Routes[process]) && Routes[process][i+1] == to {
			return true
		}
	}
	return process == "eventos" && from == "evaluado" && to == "resuelto" || process == "incidentes" && from == "escalado" && to == "diagnostico" || process == "cambios" && from == "validacion" && to == "planificado"
}
func required(details map[string]any, keys ...string) error {
	for _, k := range keys {
		v, ok := details[k].(string)
		if !ok || strings.TrimSpace(v) == "" {
			return bad("campo requerido: " + k)
		}
	}
	return nil
}

// IncidentRule is a project decision, not an SLA. Minutes preserve inclusive source boundaries.
func IncidentRule(security, interrupted bool, minutes *float64) (string, bool, error) {
	if minutes != nil && *minutes < 0 {
		return "", false, bad("duración negativa")
	}
	if security {
		return "I6", false, nil
	}
	if !interrupted {
		return "I0", false, nil
	}
	if minutes == nil || *minutes == 0 {
		return "", true, nil
	}
	x := *minutes
	switch {
	case x <= 180:
		return "I1", false, nil
	case x <= 480:
		return "I2", false, nil
	case x <= 2880:
		return "I3", false, nil
	case x <= 10080:
		return "I4", false, nil
	case x < 43200:
		return "I5", true, nil
	default:
		return "I5", false, nil
	}
}
func categoryPriority(category string) int {
	if category == "I0" {
		return 3
	}
	if category == "I1" || category == "I2" {
		return 2
	}
	return 1
}
func validateIncident(d map[string]any) error {
	category, _ := d["category"].(string)
	if len(category) != 2 || category[0] != 'I' || category[1] < '0' || category[1] > '6' {
		return bad("categoría I0-I6 requerida")
	}
	security, ok := d["security"].(bool)
	if !ok {
		return bad("security requerido")
	}
	interrupted, ok := d["interrupted"].(bool)
	if !ok {
		return bad("interrupted requerido (incluye degradación)")
	}
	var minutes *float64
	for _, k := range []string{"estimated_minutes", "actual_minutes"} {
		if v, exists := d[k]; exists && v != nil {
			f, ok := v.(float64)
			if !ok || f < 0 {
				return bad("duración inválida")
			}
			if k == "actual_minutes" || minutes == nil {
				copy := f
				minutes = &copy
			}
		}
	}
	expected, review, e := IncidentRule(security, interrupted, minutes)
	if e != nil {
		return e
	}
	if review || expected != category {
		if e = required(d, "classification_reason"); e != nil {
			return e
		}
	}
	if security && category != "I6" {
		return bad("seguridad requiere I6")
	}
	if !interrupted && !security && category != "I0" {
		return bad("sin interrupción requiere I0")
	}
	redundancy, known := d["redundancy"].(bool)
	if !known {
		if v := d["redundancy"]; v != nil {
			return bad("redundancy debe ser booleano o null")
		}
	} else if redundancy != (category == "I0") {
		if e = required(d, "redundancy_reason"); e != nil {
			return e
		}
	}
	return nil
}
