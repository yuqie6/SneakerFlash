package service

import (
	"SneakerFlash/internal/model"
	"SneakerFlash/internal/repository"
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/gorm"
)

type AuditService struct {
	repo *repository.AuditLogRepo
}

type AuditLogInput struct {
	ActorID      uint
	ActorName    string
	ActorRole    string
	Resource     string
	Action       string
	ResourceID   string
	RequestID    string
	RequestPath  string
	RequestIP    string
	RequestBody  any
	Result       string
	ErrorMessage string
}

func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{repo: repository.NewAuditLogRepo(db)}
}

func (s *AuditService) Record(ctx context.Context, input AuditLogInput) error {
	if ctx == nil {
		return fmt.Errorf("context is nil")
	}
	body := ""
	if input.RequestBody != nil {
		data, err := json.Marshal(input.RequestBody)
		if err != nil {
			return err
		}
		body = string(data)
	}
	return s.repo.Create(ctx, &model.AuditLog{
		ActorID:      input.ActorID,
		ActorName:    input.ActorName,
		ActorRole:    model.NormalizeUserRole(input.ActorRole),
		Resource:     input.Resource,
		Action:       input.Action,
		ResourceID:   input.ResourceID,
		RequestID:    input.RequestID,
		RequestPath:  input.RequestPath,
		RequestIP:    input.RequestIP,
		RequestBody:  body,
		Result:       input.Result,
		ErrorMessage: input.ErrorMessage,
	})
}

func (s *AuditService) List(ctx context.Context, filter repository.AuditLogFilter) ([]model.AuditLog, int64, error) {
	if ctx == nil {
		return nil, 0, fmt.Errorf("context is nil")
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 || filter.PageSize > 100 {
		filter.PageSize = 20
	}
	return s.repo.List(ctx, filter)
}
