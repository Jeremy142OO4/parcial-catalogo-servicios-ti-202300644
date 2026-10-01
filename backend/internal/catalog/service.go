package catalog

import (
	"context"
	"fmt"

	"github.com/jeremy/parcial-catalogo-servicios-ti-202300644/backend/internal/importer"
)

type ImportRepository interface {
	PersistImport(context.Context, *importer.Report) error
}

type Service struct {
	repository ImportRepository
}

func NewService(repository ImportRepository) *Service {
	return &Service{repository: repository}
}

func (s *Service) PersistImport(ctx context.Context, report *importer.Report) error {
	if s == nil || s.repository == nil {
		return fmt.Errorf("catalog repository is required")
	}
	if report == nil {
		return fmt.Errorf("import report is required")
	}
	if len(report.ValidationErrors) > 0 {
		return fmt.Errorf("cannot persist invalid import: %v", report.ValidationErrors)
	}
	return s.repository.PersistImport(ctx, report)
}
