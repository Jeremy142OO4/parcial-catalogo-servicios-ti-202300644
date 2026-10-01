package importer

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

const (
	DefaultSheet    = "Servicios Externos"
	HeaderRow       = 4
	FirstDataRow    = 5
	LastDataRow     = 101
	OptionsFirstRow = 112
	OptionsLastRow  = 122
	ExpectedLevel1  = 12
	ExpectedLevel2  = 46
)

type SourceRef struct {
	Sheet           string   `json:"sheet"`
	Row             int      `json:"row"`
	Range           string   `json:"range"`
	Transformations []string `json:"transformations,omitempty"`
}

type SourceName struct {
	Value string    `json:"value"`
	Ref   SourceRef `json:"source"`
}

type ServiceLevel1 struct {
	Code        string       `json:"code"`
	Name        string       `json:"name"`
	SourceNames []SourceName `json:"source_names"`
	Source      SourceRef    `json:"source"`
}

type ServiceLevel2 struct {
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	ParentCode     string    `json:"parent_code"`
	Active         *string   `json:"active"`
	ServiceClass   *string   `json:"service_class"`
	Criticality    *string   `json:"criticality"`
	ServiceType    *string   `json:"service_type"`
	Description    *string   `json:"description"`
	Metric         *string   `json:"metric"`
	Minimum        *float64  `json:"minimum"`
	Maximum        *float64  `json:"maximum"`
	ReviewRequired bool      `json:"review_required"`
	Source         SourceRef `json:"source"`
}

type Observation struct {
	Code       string      `json:"code"`
	Severity   string      `json:"severity"`
	Message    string      `json:"message"`
	SourceRefs []SourceRef `json:"sources,omitempty"`
	Details    []string    `json:"details,omitempty"`
}

type IgnoredRow struct {
	Row    int       `json:"row"`
	Reason string    `json:"reason"`
	Source SourceRef `json:"source"`
}

type Counts struct {
	Level1   int `json:"level1"`
	Level2   int `json:"level2"`
	Created  int `json:"created"`
	Updated  int `json:"updated"`
	Omitted  int `json:"omitted"`
	Observed int `json:"observed"`
}

type Report struct {
	SourceFile       string          `json:"source_file"`
	Sheet            string          `json:"sheet"`
	DataRange        string          `json:"data_range"`
	OptionsRange     string          `json:"options_range"`
	Counts           Counts          `json:"counts"`
	ServicesLevel1   []ServiceLevel1 `json:"services_level1"`
	ServicesLevel2   []ServiceLevel2 `json:"services_level2"`
	Observations     []Observation   `json:"observations"`
	IgnoredRows      []IgnoredRow    `json:"ignored_rows"`
	ValidationErrors []string        `json:"validation_errors,omitempty"`
}

type mergeRange struct {
	Ref       string
	StartCell string
	StartRow  int
	StartCol  int
	EndRow    int
	EndCol    int
	Value     string
}

type cellValue struct {
	Value     string
	Range     string
	WasMerged bool
}

type reader struct {
	file   *excelize.File
	sheet  string
	merges []mergeRange
}

