# Review: internal/services/application_service.go

- **Reviewed at:** 2026-05-21T00:00:00Z
- **Reviewer:** Go Echo Code Reviewer agent
- **File:** [internal/services/application_service.go](internal/services/application_service.go#L1-L400)

## Summary

`ApplicationService` implements submission, listing, retrieval, status history, and supplement uploads for citizen applications. Overall structure follows repository/service/handler separation and uses repositories and storage abstractions. Core flows (validation, storage, DB create, notification) are present.

## Checklist (selected items)

- [x] Role-based access: Handlers use service methods; ensure routes apply `JWTMiddleware` + `RequireRoles` (routes.go shows citizen group guarded).
- [x] Application statuses: Initial status set to `ApplicationStatusReceived` on submit.
- [x] Attachment constraints: Per-file and total size checks implemented; max count enforced.
- [~] Activity logging: Not explicit in this service (assumed to be in `CreateWithAttachments`), please verify repository implementation.
- [x] Password hashing: Not applicable to this file.
- [x] Transactional integrity: `CreateWithAttachments` likely handles transaction; code removes temp storage on error.
- [!] Input validation: `validateSubmittedData` validates required fields but treats unparseable schema as skip — confirm intentional.
- [x] Error mapping: Service returns domain errors (e.g., `ErrApplicationNotFound`) for handlers to map.
- [~] Test patterns: No unit tests found for this specific file in repo scan; ensure service tests cover error cases and storage failures.

## Findings / Issues

1) validateSubmittedData: loose handling of schema and value emptiness
   - file: [internal/services/application_service.go](internal/services/application_service.go#L263-L307)
   - details: If `schema` JSON is unparseable the function returns `nil` (skip validation). This may hide schema errors. Also the check `v == ""` compares interface{} to empty string — for non-string types (numbers, booleans, arrays) this may not behave as intended.
   - suggestion: Return an error when schema is invalid or be explicit that invalid schema means "no validation". For required-field emptiness, assert type string before comparing to `""` or use reflection to treat zero values consistently.

2) Hardcoded locale in notifications
   - file: [internal/services/application_service.go](internal/services/application_service.go#L99-L106, L141-L156)
   - details: Calls to `configs.TLang("vi", ...)` hardcode Vietnamese locale for notification titles and email bodies. If multi-locale support is expected, consider using user locale.
   - suggestion: Accept a locale parameter or lookup user's preferred locale from `userRepo.FindByID` before building messages.

3) Potential double use of application code generator
   - file: [internal/services/application_service.go](internal/services/application_service.go#L58-L76, L127-L131)
   - details: `app.ApplicationCode` is set using `utils.GenerateApplicationCode()` then `CreateWithAttachments` is passed `utils.GenerateApplicationCode` (a function). Verify `CreateWithAttachments` uses the generator only when needed; avoid inconsistent code generation.
   - suggestion: Either set code before create and pass nil, or pass generator and let repository assign one. Document expected behavior.

4) validateSubmittedData returns `ErrMissingRequiredField` when `json.Unmarshal(data, &submitted)` fails — ambiguous error
   - file: [internal/services/application_service.go](internal/services/application_service.go#L285-L293)
   - details: If submitted data is malformed JSON, returning `ErrMissingRequiredField` may be misleading.
   - suggestion: Return a clearer error like `ErrInvalidSubmittedData` or wrap with context.

5) Logging and error visibility in goroutine
   - file: [internal/services/application_service.go](internal/services/application_service.go#L135-L160)
   - details: Goroutine sends confirmation email and logs failure via `log.Printf`. Consider structured logging and surfacing failures (e.g., metrics) for observability.
   - suggestion: Use project logger and include application code in log message.

6) Resource cleanup on supplemental upload failure
   - file: [internal/services/application_service.go](internal/services/application_service.go#L180-L226)
   - details: For supplemental uploads, if `SaveApplicationFile` fails for some files earlier saved, there is no cleanup of already saved files.
   - suggestion: On failure, remove newly uploaded supplement files or rely on background cleanup process.

7) Tests missing for edge cases
   - details: Ensure unit tests cover:
     - storage failures (SaveApplicationFile returning ErrDisallowedMime and other errors)
     - repository create failures and cleanup behavior
     - validateSubmittedData edge cases (schema parse error, non-string fields)

## Suggested Code Changes (examples)

- Improve `validateSubmittedData`:

```go
// after unmarshal of submitted
for _, field := range s.Required {
    v, ok := submitted[field]
    if !ok || v == nil {
        return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
    }
    // if string, check empty string
    if str, ok := v.(string); ok && strings.TrimSpace(str) == "" {
        return fmt.Errorf("%w: %s", ErrMissingRequiredField, field)
    }
}
```

- Use user locale when building messages:

```go
user, _ := s.userRepo.FindByID(citizenUserID)
loc := "vi"
if user != nil && user.Locale != "" {
    loc = user.Locale
}
Title: configs.TLang(loc, ...)
```

- On supplement save failure, track saved urls and remove them on error using `s.storage.RemoveApplicationFiles(app.ID, savedPaths)` if available.

## Conclusion

`ApplicationService` is well-structured and implements required flows. Addressing the above findings will improve robustness (validation, cleanup), observability (structured logs), and internationalization.

---

*End of report.*
