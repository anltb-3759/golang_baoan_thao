package services

import (
	"context"
	"errors"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestCreatedByFromContext_Nil(t *testing.T) {
	result := createdByFromContext(nil)
	assert.Nil(t, result)
}

func TestCreatedByFromContext_WithActor(t *testing.T) {
	ctx := context.WithValue(context.Background(), "actor_user_id", "user-123")
	result := createdByFromContext(ctx)
	if assert.NotNil(t, result) {
		assert.Equal(t, "user-123", *result)
	}
}

func TestCreatedByFromContext_EmptyActor(t *testing.T) {
	ctx := context.WithValue(context.Background(), "actor_user_id", "   ")
	result := createdByFromContext(ctx)
	assert.Nil(t, result)
}

func TestCreatedByFromContext_NonStringValue(t *testing.T) {
	ctx := context.WithValue(context.Background(), "actor_user_id", 42)
	result := createdByFromContext(ctx)
	assert.Nil(t, result)
}

func TestServiceCatalogService_LogActivity_NilLogger(t *testing.T) {
	svc := NewServiceCatalogService(nil)
	// should not panic
	svc.logActivity(&models.ActivityLog{Action: "test"})
}

func TestServiceCatalogService_LogActivity_NilEntry(t *testing.T) {
	svc := NewServiceCatalogService(nil, &fakeActivityLogger{})
	// should not panic
	svc.logActivity(nil)
}

func TestServiceCatalogService_LogActivity_LoggerError(t *testing.T) {
	logger := &fakeActivityLogger{err: errors.New("log failed")}
	svc := NewServiceCatalogService(nil, logger)
	// should not propagate error
	svc.logActivity(&models.ActivityLog{Action: "service_type.create"})
	assert.Len(t, logger.entries, 1)
}

func TestServiceCatalogService_LogActivity_Success(t *testing.T) {
	logger := &fakeActivityLogger{}
	svc := NewServiceCatalogService(nil, logger)
	svc.logActivity(&models.ActivityLog{Action: "service_type.create"})
	assert.Len(t, logger.entries, 1)
	assert.Equal(t, "service_type.create", logger.entries[0].Action)
}
