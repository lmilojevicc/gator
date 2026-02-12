package handlers

import (
	"context"
	"fmt"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/rss"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerAggregate(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	url := cmd.Arguments[0]

	feed, err := rss.FetchFeed(context.Background(), url)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}

	fmt.Printf("%v", feed.Channel.Item)

	return nil
}