func Run(path, sheet string) (*Report, error) {
	if sheet == "" {
		sheet = DefaultSheet
	}

	file, err := excelize.OpenFile(path)
	if err != nil {
		return nil, fmt.Errorf("open workbook: %w", err)
	}
	defer file.Close()

	sheetIndex, err := file.GetSheetIndex(sheet)
	if err != nil {
		return nil, fmt.Errorf("find sheet %q: %w", sheet, err)
	}
	if sheetIndex < 0 {
		return nil, fmt.Errorf("sheet %q not found", sheet)
	}

	merges, err := buildMergeRanges(file, sheet)
	if err != nil {
		return nil, fmt.Errorf("read merged ranges: %w", err)
	}
	r := &reader{file: file, sheet: sheet, merges: merges}
	report := &Report{
		SourceFile:   path,
		Sheet:        sheet,
		DataRange:    fmt.Sprintf("A%d:L%d", FirstDataRow, LastDataRow),
		OptionsRange: fmt.Sprintf("E%d:H%d", OptionsFirstRow, OptionsLastRow),
	}

	if err := validateHeaders(r); err != nil {
		return nil, err
	}

	level1ByCode := make(map[string]*ServiceLevel1)
	for row := FirstDataRow; row <= LastDataRow; row++ {
		code, ok, source := r.physicalCode(row, 1)
		if !ok {
			continue
		}
		name := strings.TrimSpace(r.resolve(row, 2).Value)
		nameRef := SourceRef{Sheet: sheet, Row: row, Range: fmt.Sprintf("A%d:B%d", row, row)}
		if resolved := r.resolve(row, 2); resolved.WasMerged {
			nameRef.Range = resolved.Range
			nameRef.Transformations = []string{"merged_parent_value_recovered_from_main_cell"}
		}
		candidate := SourceName{Value: name, Ref: nameRef}
		if existing, found := level1ByCode[code]; found {
			existing.SourceNames = append(existing.SourceNames, candidate)
			for i := range report.ServicesLevel1 {
				if report.ServicesLevel1[i].Code == code {
					report.ServicesLevel1[i] = *existing
					break
				}
			}
			if name != "" && existing.Name != name {
				report.Observations = append(report.Observations, Observation{
					Code:       "N1_NAME_CONFLICT",
					Severity:   "warning",
					Message:    fmt.Sprintf("El código %s aparece con nombres diferentes; se conserva el primer nombre como canónico.", code),
					SourceRefs: []SourceRef{existing.Source, source},
					Details:    []string{fmt.Sprintf("canónico: %s", existing.Name), fmt.Sprintf("alternativo: %s", name)},
				})
			}
			continue
		}
		entry := ServiceLevel1{
			Code:        code,
			Name:        name,
			SourceNames: []SourceName{candidate},
			Source:      source,
		}
		level1ByCode[code] = &entry
		report.ServicesLevel1 = append(report.ServicesLevel1, entry)
	}

	level2Seen := make(map[string]bool)
	for row := FirstDataRow; row <= LastDataRow; row++ {
		code, ok, source := r.physicalCode(row, 3)
		if !ok {
			if r.hasVisibleValue(row) {
				report.IgnoredRows = append(report.IgnoredRows, IgnoredRow{
					Row:    row,
					Reason: "continuation_or_non_service_row_without_level2_code",
					Source: SourceRef{Sheet: sheet, Row: row, Range: fmt.Sprintf("A%d:L%d", row, row)},
				})
			}
			continue
		}
		if level2Seen[code] {
			report.Observations = append(report.Observations, Observation{
				Code: "DUPLICATE_LEVEL2_CODE", Severity: "error",
				Message:    fmt.Sprintf("El código de nivel 2 %s aparece más de una vez.", code),
				SourceRefs: []SourceRef{source},
			})
			continue
		}
		level2Seen[code] = true

		parentCode := parentCodeFromLevel2(code)
		if _, found := level1ByCode[parentCode]; !found {
			report.Observations = append(report.Observations, Observation{
				Code: "MISSING_LEVEL1_PARENT", Severity: "error",
				Message:    fmt.Sprintf("El servicio %s no tiene un nivel 1 identificado por el prefijo %s.", code, parentCode),
				SourceRefs: []SourceRef{source},
			})
		}

		entry := ServiceLevel2{
			Code:         code,
			Name:         strings.TrimSpace(r.resolve(row, 4).Value),
			ParentCode:   parentCode,
			Active:       nullable(r.resolve(row, 5).Value),
			ServiceClass: nullable(r.resolve(row, 6).Value),
			Criticality:  nullable(r.resolve(row, 7).Value),
			ServiceType:  nullable(r.resolve(row, 8).Value),
			Description:  nullable(r.resolve(row, 9).Value),
			Metric:       nullable(r.resolve(row, 10).Value),
			Source:       source,
		}
		entry.Source.Transformations = []string{"parent_code_derived_from_level2_code_prefix"}
		entry.Minimum, entry.Maximum = r.number(row, 11, report), r.number(row, 12, report)
		if entry.Minimum != nil && entry.Maximum != nil && *entry.Minimum > *entry.Maximum {
			report.Observations = append(report.Observations, Observation{
				Code: "INVALID_THRESHOLD_RANGE", Severity: "error",
				Message:    fmt.Sprintf("El mínimo es mayor que el máximo para %s.", code),
				SourceRefs: []SourceRef{source},
			})
		}
		if parentCode == "SE.12" {
			entry.ReviewRequired = true
			report.Observations = append(report.Observations, Observation{
				Code: "INCOMPLETE_SE12_ATTRIBUTES", Severity: "warning",
				Message:    fmt.Sprintf("El servicio %s se importó con atributos desconocidos y requiere revisión.", code),
				SourceRefs: []SourceRef{source},
				Details:    []string{"active", "service_class", "criticality", "service_type", "description", "metric", "minimum", "maximum"},
			})
		}
		report.ServicesLevel2 = append(report.ServicesLevel2, entry)
	}

	report.Counts = Counts{
		Level1:   len(report.ServicesLevel1),
		Level2:   len(report.ServicesLevel2),
		Created:  len(report.ServicesLevel1) + len(report.ServicesLevel2),
		Updated:  0,
		Omitted:  len(report.IgnoredRows),
		Observed: len(report.Observations),
	}

	if report.Counts.Level1 != ExpectedLevel1 {
		report.ValidationErrors = append(report.ValidationErrors, fmt.Sprintf("se esperaban %d códigos de nivel 1 y se obtuvieron %d", ExpectedLevel1, report.Counts.Level1))
	}
	if report.Counts.Level2 != ExpectedLevel2 {
		report.ValidationErrors = append(report.ValidationErrors, fmt.Sprintf("se esperaban %d servicios de nivel 2 y se obtuvieron %d", ExpectedLevel2, report.Counts.Level2))
	}
	if len(report.ValidationErrors) > 0 {
		return report, fmt.Errorf("import validation failed: %s", strings.Join(report.ValidationErrors, "; "))
	}
	return report, nil
}

