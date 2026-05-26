package configs

import (
	"errors"
	"net/http"
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestCustomValidator_ValidStruct(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Email string `validate:"required,email"`
		Name  string `validate:"required,min=2"`
	}
	input := TestInput{Email: "a@b.com", Name: "An"}
	if err := cv.Validate(input); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestCustomValidator_RequiredField(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Email string `validate:"required"`
	}
	input := TestInput{}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T: %v", err, err)
	}
	if ve.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", ve.Code)
	}
	if len(ve.Messages) == 0 {
		t.Fatal("expected at least one message")
	}
	if ve.Messages[0].Key != "validation.required" {
		t.Fatalf("expected validation.required, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_EmailTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Email string `validate:"required,email"`
	}
	input := TestInput{Email: "not-an-email"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.email" {
		t.Fatalf("expected validation.email, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_MinTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Name string `validate:"min=5"`
	}
	input := TestInput{Name: "An"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.min" {
		t.Fatalf("expected validation.min, got %q", ve.Messages[0].Key)
	}
	if ve.Messages[0].Params["param"] != "5" {
		t.Fatalf("expected param 5, got %q", ve.Messages[0].Params["param"])
	}
}

func TestCustomValidator_MaxTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Name string `validate:"max=2"`
	}
	input := TestInput{Name: "TooLongName"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.max" {
		t.Fatalf("expected validation.max, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_LenTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Code string `validate:"len=6"`
	}
	input := TestInput{Code: "abc"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.len" {
		t.Fatalf("expected validation.len, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_NumericTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Code string `validate:"numeric"`
	}
	input := TestInput{Code: "abc123"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.numeric" {
		t.Fatalf("expected validation.numeric, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_OneofTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		Status string `validate:"oneof=active inactive"`
	}
	input := TestInput{Status: "unknown"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.invalid" {
		t.Fatalf("expected validation.invalid, got %q", ve.Messages[0].Key)
	}
}

func TestCustomValidator_DefaultTag(t *testing.T) {
	cv := &CustomValidator{Validator: validator.New()}

	type TestInput struct {
		URL string `validate:"url"`
	}
	input := TestInput{URL: "not-a-url"}
	err := cv.Validate(input)
	var ve *ValidatorError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidatorError, got %T", err)
	}
	if ve.Messages[0].Key != "validation.invalid" {
		t.Fatalf("expected validation.invalid for unknown tag, got %q", ve.Messages[0].Key)
	}
}

func TestValidatorError_Error(t *testing.T) {
	ve := &ValidatorError{Code: 400}
	if ve.Error() != "validation error" {
		t.Fatalf("expected 'validation error', got %q", ve.Error())
	}
}
