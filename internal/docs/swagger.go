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
    {
      "name": "Auth",
      "description": "Authentication APIs"
    }
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
