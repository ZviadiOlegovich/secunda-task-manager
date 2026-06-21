//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/stats"
	"github.com/zoshc/secunda-task-manager/internal/services/task"
	"github.com/zoshc/secunda-task-manager/internal/services/team"
)

type StatsRepoSuite struct {
	baseSuite
	repo     stats.Repository
	teamRepo team.Repository
}

func TestStatsRepo(t *testing.T) {
	suite.Run(t, new(StatsRepoSuite))
}

func (s *StatsRepoSuite) SetupSuite() {
	s.baseSuite.SetupSuite()
	s.repo = repository.NewStatsRepository(s.db)
	s.teamRepo = repository.NewTeamRepository(s.db)
}

func (s *StatsRepoSuite) seedTask(teamID, createdBy int64, status task.TaskStatus, assigneeID *int64) {
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO tasks (team_id, title, status, priority, created_by, assignee_id) VALUES (?, ?, ?, ?, ?, ?)`,
		teamID, "task", status, task.PriorityMedium, createdBy, assigneeID,
	)
	s.Require().NoError(err)
}

func (s *StatsRepoSuite) seedTaskAt(teamID, createdBy int64, status task.TaskStatus, updatedAt time.Time) {
	_, err := s.db.ExecContext(context.Background(),
		`INSERT INTO tasks (team_id, title, status, priority, created_by, updated_at) VALUES (?, ?, ?, ?, ?, ?)`,
		teamID, "task", status, task.PriorityMedium, createdBy, updatedAt,
	)
	s.Require().NoError(err)
}

func (s *StatsRepoSuite) TestTeamStats() {
	ctx := context.Background()

	ownerA := seedUser(s.T(), s.db)
	memberA1 := seedUser(s.T(), s.db)
	memberA2 := seedUser(s.T(), s.db)
	teamA, err := s.teamRepo.CreateWithOwner(ctx, &team.Team{Name: "Team A", CreatedBy: ownerA}, ownerA)
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamA, UserID: memberA1, Role: team.RoleMember}))
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamA, UserID: memberA2, Role: team.RoleMember}))

	s.seedTask(teamA, ownerA, task.StatusDone, nil)
	s.seedTask(teamA, memberA1, task.StatusDone, nil)
	s.seedTask(teamA, memberA2, task.StatusDone, nil)
	s.seedTask(teamA, ownerA, task.StatusTodo, nil)
	s.seedTask(teamA, ownerA, task.StatusTodo, nil)
	s.seedTaskAt(teamA, ownerA, task.StatusDone, time.Now().AddDate(0, 0, -10))

	ownerB := seedUser(s.T(), s.db)
	memberB := seedUser(s.T(), s.db)
	teamB, err := s.teamRepo.CreateWithOwner(ctx, &team.Team{Name: "Team B", CreatedBy: ownerB}, ownerB)
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamB, UserID: memberB, Role: team.RoleMember}))
	s.seedTask(teamB, ownerB, task.StatusTodo, nil)
	s.seedTask(teamB, ownerB, task.StatusInProgress, nil)

	result, err := s.repo.TeamStats(ctx)
	s.Require().NoError(err)

	statA := s.findTeamStat(result, teamA)
	s.Require().NotNil(statA, "team A not found")
	s.Equal(3, statA.MemberCount, "team A: member_count")
	s.Equal(3, statA.DoneLastWeek, "team A: done_last_week (old task excluded)")

	statB := s.findTeamStat(result, teamB)
	s.Require().NotNil(statB, "team B not found")
	s.Equal(2, statB.MemberCount, "team B: member_count")
	s.Equal(0, statB.DoneLastWeek, "team B: done_last_week")
}

func (s *StatsRepoSuite) TestTopUsers() {
	ctx := context.Background()

	u1 := seedUser(s.T(), s.db)
	u2 := seedUser(s.T(), s.db)
	u3 := seedUser(s.T(), s.db)
	u4 := seedUser(s.T(), s.db)

	teamID, err := s.teamRepo.CreateWithOwner(ctx, &team.Team{Name: "Top Team", CreatedBy: u1}, u1)
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: u2, Role: team.RoleMember}))
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: u3, Role: team.RoleMember}))
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: u4, Role: team.RoleMember}))

	for i := 0; i < 5; i++ {
		s.seedTask(teamID, u1, task.StatusTodo, nil)
	}
	for i := 0; i < 3; i++ {
		s.seedTask(teamID, u2, task.StatusTodo, nil)
	}
	s.seedTask(teamID, u3, task.StatusTodo, nil)

	result, err := s.repo.TopUsers(ctx)
	s.Require().NoError(err)

	top := s.filterTopUsers(result, teamID)
	s.Require().Len(top, 3, "expected top-3 for team")

	s.Equal(u1, top[0].UserID, "rank1 — u1")
	s.Equal(1, top[0].Rank)
	s.Equal(5, top[0].TaskCount)

	s.Equal(u2, top[1].UserID, "rank2 — u2")
	s.Equal(2, top[1].Rank)
	s.Equal(3, top[1].TaskCount)

	s.Equal(u3, top[2].UserID, "rank3 — u3")
	s.Equal(3, top[2].Rank)
	s.Equal(1, top[2].TaskCount)
}

func (s *StatsRepoSuite) TestTopUsers_Tie() {
	ctx := context.Background()

	u1 := seedUser(s.T(), s.db)
	u2 := seedUser(s.T(), s.db)
	u3 := seedUser(s.T(), s.db)

	teamID, err := s.teamRepo.CreateWithOwner(ctx, &team.Team{Name: "Tie Team", CreatedBy: u1}, u1)
	s.Require().NoError(err)
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: u2, Role: team.RoleMember}))
	s.Require().NoError(s.teamRepo.AddMember(ctx, &team.TeamMember{TeamID: teamID, UserID: u3, Role: team.RoleMember}))

	for _, uid := range []int64{u1, u2, u3} {
		s.seedTask(teamID, uid, task.StatusTodo, nil)
		s.seedTask(teamID, uid, task.StatusTodo, nil)
	}

	result, err := s.repo.TopUsers(ctx)
	s.Require().NoError(err)

	top := s.filterTopUsers(result, teamID)
	s.Require().Len(top, 3, "all three tied users must appear")
	for _, u := range top {
		s.Equal(1, u.Rank, "tie → all rank=1")
		s.Equal(2, u.TaskCount)
	}
}

func (s *StatsRepoSuite) TestTasksWithInvalidAssignee() {
	ctx := context.Background()

	member := seedUser(s.T(), s.db)
	outsider1 := seedUser(s.T(), s.db)
	outsider2 := seedUser(s.T(), s.db)

	teamID, err := s.teamRepo.CreateWithOwner(ctx, &team.Team{Name: "Invalid Team", CreatedBy: member}, member)
	s.Require().NoError(err)

	s.seedTask(teamID, member, task.StatusTodo, &member)
	s.seedTask(teamID, member, task.StatusInProgress, &member)

	s.seedTask(teamID, member, task.StatusTodo, &outsider1)
	s.seedTask(teamID, member, task.StatusTodo, &outsider1)
	s.seedTask(teamID, member, task.StatusDone, &outsider2)

	result, err := s.repo.TasksWithInvalidAssignee(ctx)
	s.Require().NoError(err)

	invalid := s.filterInvalidByTeam(result, teamID)
	s.Len(invalid, 3, "expected exactly 3 tasks with invalid assignee")

	count := make(map[int64]int)
	for _, t := range invalid {
		count[t.AssigneeID]++
	}
	s.Equal(2, count[outsider1], "outsider1 assigned to 2 tasks")
	s.Equal(1, count[outsider2], "outsider2 assigned to 1 task")
}

func (s *StatsRepoSuite) findTeamStat(result []stats.TeamStat, teamID int64) *stats.TeamStat {
	for i := range result {
		if result[i].TeamID == teamID {
			return &result[i]
		}
	}
	return nil
}

func (s *StatsRepoSuite) filterTopUsers(result []stats.TopUser, teamID int64) []stats.TopUser {
	var out []stats.TopUser
	for _, u := range result {
		if u.TeamID == teamID {
			out = append(out, u)
		}
	}
	return out
}

func (s *StatsRepoSuite) filterInvalidByTeam(result []stats.TaskWithInvalidAssignee, teamID int64) []stats.TaskWithInvalidAssignee {
	var out []stats.TaskWithInvalidAssignee
	for _, t := range result {
		if t.TeamID == teamID {
			out = append(out, t)
		}
	}
	return out
}
