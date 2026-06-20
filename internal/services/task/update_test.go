package task

import (
	"testing"
	"time"
)

func TestApplyUpdate(t *testing.T) {
	desc := "old desc"
	est := EstimateM
	assignee := int64(5)
	due := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	base := &Task{
		ID: 1, TeamID: 1, CreatedBy: 1,
		Title: "old", Status: StatusTodo, Priority: PriorityMedium,
		Description: &desc, Estimate: &est, AssigneeID: &assignee, DueDate: &due,
	}

	sameAsBase := func() UpdateTaskInput {
		return UpdateTaskInput{
			UpdatedBy: 2, Title: "old", Status: StatusTodo, Priority: PriorityMedium,
			Description: &desc, Estimate: &est, AssigneeID: &assignee, DueDate: &due,
		}
	}

	t.Run("no changes", func(t *testing.T) {
		_, entries := applyUpdate(base, sameAsBase())
		if len(entries) != 0 {
			t.Errorf("expected 0 entries, got %d", len(entries))
		}
	})

	t.Run("title changed", func(t *testing.T) {
		in := sameAsBase()
		in.Title = "new"
		_, entries := applyUpdate(base, in)
		if len(entries) != 1 || entries[0].Field != "title" {
			t.Errorf("expected 1 title entry, got %+v", entries)
		}
		if *entries[0].OldValue != "old" || *entries[0].NewValue != "new" {
			t.Errorf("unexpected values: %+v", entries[0])
		}
	})

	t.Run("description cleared", func(t *testing.T) {
		in := sameAsBase()
		in.Description = nil
		_, entries := applyUpdate(base, in)
		if len(entries) != 1 || entries[0].Field != "description" {
			t.Errorf("expected 1 description entry, got %+v", entries)
		}
		if entries[0].NewValue != nil {
			t.Error("expected nil new value for cleared description")
		}
	})

	t.Run("estimate changed", func(t *testing.T) {
		in := sameAsBase()
		newEst := EstimateXL
		in.Estimate = &newEst
		_, entries := applyUpdate(base, in)
		if len(entries) != 1 || entries[0].Field != "estimate" {
			t.Errorf("expected 1 estimate entry, got %+v", entries)
		}
	})

	t.Run("assignee cleared", func(t *testing.T) {
		in := sameAsBase()
		in.AssigneeID = nil
		_, entries := applyUpdate(base, in)
		if len(entries) != 1 || entries[0].Field != "assignee_id" {
			t.Errorf("expected 1 assignee_id entry, got %+v", entries)
		}
	})

	t.Run("multiple fields", func(t *testing.T) {
		in := sameAsBase()
		in.Title = "new"
		in.Status = StatusDone
		_, entries := applyUpdate(base, in)
		if len(entries) != 2 {
			t.Errorf("expected 2 entries, got %d", len(entries))
		}
	})
}
