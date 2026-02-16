package middleware

import (
	"context"
	"fmt"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/state"
)

type authenticationHandler = func(s *state.State, cmd cli.Command, user database.User) error

func LoggedIn(handler authenticationHandler) func(*state.State, cli.Command) error {
	return func(s *state.State, cmd cli.Command) error {
		currentUser := s.Cfg.CurrentUserName
		if currentUser == "" {
			return fmt.Errorf("user must be logged in")
		}

		dbUser, err := s.Queries.GetUserByName(context.Background(), currentUser)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}

		return handler(s, cmd, dbUser)
	}
}
