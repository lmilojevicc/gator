package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/rss"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerAggregate(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 1 {
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

func HandlerAddFeed(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}

	currentUser := s.Cfg.CurrentUserName
	dbUser, err := s.DB.GetUserByName(context.Background(), currentUser)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]

	err = s.DB.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   feedName,
		Url:    feedURL,
		UserID: dbUser.ID,
	})
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}

	return nil
}

func HandlerFeeds(s *state.State, cmd cli.Command) error {
	feeds, err := s.DB.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("getting feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* Name:\t%s\n", feed.Name)
		fmt.Printf("* URL:\t%s\n", feed.Url)
		user, err := s.DB.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}
		fmt.Printf("* User:\t%s\n", user.Name)
	}

	return nil
}
