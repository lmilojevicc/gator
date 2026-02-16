package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/rss"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerAggregate(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <time_between_reqs>", cmd.Name)
	}

	timeArg := cmd.Arguments[0]

	interval, err := time.ParseDuration(timeArg)
	if err != nil {
		return fmt.Errorf("invalid duration format (use 2h, 2m, 2s, etc...): %w", err)
	}

	ticker := time.NewTicker(interval)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	return nil
}

func HandlerAddFeed(s *state.State, cmd cli.Command, dbUser database.User) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf("usage: %s <name> <url>", cmd.Name)
	}

	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]

	createdFeed, err := s.Queries.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:     uuid.New(),
		Name:   feedName,
		Url:    feedURL,
		UserID: dbUser.ID,
	})
	if err != nil {
		return fmt.Errorf("creating feed: %w", err)
	}

	_, err = s.Queries.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: dbUser.ID,
		FeedID: createdFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("creating follow for created feed: %w", err)
	}

	return nil
}

func HandlerFeeds(s *state.State, cmd cli.Command) error {
	feeds, err := s.Queries.GetAllFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("getting feeds: %w", err)
	}

	for _, feed := range feeds {
		fmt.Printf("* Name:\t%s\n", feed.Name)
		fmt.Printf("* URL:\t%s\n", feed.Url)
		user, err := s.Queries.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("getting user: %w", err)
		}
		fmt.Printf("* User:\t%s\n", user.Name)
	}

	return nil
}

func scrapeFeeds(s *state.State) error {
	tx, err := s.Conn.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := s.Queries.WithTx(tx)

	nextFeedToFetch, err := qtx.GetNextFeedToFetch(context.Background())
	if err == sql.ErrNoRows {
		return fmt.Errorf("no feeds to fetch: %w", err)
	}
	if err != nil {
		return fmt.Errorf("getting next feed to fetch: %w", err)
	}

	err = qtx.MarkFeedFetched(context.Background(), nextFeedToFetch.ID)
	if err != nil {
		return fmt.Errorf("marking feed fetched: %w", err)
	}

	feed, err := rss.FetchFeed(context.Background(), nextFeedToFetch.Url)
	if err != nil {
		return fmt.Errorf("fetching feed: %w", err)
	}

	for _, item := range feed.Channel.Items {
		fmt.Printf("%s\n", item.Title)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commiting transaction: %w", err)
	}

	return nil
}
