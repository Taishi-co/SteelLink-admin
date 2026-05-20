package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Taishi-co/SteelLink-admin/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

func TestBusinessReadRoutesRejectOrdinaryRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine)

	token, err := auth.GenerateAdminToken(100, "ordinary", "Ordinary", 4)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/v1/dashboard/stats"},
		{http.MethodGet, "/api/v1/users"},
		{http.MethodGet, "/api/v1/enterprises"},
		{http.MethodGet, "/api/v1/products"},
		{http.MethodGet, "/api/v1/orders"},
		{http.MethodGet, "/api/v1/operation-logs"},
		{http.MethodPost, "/api/v1/files/upload"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		recorder := httptest.NewRecorder()

		engine.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusForbidden {
			t.Fatalf("%s %s status=%d, want %d", tc.method, tc.path, recorder.Code, http.StatusForbidden)
		}
	}
}

func TestOrdinaryRoleCanReadOwnAdminInfo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine)

	token, err := auth.GenerateAdminToken(100, "ordinary", "Ordinary", 4)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestOrdinaryRoleGetsEmptyMenus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine)

	token, err := auth.GenerateAdminToken(100, "ordinary", "Ordinary", 4)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/menus", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status=%d, want %d", recorder.Code, http.StatusOK)
	}
	if !strings.Contains(recorder.Body.String(), `"data":[]`) {
		t.Fatalf("ordinary menus body=%s, want empty data array", recorder.Body.String())
	}
}

func TestAIProcurementMonitorAllowsOperatorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine)

	token, err := auth.GenerateAdminToken(101, "operator", "Operator", 2)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-procurement/monitor", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code == http.StatusForbidden {
		t.Fatalf("operator should not be forbidden")
	}
}

func TestAIProcurementMonitorRejectsAuditorRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRoutes(engine)

	token, err := auth.GenerateAdminToken(102, "auditor", "Auditor", 3)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/ai-procurement/monitor", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	recorder := httptest.NewRecorder()

	engine.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status=%d, want %d", recorder.Code, http.StatusForbidden)
	}
}
