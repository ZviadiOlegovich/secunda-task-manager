package task

import (
	"strconv"
	"time"
)

func applyUpdate(t *Task, update UpdateTaskInput) (*Task, []TaskHistoryEntry) {
	out := *t
	var entries []TaskHistoryEntry

	record := func(field string, old, new *string) {
		entries = append(entries, TaskHistoryEntry{
			TaskID: t.ID, UserID: update.UpdatedBy,
			Field: field, OldValue: old, NewValue: new,
		})
	}

	if update.Title != t.Title {
		old := t.Title
		record("title", &old, &update.Title)
		out.Title = update.Title
	}

	if update.Status != t.Status {
		old, new := string(t.Status), string(update.Status)
		record("status", &old, &new)
		out.Status = update.Status
	}

	if update.Priority != t.Priority {
		old, new := string(t.Priority), string(update.Priority)
		record("priority", &old, &new)
		out.Priority = update.Priority
	}

	if !ptrEqual(t.Description, update.Description) {
		record("description", t.Description, update.Description)
		out.Description = update.Description
	}

	if !ptrEqual(t.Estimate, update.Estimate) {
		record("estimate", enumPtrStr(t.Estimate), enumPtrStr(update.Estimate))
		out.Estimate = update.Estimate
	}

	if !ptrEqual(t.AssigneeID, update.AssigneeID) {
		record("assignee_id", int64PtrStr(t.AssigneeID), int64PtrStr(update.AssigneeID))
		out.AssigneeID = update.AssigneeID
	}

	if !timePtrEqual(t.DueDate, update.DueDate) {
		record("due_date", timePtrStr(t.DueDate), timePtrStr(update.DueDate))
		out.DueDate = update.DueDate
	}

	return &out, entries
}

func ptrEqual[T comparable](a, b *T) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func timePtrEqual(a, b *time.Time) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return a.Equal(*b)
}

func enumPtrStr[T ~string](v *T) *string {
	if v == nil {
		return nil
	}
	s := string(*v)
	return &s
}

func int64PtrStr(v *int64) *string {
	if v == nil {
		return nil
	}
	s := strconv.FormatInt(*v, 10)
	return &s
}

func timePtrStr(v *time.Time) *string {
	if v == nil {
		return nil
	}
	s := v.Format(time.RFC3339)
	return &s
}
