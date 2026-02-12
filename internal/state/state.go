package state

import (
	"github.com/lmilojevicc/gator/internal/config"
	"github.com/lmilojevicc/gator/internal/database"
)

type State struct {
	DB  *database.Queries
	Cfg *config.Config
}
