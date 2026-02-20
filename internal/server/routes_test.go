package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dangbamgong-backend/internal/dto"
	"dangbamgong-backend/internal/handler"
	"dangbamgong-backend/internal/service"

	"github.com/labstack/echo/v4"
)

type mockHealthRepo struct{}

func (m *mockHealthRepo) Ping() error { return nil }

func TestHelloWorldHandler(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp := httptest.NewRecorder()
	c := e.NewContext(req, resp)

	repo := &mockHealthRepo{}
	svc := service.NewHealthService(repo)
	h := handler.NewHealthHandler(svc)

	if err := h.HelloWorld(c); err != nil {
		t.Errorf("handler() error = %v", err)
		return
	}
	if resp.Code != http.StatusOK {
		t.Errorf("handler() wrong status code = %v", resp.Code)
		return
	}

	var actual dto.Response[map[string]string]
	if err := json.NewDecoder(resp.Body).Decode(&actual); err != nil {
		t.Errorf("handler() error decoding response body: %v", err)
		return
	}
	if !actual.Success {
		t.Errorf("handler() expected success=true, got false")
		return
	}
	if actual.Data["message"] != "Hello World" {
		t.Errorf("handler() wrong message = %v", actual.Data["message"])
		return
	}
}
