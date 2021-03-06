package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

type AuditLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	GithubID  string    `json:"github_id" db:"github_id"`
}

// String is not required by pop and may be deleted
func (a AuditLog) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// AuditLogs is not required by pop and may be deleted
type AuditLogs []AuditLog

// String is not required by pop and may be deleted
func (a AuditLogs) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate*" (pop.ValidateAndSave, pop.ValidateAndCreate, pop.ValidateAndUpdate) method.
// This method is not required and may be deleted.
func (a *AuditLog) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: a.GithubID, Name: "GithubID"},
	), nil
}
