//go:build integration

package integration

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"github.com/zoshc/secunda-task-manager/internal/repository"
	"github.com/zoshc/secunda-task-manager/internal/services/errs"
	"github.com/zoshc/secunda-task-manager/internal/services/user"
)

type UserRepoSuite struct {
	baseSuite
	repo user.Repository
}

func TestUserRepo(t *testing.T) {
	suite.Run(t, new(UserRepoSuite))
}

func (s *UserRepoSuite) SetupSuite() {
	s.baseSuite.SetupSuite()
	s.repo = repository.NewUserRepository(s.db)
}

func (s *UserRepoSuite) TestCreate() {
	email := fmt.Sprintf("alice-%d@test.com", time.Now().UnixNano())
	err := s.repo.Create(context.Background(), &user.User{
		Email:        email,
		PasswordHash: "hash",
		Name:         "Alice",
	})
	s.NoError(err)
}

func (s *UserRepoSuite) TestCreate_EmailTaken() {
	email := fmt.Sprintf("bob-%d@test.com", time.Now().UnixNano())
	u := &user.User{Email: email, PasswordHash: "hash", Name: "Bob"}

	s.Require().NoError(s.repo.Create(context.Background(), u))

	err := s.repo.Create(context.Background(), u)
	s.True(errors.Is(err, user.ErrEmailTaken), "want ErrEmailTaken, got %v", err)
}

func (s *UserRepoSuite) TestGetByEmail() {
	ctx := context.Background()
	email := fmt.Sprintf("carol-%d@test.com", time.Now().UnixNano())
	s.Require().NoError(s.repo.Create(ctx, &user.User{Email: email, PasswordHash: "hash", Name: "Carol"}))

	got, err := s.repo.GetByEmail(ctx, email)
	s.Require().NoError(err)
	s.Equal(email, got.Email)
	s.Equal("Carol", got.Name)
	s.Greater(got.ID, int64(0))
}

func (s *UserRepoSuite) TestGetByEmail_NotFound() {
	_, err := s.repo.GetByEmail(context.Background(), "nobody@test.com")
	s.True(errors.Is(err, errs.ErrNotFound), "want ErrNotFound, got %v", err)
}

func (s *UserRepoSuite) TestUpdateRefreshToken() {
	ctx := context.Background()
	email := fmt.Sprintf("dave-%d@test.com", time.Now().UnixNano())
	s.Require().NoError(s.repo.Create(ctx, &user.User{Email: email, PasswordHash: "hash", Name: "Dave"}))

	u, err := s.repo.GetByEmail(ctx, email)
	s.Require().NoError(err)

	s.Require().NoError(s.repo.UpdateRefreshToken(ctx, u.ID, "my-token"))

	updated, err := s.repo.GetByEmail(ctx, email)
	s.Require().NoError(err)
	s.Require().NotNil(updated.RefreshToken)
	s.Equal("my-token", *updated.RefreshToken)
}

func (s *UserRepoSuite) TestGetByRefreshToken() {
	ctx := context.Background()
	email := fmt.Sprintf("eve-%d@test.com", time.Now().UnixNano())
	s.Require().NoError(s.repo.Create(ctx, &user.User{Email: email, PasswordHash: "hash", Name: "Eve"}))

	u, err := s.repo.GetByEmail(ctx, email)
	s.Require().NoError(err)
	s.Require().NoError(s.repo.UpdateRefreshToken(ctx, u.ID, "eve-token"))

	found, err := s.repo.GetByRefreshToken(ctx, "eve-token")
	s.Require().NoError(err)
	s.Equal(email, found.Email)
}

func (s *UserRepoSuite) TestGetByRefreshToken_NotFound() {
	_, err := s.repo.GetByRefreshToken(context.Background(), "nonexistent-token")
	s.True(errors.Is(err, errs.ErrNotFound), "want ErrNotFound, got %v", err)
}
