package templates

import (
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
	"github.com/labstack/echo/v5"
)

type TemplateRenderer struct {
	templates *template.Template
}

func NewRenderer(root string) (*TemplateRenderer, error) {
	t := template.New("").Funcs(template.FuncMap{
		"t": func(key string, params ...map[string]string) string {
			return key
		},
		"currentLocale": func() string {
			return "vi"
		},
	})
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".html" {
			relPath, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			templateName := filepath.ToSlash(relPath)
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if _, err := t.New(templateName).Parse(strings.TrimSpace(string(content))); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &TemplateRenderer{templates: t}, nil
}

func (r *TemplateRenderer) Render(c *echo.Context, w io.Writer, name string, data any) error {
	funcMap := template.FuncMap{
		"t": func(key string, params ...map[string]string) string {
			paramMap := map[string]string{}
			if len(params) > 0 {
				paramMap = params[0]
			}
			return configs.T(c, key, paramMap)
		},
		"currentLocale": func() string {
			return configs.LocaleFromContext(c)
		},
	}

	cloned, err := r.templates.Funcs(funcMap).Clone()
	if err != nil {
		return err
	}

	return cloned.ExecuteTemplate(w, name, data)
}
