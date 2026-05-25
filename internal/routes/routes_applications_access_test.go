package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/awesome-academy/golang_baoan_thao/internal/middlewares"
	"github.com/awesome-academy/golang_baoan_thao/internal/models"
	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
)

func TestApplicationsRouteAccessPolicy(t *testing.T) {
	e := echo.New()

	// Inject claims into context to exercise route role middleware in isolation.
	admin := e.Group("/admin", func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			role := c.Request().Header.Get("X-Role")
			if role != "" {
				c.Set("user", &configs.JwtCustomClaims{ID: "u1", Role: role})
			}
			return next(c)
		}
	})

	ok := func(c *echo.Context) error { return c.String(http.StatusOK, "ok") }

	appsRead := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager, models.UserRoleSuperAdmin))
	appsRead.GET("", ok)
	appsRead.GET("/export", ok)
	appsRead.GET("/:id", ok)

	appsWrite := admin.Group("/applications", middlewares.AdminWebRequireRoles(models.UserRoleManager))
	appsWrite.POST("/:id/process", ok)
	appsWrite.GET("/:id/assign", ok)
	appsWrite.POST("/:id/assign", ok)

	tests := []struct {
		name   string
		role   string
		method string
		path   string
		want   int
	}{
		{
			name:   "super_admin can read list",
			role:   "super_admin",
			method: http.MethodGet,
			path:   "/admin/applications",
			want:   http.StatusOK,
		},
		{
			name:   "super_admin can export",
			role:   "super_admin",
			method: http.MethodGet,
			path:   "/admin/applications/export",
			want:   http.StatusOK,
		},
		{
			name:   "super_admin can view detail",
			role:   "super_admin",
			method: http.MethodGet,
			path:   "/admin/applications/a1",
			want:   http.StatusOK,
		},
		{
			name:   "super_admin cannot process",
			role:   "super_admin",
			method: http.MethodPost,
			path:   "/admin/applications/a1/process",
			want:   http.StatusSeeOther,
		},
		{
			name:   "super_admin cannot open assign form",
			role:   "super_admin",
			method: http.MethodGet,
			path:   "/admin/applications/a1/assign",
			want:   http.StatusSeeOther,
		},
		{
			name:   "manager can process",
			role:   "manager",
			method: http.MethodPost,
			path:   "/admin/applications/a1/process",
			want:   http.StatusOK,
		},
		{
			name:   "manager can assign",
			role:   "manager",
			method: http.MethodPost,
			path:   "/admin/applications/a1/assign",
			want:   http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("X-Role", tt.role)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.want, rec.Code)
		})
	}
}

