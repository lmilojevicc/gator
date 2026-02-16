package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/config"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/handlers"
	"github.com/lmilojevicc/gator/internal/middleware"
	"github.com/lmilojevicc/gator/internal/state"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Failed to read config: %v\n", err)
	}

	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("failed connecting to db: %v", err)
	}

	dbQueries := database.New(db)

	programState := state.State{
		Cfg:     &cfg,
		Queries: dbQueries,
		Conn:    db,
	}

	cmds := cli.Commands{
		RegisteredCommands: make(map[string]func(*state.State, cli.Command) error),
	}

	cmds.Register("login", handlers.HandlerLogin)
	cmds.Register("register", handlers.HandlerRegister)
	cmds.Register("reset", handlers.HandlerReset)
	cmds.Register("users", handlers.HandlerUsers)
	cmds.Register("agg", handlers.HandlerAggregate)
	cmds.Register("addfeed", middleware.LoggedIn(handlers.HandlerAddFeed))
	cmds.Register("feeds", handlers.HandlerFeeds)
	cmds.Register("follow", middleware.LoggedIn(handlers.HandlerFollow))
	cmds.Register("following", middleware.LoggedIn(handlers.HandlerFollowing))
	cmds.Register("unfollow", middleware.LoggedIn(handlers.HandlerUnfollow))
	cmds.Register("browse", middleware.LoggedIn(handlers.HandlerBrowse))

	if len(os.Args) < 2 {
		log.Fatalf("Usage: cli <command> [args...]")
	}

	cmdName := os.Args[1]
	cmdArgs := os.Args[2:]

	err = cmds.Run(&programState, cli.Command{Name: cmdName, Arguments: cmdArgs})
	if err != nil {
		log.Fatal(err)
	}
}
