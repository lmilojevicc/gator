package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}

	username := cmd.Arguments[0]
	dbUser, err := s.db.GetUserByName(context.Background(), username)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	err = s.cfg.SetUser(dbUser.Name)
	if err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}

	username := cmd.Arguments[0]

	dbUser, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:   uuid.New(),
		Name: username,
	})
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	err = s.cfg.SetUser(dbUser.Name)
	if err != nil {
		return fmt.Errorf("writing user to config: %w", err)
	}

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("reseting users: %w", err)
	}

	return nil
}

func handlerUsers(s *state, cmd command) error {
	dbUsers, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("getting users: %w", err)
	}

	currentUser := s.cfg.CurrentUserName
	for _, user := range dbUsers {
		if user.Name == currentUser {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}
