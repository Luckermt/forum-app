package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luckermt/forum-app/forum-service/internal/service/mocks"
	"github.com/luckermt/forum-app/shared/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestForumHandler_GetTopics(t *testing.T) {
	mockSvc := new(mocks.ForumService)

	// Set up mock expectations for all called methods
	mockSvc.On("GetTopics", 1, 10, "").Return([]*models.Topic{
		{ID: "1", Title: "Test Topic"},
	}, 1, nil)
	mockSvc.On("BlockUser", mock.Anything, mock.Anything).Return(nil) // Add this line

	handler := NewForumHandler(mockSvc)
	req := httptest.NewRequest("GET", "/topics", nil)
	w := httptest.NewRecorder()

	handler.GetTopics(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	mockSvc.AssertExpectations(t)
}
