package docs

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "Service Management System API",
    "description": "API documentation for the Service Management System backend.",
    "version": "1.0.0"
  },
  "servers": [
    {
      "url": "http://localhost:8080",
      "description": "Local development"
    }
  ],
  "tags": [
    { "name": "Auth",           "description": "Authentication APIs" },
    { "name": "Citizen Profile","description": "Citizen profile and applications" },
    { "name": "Service Catalog","description": "Public service catalog" }
  ],
  "paths": {
    "/api/auth/register": {
      "post": {
        "tags": ["Auth"],
        "summary": "Register a new user",
        "parameters": [
          {
            "name": "Accept-Language",
            "in": "header",
            "schema": {
              "type": "string",
              "enum": ["vi", "en"],
              "default": "vi"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/RegisterRequest"
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "User registered successfully",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/UserResponse"
                }
              }
            }
          },
          "400": {
            "description": "Invalid request or validation error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "409": {
            "description": "Email already exists",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "/api/auth/login": {
      "post": {
        "tags": ["Auth"],
        "summary": "Login",
        "description": "Returns an access token in the response body and sets the refresh token in an HttpOnly cookie.",
        "parameters": [
          {
            "name": "Accept-Language",
            "in": "header",
            "schema": {
              "type": "string",
              "enum": ["vi", "en"],
              "default": "vi"
            }
          }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "application/json": {
              "schema": {
                "$ref": "#/components/schemas/LoginRequest"
              }
            }
          }
        },
        "responses": {
          "200": {
            "description": "Login successfully",
            "headers": {
              "Set-Cookie": {
                "description": "HttpOnly refresh token cookie",
                "schema": {
                  "type": "string",
                  "example": "refresh_token=...; Path=/; HttpOnly; SameSite=Lax"
                }
              }
            },
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/LoginResponse"
                }
              }
            }
          },
          "400": {
            "description": "Invalid request or validation error",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          },
          "401": {
            "description": "Invalid credentials",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "/api/auth/refresh": {
      "post": {
        "tags": ["Auth"],
        "summary": "Refresh access token",
        "description": "Reads the refresh token from the HttpOnly cookie and returns a new access token.",
        "parameters": [
          {
            "name": "Accept-Language",
            "in": "header",
            "schema": {
              "type": "string",
              "enum": ["vi", "en"],
              "default": "vi"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Access token refreshed successfully",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/TokenResponse"
                }
              }
            }
          },
          "401": {
            "description": "Missing, invalid, or expired refresh token",
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/ErrorResponse"
                }
              }
            }
          }
        }
      }
    },
    "/api/citizens/me": {
      "get": {
        "tags": ["Citizen Profile"],
        "summary": "Get my profile",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "responses": {
          "200": {
            "description": "Profile retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProfileResponse" } } }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Profile not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      },
      "put": {
        "tags": ["Citizen Profile"],
        "summary": "Update my profile",
        "description": "All fields are optional. CitizenIDNumber cannot be changed.",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": { "application/json": { "schema": { "$ref": "#/components/schemas/UpdateProfileRequest" } } }
        },
        "responses": {
          "200": {
            "description": "Profile updated successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ProfileResponse" } } }
          },
          "400": { "description": "Invalid request or validation error", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Profile not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/me/applications": {
      "post": {
        "tags": ["Applications"],
        "summary": "Submit a new application",
        "description": "Multipart form-data. Field 'data' (JSON string) + optional 'attachments[]' files (PDF/JPG/PNG, max 10 files, 10 MiB each, 30 MiB total).",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } }
        ],
        "requestBody": {
          "required": true,
          "content": {
            "multipart/form-data": {
              "schema": {
                "type": "object",
                "required": ["data"],
                "properties": {
                  "data": {
                    "type": "string",
                    "description": "JSON string of SubmitApplicationRequest",
                    "example": "{\"service_type_id\":\"uuid\",\"submitted_data\":{\"full_name\":\"Nguyen Van A\"}}"
                  },
                  "attachments[]": {
                    "type": "array",
                    "items": { "type": "string", "format": "binary" },
                    "description": "Optional file attachments (PDF/JPG/PNG)"
                  }
                }
              }
            }
          }
        },
        "responses": {
          "201": {
            "description": "Application submitted successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationDetailResponse" } } }
          },
          "400": { "description": "Missing or invalid 'data' field", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "422": { "description": "Service not found / inactive / missing required field / attachment limit exceeded", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      },
      "get": {
        "tags": ["Applications"],
        "summary": "List my submitted applications",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "page",  "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "limit", "in": "query", "schema": { "type": "integer", "default": 10, "maximum": 100 } }
        ],
        "responses": {
          "200": {
            "description": "Applications retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationListResponse" } } }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/me/applications/{id}": {
      "get": {
        "tags": ["Applications"],
        "summary": "Get a specific application detail",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": {
          "200": {
            "description": "Application retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ApplicationDetailResponse" } } }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Application not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/services": {
      "get": {
        "tags": ["Service Catalog"],
        "summary": "List available services",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "page",     "in": "query", "schema": { "type": "integer", "default": 1 } },
          { "name": "limit",    "in": "query", "schema": { "type": "integer", "default": 20, "maximum": 100 } },
          { "name": "category", "in": "query", "schema": { "type": "string" }, "description": "Filter by category" },
          { "name": "search",   "in": "query", "schema": { "type": "string" }, "description": "Search by name" }
        ],
        "responses": {
          "200": {
            "description": "Services retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ServiceListResponse" } } }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/citizens/services/{id}": {
      "get": {
        "tags": ["Service Catalog"],
        "summary": "Get service detail",
        "security": [{ "BearerAuth": [] }],
        "parameters": [
          { "name": "Accept-Language", "in": "header", "schema": { "type": "string", "enum": ["vi", "en"], "default": "vi" } },
          { "name": "id", "in": "path", "required": true, "schema": { "type": "string", "format": "uuid" } }
        ],
        "responses": {
          "200": {
            "description": "Service retrieved successfully",
            "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ServiceDetailResponse" } } }
          },
          "401": { "description": "Unauthorized", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } },
          "404": { "description": "Service not found", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrorResponse" } } } }
        }
      }
    },
    "/api/auth/logout": {
      "post": {
        "tags": ["Auth"],
        "summary": "Logout",
        "description": "Clears the refresh token HttpOnly cookie.",
        "parameters": [
          {
            "name": "Accept-Language",
            "in": "header",
            "schema": {
              "type": "string",
              "enum": ["vi", "en"],
              "default": "vi"
            }
          }
        ],
        "responses": {
          "200": {
            "description": "Logout successfully",
            "headers": {
              "Set-Cookie": {
                "description": "Clears the refresh token cookie",
                "schema": {
                  "type": "string",
                  "example": "refresh_token=; Path=/; Max-Age=0; HttpOnly; SameSite=Lax"
                }
              }
            },
            "content": {
              "application/json": {
                "schema": {
                  "$ref": "#/components/schemas/MessageResponse"
                }
              }
            }
          }
        }
      }
    }
  },
  "components": {
    "schemas": {
      "RegisterRequest": {
        "type": "object",
        "required": ["name", "email", "password"],
        "properties": {
          "name": {
            "type": "string",
            "example": "Nguyen Van A"
          },
          "email": {
            "type": "string",
            "format": "email",
            "example": "user@example.com"
          },
          "password": {
            "type": "string",
            "minLength": 6,
            "example": "123456"
          }
        }
      },
      "LoginRequest": {
        "type": "object",
        "required": ["email", "password"],
        "properties": {
          "email": {
            "type": "string",
            "format": "email",
            "example": "user@example.com"
          },
          "password": {
            "type": "string",
            "example": "123456"
          }
        }
      },
      "User": {
        "type": "object",
        "properties": {
          "id": {
            "type": "string",
            "format": "uuid"
          },
          "name": {
            "type": "string"
          },
          "email": {
            "type": "string",
            "format": "email"
          },
          "phone": {
            "type": "string"
          },
          "address": {
            "type": "string"
          },
          "role": {
            "type": "string",
            "enum": ["citizen", "staff", "manager", "super_admin"]
          },
          "status": {
            "type": "string",
            "enum": ["active", "blocked"]
          },
          "created_at": {
            "type": "string",
            "format": "date-time"
          },
          "updated_at": {
            "type": "string",
            "format": "date-time"
          }
        }
      },
      "UserResponse": {
        "type": "object",
        "properties": {
          "user:": {
            "$ref": "#/components/schemas/User"
          }
        }
      },
      "LoginResponse": {
        "type": "object",
        "properties": {
          "user": {
            "$ref": "#/components/schemas/User"
          },
          "token": {
            "type": "string",
            "example": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
          }
        }
      },
      "TokenResponse": {
        "type": "object",
        "properties": {
          "token": {
            "type": "string"
          }
        }
      },
      "MessageResponse": {
        "type": "object",
        "properties": {
          "message": {
            "type": "string"
          }
        }
      },
      "UpdateProfileRequest": {
        "type": "object",
        "properties": {
          "name":                       { "type": "string", "minLength": 1 },
          "phone":                      { "type": "string" },
          "address":                    { "type": "string" },
          "gender":                     { "type": "string" },
          "permanent_address":          { "type": "string" },
          "date_of_birth":              { "type": "string", "format": "date-time" },
          "email_notification_enabled": { "type": "boolean" }
        }
      },
      "CitizenProfile": {
        "type": "object",
        "properties": {
          "user_id":                    { "type": "string", "format": "uuid" },
          "name":                       { "type": "string" },
          "email":                      { "type": "string", "format": "email" },
          "phone":                      { "type": "string" },
          "address":                    { "type": "string" },
          "citizen_id_number":          { "type": "string", "example": "123456789012" },
          "gender":                     { "type": "string" },
          "permanent_address":          { "type": "string" },
          "date_of_birth":              { "type": "string", "format": "date-time", "nullable": true },
          "email_notification_enabled": { "type": "boolean" },
          "created_at":                 { "type": "string", "format": "date-time" },
          "updated_at":                 { "type": "string", "format": "date-time" }
        }
      },
      "ProfileResponse": {
        "type": "object",
        "properties": {
          "profile": { "$ref": "#/components/schemas/CitizenProfile" }
        }
      },
      "ApplicationItem": {
        "type": "object",
        "properties": {
          "id":                { "type": "string", "format": "uuid" },
          "application_code":  { "type": "string", "example": "APP-20240101-ABCDEF" },
          "service_type_id":   { "type": "string", "format": "uuid" },
          "service_type_name": { "type": "string" },
          "status":            { "type": "string", "enum": ["received", "processing", "approved", "rejected"] },
          "submitted_at":      { "type": "string", "format": "date-time" }
        }
      },
      "Pagination": {
        "type": "object",
        "properties": {
          "page":  { "type": "integer" },
          "limit": { "type": "integer" },
          "total": { "type": "integer" }
        }
      },
      "ApplicationListResponse": {
        "type": "object",
        "properties": {
          "applications": { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationItem" } },
          "pagination":   { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ApplicationAttachment": {
        "type": "object",
        "properties": {
          "id":        { "type": "string", "format": "uuid" },
          "file_name": { "type": "string" },
          "file_url":  { "type": "string" },
          "file_type": { "type": "string", "example": "application/pdf" },
          "file_size": { "type": "integer", "nullable": true }
        }
      },
      "ApplicationDetail": {
        "type": "object",
        "properties": {
          "id":                { "type": "string", "format": "uuid" },
          "application_code":  { "type": "string", "example": "APP-20260520-ABCDEF" },
          "service_type_id":   { "type": "string", "format": "uuid" },
          "service_type_name": { "type": "string" },
          "status":            { "type": "string", "enum": ["received", "processing", "need_more_info", "approved", "rejected"] },
          "submitted_data":    { "type": "object", "additionalProperties": true },
          "submitted_at":      { "type": "string", "format": "date-time" },
          "attachments":       { "type": "array", "items": { "$ref": "#/components/schemas/ApplicationAttachment" } }
        }
      },
      "ApplicationDetailResponse": {
        "type": "object",
        "properties": {
          "application": { "$ref": "#/components/schemas/ApplicationDetail" }
        }
      },
      "ServiceType": {
        "type": "object",
        "properties": {
          "id":          { "type": "string", "format": "uuid" },
          "name":        { "type": "string" },
          "code":        { "type": "string" },
          "description": { "type": "string" },
          "category":    { "type": "string" },
          "is_active":   { "type": "boolean" },
          "created_at":  { "type": "string", "format": "date-time" }
        }
      },
      "ServiceListResponse": {
        "type": "object",
        "properties": {
          "services":   { "type": "array", "items": { "$ref": "#/components/schemas/ServiceType" } },
          "pagination": { "$ref": "#/components/schemas/Pagination" }
        }
      },
      "ServiceDetailResponse": {
        "type": "object",
        "properties": {
          "service": { "$ref": "#/components/schemas/ServiceType" }
        }
      },
      "ErrorResponse": {
        "type": "object",
        "properties": {
          "code": {
            "type": "integer",
            "example": 400
          },
          "errors": {
            "type": "array",
            "items": {
              "oneOf": [
                {
                  "type": "string"
                },
                {
                  "type": "object",
                  "additionalProperties": {
                    "type": "string"
                  }
                }
              ]
            }
          }
        }
      }
    },
    "securitySchemes": {
      "BearerAuth": {
        "type": "http",
        "scheme": "bearer",
        "bearerFormat": "JWT"
      }
    }
  }
}`

const swaggerUI = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Service Management System API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: "/swagger/doc.json",
        dom_id: "#swagger-ui",
        presets: [SwaggerUIBundle.presets.apis],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`

func SetupSwaggerRoutes(e *echo.Echo) {
	e.GET("/swagger", swaggerIndex)
	e.GET("/swagger/", swaggerIndex)
	e.GET("/swagger/doc.json", swaggerSpec)
}

func swaggerIndex(c *echo.Context) error {
	return c.HTML(http.StatusOK, swaggerUI)
}

func swaggerSpec(c *echo.Context) error {
	return c.Blob(http.StatusOK, "application/json", []byte(openAPISpec))
}
