package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerLogin(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}

	username := cmd.Arguments[0]
	dbUser, err := s.DB.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	err = s.Cfg.SetUser(dbUser.Name)
	if err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	return nil
}

func HandlerRegister(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}

	username := cmd.Arguments[0]

	dbUser, err := s.DB.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: username,
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	err = s.Cfg.SetUser(dbUser.Name)
	if err != nil {
		return fmt.Errorf("writing user to config: %w", err)
	}

	return nil
}

func HandlerReset(s *state.State, cmd cli.Command) error {
	err := s.DB.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("reseting users: %w", err)
	}

	return nil
}

func HandlerUsers(s *state.State, cmd cli.Command) error {
	dbUsers, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("getting users: %w", err)
	}

	currentUser := s.Cfg.CurrentUserName
	for _, user := range dbUsers {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}
