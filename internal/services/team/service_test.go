package team

import (
	"context"
	"errors"
	"testing"

	"github.com/zoshc/secunda-task-manager/internal/services/errs"
)

type mockRepo struct {
	createWithOwnerFn func(ctx context.Context, t *Team, ownerID int64) (int64, error)
	addMemberFn       func(ctx context.Context, m *TeamMember) error
	getByIDFn         func(ctx context.Context, id int64) (*Team, error)
	getByUserIDFn     func(ctx context.Context, userID int64) ([]*Team, error)
	getMemberFn       func(ctx context.Context, teamID, userID int64) (*TeamMember, error)
}

type mockEmailSvc struct{}

func (m *mockEmailSvc) SendInvite(_ context.Context, _, _ string) error { return nil }

var okEmail = &mockEmailSvc{}
var okTeamByID = func(_ context.Context, _ int64) (*Team, error) {
	return &Team{ID: 1, Name: "Alpha"}, nil
}

func (m *mockRepo) CreateWithOwner(ctx context.Context, t *Team, ownerID int64) (int64, error) {
	return m.createWithOwnerFn(ctx, t, ownerID)
}

func (m *mockRepo) AddMember(ctx context.Context, member *TeamMember) error {
	return m.addMemberFn(ctx, member)
}

func (m *mockRepo) GetByID(ctx context.Context, id int64) (*Team, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockRepo) GetByUserID(ctx context.Context, userID int64) ([]*Team, error) {
	return m.getByUserIDFn(ctx, userID)
}

func (m *mockRepo) GetMember(ctx context.Context, teamID, userID int64) (*TeamMember, error) {
	return m.getMemberFn(ctx, teamID, userID)
}

func TestService_CreateTeam(t *testing.T) {
	okRepo := &mockRepo{
		createWithOwnerFn: func(_ context.Context, _ *Team, _ int64) (int64, error) { return 1, nil },
	}

	tests := []struct {
		name    string
		input   string
		repo    Repository
		wantErr error
	}{
		{
			name:  "success",
			input: "Alpha",
			repo:  okRepo,
		},
		{
			name:    "empty name",
			input:   "",
			repo:    okRepo,
			wantErr: ErrInvalidName,
		},
		{
			name:    "blank name",
			input:   "   ",
			repo:    okRepo,
			wantErr: ErrInvalidName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, okEmail)
			id, err := svc.CreateTeam(context.Background(), 1, tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
			if tt.wantErr == nil && id <= 0 {
				t.Error("expected positive id, got 0")
			}
		})
	}
}

func TestService_InviteUser(t *testing.T) {
	member := func(role Role) *TeamMember {
		return &TeamMember{TeamID: 1, UserID: 10, Role: role}
	}
	input := func(role Role) InviteUserInput {
		return InviteUserInput{TeamID: 1, InviterID: 10, InviteeID: 20, Role: role}
	}

	tests := []struct {
		name    string
		input   InviteUserInput
		repo    Repository
		wantErr error
	}{
		{
			name:  "owner invites member",
			input: input(RoleMember),
			repo: &mockRepo{
				getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleOwner), nil },
				getByIDFn:   okTeamByID,
				addMemberFn: func(_ context.Context, _ *TeamMember) error { return nil },
			},
		},
		{
			name:  "owner invites admin",
			input: input(RoleAdmin),
			repo: &mockRepo{
				getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleOwner), nil },
				getByIDFn:   okTeamByID,
				addMemberFn: func(_ context.Context, _ *TeamMember) error { return nil },
			},
		},
		{
			name:  "admin invites member",
			input: input(RoleMember),
			repo: &mockRepo{
				getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleAdmin), nil },
				getByIDFn:   okTeamByID,
				addMemberFn: func(_ context.Context, _ *TeamMember) error { return nil },
			},
		},
		{
			name:    "admin invites admin — denied",
			input:   input(RoleAdmin),
			repo:    &mockRepo{getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleAdmin), nil }},
			wantErr: ErrPermissionDenied,
		},
		{
			name:    "member invites — denied",
			input:   input(RoleMember),
			repo:    &mockRepo{getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleMember), nil }},
			wantErr: ErrPermissionDenied,
		},
		{
			name:    "inviter not in team — denied",
			input:   input(RoleMember),
			repo:    &mockRepo{getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return nil, errs.ErrNotFound }},
			wantErr: ErrPermissionDenied,
		},
		{
			name:  "invitee already member",
			input: input(RoleMember),
			repo: &mockRepo{
				getMemberFn: func(_ context.Context, _, _ int64) (*TeamMember, error) { return member(RoleOwner), nil },
				getByIDFn:   okTeamByID,
				addMemberFn: func(_ context.Context, _ *TeamMember) error { return ErrAlreadyMember },
			},
			wantErr: ErrAlreadyMember,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := New(tt.repo, okEmail)
			err := svc.InviteUser(context.Background(), tt.input)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
