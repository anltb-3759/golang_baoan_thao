package utils

import "os"

// SetIfNotNil assigns *src to *dst only when src is non-nil.
func SetIfNotNil[T any](dst *T, src *T) {
	if src != nil {
		*dst = *src
	}
}

// EnvOr returns the value of the environment variable key,
// or fallback if the variable is unset or empty.
func EnvOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
