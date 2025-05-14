package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luckermt/forum-app/forum-service/internal/service/mocks"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/assert"
)

func TestForumHandler_GetTopics(t *testing.T) {
	mockSvc := new(mocks.ForumService)
	mockSvc.On("GetTopics").Return([]*models.Topic{
		{ID: "1", Title: "Test Topic"},
	}, nil)

	handler := NewForumHandler(mockSvc)
	req := httptest.NewRequest("GET", "/topics", nil)
	w := httptest.NewRecorder()

	handler.GetTopics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}
