package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestNewRegistersOnlyCalendarV1Routes(t *testing.T) {
	engine := New(&HttpHandlers{
		Calendar: &Server{},
		HealthChecker: func(ctx *gin.Context) {
			ctx.Status(http.StatusOK)
		},
	}, zap.NewNop(), "appointment-service")

	for _, path := range []string{
		"/v1/calendar-events",
		"/v1/calendar-events/:id",
		"/v1/calendar-events/:calendar_event_id/reminder/resend",
		"/v1/services",
	} {
		if !hasRoute(engine, path) {
			t.Errorf("route %s is not registered", path)
		}
	}

	for _, path := range []string{"/appointments", "/admin/activities"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()
		engine.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Errorf("legacy route %s status = %d, want %d", path, response.Code, http.StatusNotFound)
		}
	}
}

func hasRoute(engine *gin.Engine, path string) bool {
	for _, route := range engine.Routes() {
		if route.Path == path {
			return true
		}
	}
	return false
}
