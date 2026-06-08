// Package versioning contains a simple status-transition model and helpers for
// versioned entities (entity → versions → active version pointer).
//
// For more complex workflows, prefer a database-driven workflow engine over the
// constants here.
package versioning

import "fmt"

type Status string

const (
	StatusDraft           Status = "DRAFT"
	StatusPendingApproval Status = "PENDING_APPROVAL"
	StatusApproved        Status = "APPROVED"
	StatusRejected        Status = "REJECTED"
)

func (s Status) String() string { return string(s) }

func (s Status) IsValid() bool {
	switch s {
	case StatusDraft, StatusPendingApproval, StatusApproved, StatusRejected:
		return true
	}
	return false
}

var allowedTransitions = map[Status][]Status{
	StatusDraft:           {StatusPendingApproval, StatusApproved},
	StatusPendingApproval: {StatusApproved, StatusRejected},
	StatusRejected:        {StatusDraft},
}

func CanTransition(from, to Status) bool {
	targets, ok := allowedTransitions[from]
	if !ok {
		return false
	}
	for _, t := range targets {
		if t == to {
			return true
		}
	}
	return false
}

func ValidateTransition(from, to Status) error {
	if !from.IsValid() {
		return fmt.Errorf("invalid source status: %s", from)
	}
	if !to.IsValid() {
		return fmt.Errorf("invalid target status: %s", to)
	}
	if !CanTransition(from, to) {
		return fmt.Errorf("transition %s → %s is not allowed", from, to)
	}
	return nil
}
