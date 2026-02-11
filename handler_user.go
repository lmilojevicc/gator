package main

import (
	"fmt"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("usage: %s <username>", cmd.Name)
	}

	username := cmd.Arguments[0]

	err := s.cfg.SetUser(username)
	if err != nil {
		return fmt.Errorf("set user: %w", err)
	}

	return nil
}
