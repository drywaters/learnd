package ui

import (
	"github.com/drywaters/learnd/internal/model"
)

// EntryView decorates an entry with UI-only fields.
type EntryView struct {
	model.Entry
	DuplicateCount int
	SwapOOB        bool
}
