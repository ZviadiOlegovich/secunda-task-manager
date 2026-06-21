//go:build integration

package integration

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
)

type teamFullRepo interface {
	team.Repository
	task.TeamRepository
}

type TeamRepoSuite struct {
	baseSuite
	repo teamFullRepo
}

func TestTeamRepo(t *testing.T) {
	suite.Run(t, new(TeamRepoSuite))
}

func (s *TeamRepoSuite) SetupSuite() {
	s.baseSuite.SetupSuite()
	s.repo = repository.NewTeamRepository(s.db)
}

func (s *TeamRepoSuite) TestCreateWithOwner() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)

	id, err := s.repo.CreateWithOwner(ctx, &team.Team{Name: "Alpha", CreatedBy: userID}, userID)
	s.Require().NoError(err)
	s.Greater(id, int64(0))

	m, err := s.repo.GetMember(ctx, id, userID)
	s.Require().NoError(err)
	s.Equal(team.RoleOwner, m.Role)
}

func (s *TeamRepoSuite) TestGetByID() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, userID)

	got, err := s.repo.GetByID(ctx, teamID)
	s.Require().NoError(err)
	s.Equal(teamID, got.ID)
	s.Equal("Test Team", got.Name)
}

func (s *TeamRepoSuite) TestGetByID_NotFound() {
	_, err := s.repo.GetByID(context.Background(), 999999)
	s.True(errors.Is(err, errs.ErrNotFound), "want ErrNotFound, got %v", err)
}

func (s *TeamRepoSuite) TestAddMember() {
	ctx := context.Background()
	ownerID := seedUser(s.T(), s.db)
	memberID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, ownerID)

	err := s.repo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: memberID, Role: team.RoleMember})
	s.Require().NoError(err)

	m, err := s.repo.GetMember(ctx, teamID, memberID)
	s.Require().NoError(err)
	s.Equal(team.RoleMember, m.Role)
}

func (s *TeamRepoSuite) TestAddMember_AlreadyMember() {
	ctx := context.Background()
	ownerID := seedUser(s.T(), s.db)
	memberID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, ownerID)

	s.Require().NoError(s.repo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: memberID, Role: team.RoleMember}))

	err := s.repo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: memberID, Role: team.RoleMember})
	s.True(errors.Is(err, team.ErrAlreadyMember), "want ErrAlreadyMember, got %v", err)
}

func (s *TeamRepoSuite) TestGetMember() {
	ctx := context.Background()
	ownerID := seedUser(s.T(), s.db)
	memberID := seedUser(s.T(), s.db)
	teamID := seedTeam(s.T(), s.db, ownerID)
	s.Require().NoError(s.repo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: memberID, Role: team.RoleAdmin}))

	m, err := s.repo.GetMember(ctx, teamID, memberID)
	s.Require().NoError(err)
	s.Equal(teamID, m.TeamID)
	s.Equal(memberID, m.UserID)
	s.Equal(team.RoleAdmin, m.Role)
}

func (s *TeamRepoSuite) TestGetByUserID() {
	ctx := context.Background()
	userID := seedUser(s.T(), s.db)
	otherID := seedUser(s.T(), s.db)

	_, err := s.repo.CreateWithOwner(ctx, &team.Team{Name: "Team 1", CreatedBy: userID}, userID)
	s.Require().NoError(err)
	_, err = s.repo.CreateWithOwner(ctx, &team.Team{Name: "Team 2", CreatedBy: userID}, userID)
	s.Require().NoError(err)
	_, err = s.repo.CreateWithOwner(ctx, &team.Team{Name: "Other Team", CreatedBy: otherID}, otherID)
	s.Require().NoError(err)

	teams, err := s.repo.GetByUserID(ctx, userID)
	s.Require().NoError(err)
	s.Len(teams, 2)
}

func (s *TeamRepoSuite) TestAreMembersOf() {
	ctx := context.Background()
	ownerID := seedUser(s.T(), s.db)
	memberID := seedUser(s.T(), s.db)
	outsiderID := seedUser(s.T(), s.db)

	teamID, err := s.repo.CreateWithOwner(ctx, &team.Team{Name: "Beta", CreatedBy: ownerID}, ownerID)
	s.Require().NoError(err)
	s.Require().NoError(s.repo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: memberID, Role: team.RoleMember}))

	s.Run("all members", func() {
		err := s.repo.AreMembersOf(ctx, teamID, []int64{ownerID, memberID})
		s.NoError(err)
	})

	s.Run("one not member", func() {
		err := s.repo.AreMembersOf(ctx, teamID, []int64{ownerID, outsiderID})
		s.True(errors.Is(err, errs.ErrNotFound), "want ErrNotFound, got %v", err)
	})
}
