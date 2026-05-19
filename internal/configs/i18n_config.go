package configs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v5"
)

const (
	DefaultLocale = "vi"
	LocaleKey     = "locale"
)

var messages = map[string]map[string]string{}

func LoadI18nMessages(localesDir string) error {
	entries, err := os.ReadDir(localesDir)
	if err != nil {
		return fmt.Errorf("read locales directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		locale := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		content, err := os.ReadFile(filepath.Join(localesDir, entry.Name()))
		if err != nil {
			return fmt.Errorf("read locale %s: %w", locale, err)
		}

		localeMessages := map[string]string{}
		if err := json.Unmarshal(content, &localeMessages); err != nil {
			return fmt.Errorf("parse locale %s: %w", locale, err)
		}

		messages[locale] = localeMessages
	}

	if _, ok := messages[DefaultLocale]; !ok {
		return fmt.Errorf("default locale %q is not loaded", DefaultLocale)
	}

	return nil
}

func NormalizeLocale(locale string) string {
	locale = strings.ToLower(strings.TrimSpace(locale))
	if locale == "" {
		return DefaultLocale
	}

	locale = strings.Split(locale, ",")[0]
	locale = strings.Split(locale, "-")[0]

	if _, ok := messages[locale]; ok {
		return locale
	}

	return DefaultLocale
}

func LocaleFromContext(c *echo.Context) string {
	locale, ok := c.Get(LocaleKey).(string)
	if !ok {
		return DefaultLocale
	}

	return NormalizeLocale(locale)
}

func T(c *echo.Context, key string, params map[string]string) string {
	locale := LocaleFromContext(c)
	template, ok := messages[locale][key]
	if !ok {
		template, ok = messages[DefaultLocale][key]
	}
	if !ok {
		return key
	}

	for name, value := range params {
		template = strings.ReplaceAll(template, "{"+name+"}", value)
	}

	return template
}