func WriteJSON(report *Report, path string) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal import report: %w", err)
	}
	return osWriteFile(path, append(data, '\n'))
}

var osWriteFile = func(path string, data []byte) error {
	return writeFile(path, data)
}

func buildMergeRanges(file *excelize.File, sheet string) ([]mergeRange, error) {
	merged, err := file.GetMergeCells(sheet)
	if err != nil {
		return nil, err
	}
	result := make([]mergeRange, 0, len(merged))
	for _, item := range merged {
		start, end := item.GetStartAxis(), item.GetEndAxis()
		startCol, startRow, err := excelize.CellNameToCoordinates(start)
		if err != nil {
			return nil, err
		}
		endCol, endRow, err := excelize.CellNameToCoordinates(end)
		if err != nil {
			return nil, err
		}
		value, err := file.GetCellValue(sheet, start)
		if err != nil {
			return nil, err
		}
		result = append(result, mergeRange{Ref: start + ":" + end, StartCell: start, StartRow: startRow, StartCol: startCol, EndRow: endRow, EndCol: endCol, Value: value})
	}
	return result, nil
}

func (r *reader) resolve(row, col int) cellValue {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	for _, merge := range r.merges {
		if row >= merge.StartRow && row <= merge.EndRow && col >= merge.StartCol && col <= merge.EndCol {
			return cellValue{Value: merge.Value, Range: merge.Ref, WasMerged: true}
		}
	}
	value, _ := r.file.GetCellValue(r.sheet, cell)
	return cellValue{Value: value}
}

func (r *reader) isPhysicalOrMergeMain(row, col int) bool {
	cell, _ := excelize.CoordinatesToCellName(col, row)
	for _, merge := range r.merges {
		if row >= merge.StartRow && row <= merge.EndRow && col >= merge.StartCol && col <= merge.EndCol {
			return cell == merge.StartCell
		}
	}
	return true
}

func (r *reader) physicalCode(row, col int) (string, bool, SourceRef) {
	ref := SourceRef{Sheet: r.sheet, Row: row, Range: fmt.Sprintf("%s%d", columnName(col), row)}
	if !r.isPhysicalOrMergeMain(row, col) {
		return "", false, ref
	}
	value := strings.TrimSpace(r.resolve(row, col).Value)
	return value, value != "", ref
}

func (r *reader) hasVisibleValue(row int) bool {
	for col := 1; col <= 12; col++ {
		if strings.TrimSpace(r.resolve(row, col).Value) != "" {
			return true
		}
	}
	return false
}

func (r *reader) number(row, col int, report *Report) *float64 {
	value := strings.TrimSpace(r.resolve(row, col).Value)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		report.Observations = append(report.Observations, Observation{
			Code: "INVALID_NUMERIC_VALUE", Severity: "error",
			Message:    fmt.Sprintf("El valor %q no es numérico.", value),
			SourceRefs: []SourceRef{{Sheet: r.sheet, Row: row, Range: fmt.Sprintf("%s%d", columnName(col), row)}},
		})
		return nil
	}
	return &parsed
}

func validateHeaders(r *reader) error {
	expected := []string{"COD.N1", "SERVICIO - Nivel 1", "COD.N2", "SERVICIO - Nivel 2", "ACTIVO", "CLASE DE SERVICIO", "CRITICIDAD", "TIPO DE SERVICIO", "Descripción", "Métrica", "Minimo", "Maximo"}
	for i, want := range expected {
		got := strings.TrimSpace(r.resolve(HeaderRow, i+1).Value)
		if got != want {
			return fmt.Errorf("unexpected header at %s%d: got %q, want %q", columnName(i+1), HeaderRow, got, want)
		}
	}
	return nil
}

func parentCodeFromLevel2(code string) string {
	parts := strings.Split(code, ".")
	if len(parts) < 3 {
		return ""
	}
	return strings.Join(parts[:2], ".")
}

func nullable(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	return &value
}

func columnName(col int) string {
	name, _ := excelize.ColumnNumberToName(col)
	return name
}

func writeFile(path string, data []byte) error {
	return writeFileImpl(path, data)
}
