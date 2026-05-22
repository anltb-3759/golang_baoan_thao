package handlers

import (
	"os"
	"testing"

	"github.com/awesome-academy/golang_baoan_thao/internal/configs"
)

func TestMain(m *testing.M) {
	_ = configs.LoadI18nMessages("../../locales")
	os.Exit(m.Run())
}
