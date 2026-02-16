package cli

import (
	"fmt"

	"github.com/lmilojevicc/gator/internal/state"
)

type Command struct {
	Name      string
	Arguments []string
}

type Commands struct {
	RegisteredCommands map[string]func(*state.State, Command) error
}

func (c Commands) Run(s *state.State, cmd Command) error {
	f, ok := c.RegisteredCommands[cmd.Name]
	if !ok {
		return fmt.Errorf("%s command not found", cmd.Name)
	}
	return f(s, cmd)
}

func (c Commands) Register(name string, f func(*state.State, Command) error) {
	c.RegisteredCommands[name] = f
}
