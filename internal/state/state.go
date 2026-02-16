package state

import (
	"database/sql"

	"github.com/lmilojevicc/gator/internal/config"
	"github.com/lmilojevicc/gator/internal/database"
)

type State struct {
	Queries *database.Queries
	Cfg     *config.Config
	Conn    *sql.DB
}
