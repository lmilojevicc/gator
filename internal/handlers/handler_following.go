package handlers

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerFollow(s *state.State, cmd cli.Command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	feedURL := cmd.Arguments[0]

	dbUser, err := s.DB.GetUserByName(context.Background(), s.Cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	dbFeed, err := s.DB.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("getting feed: %w", err)
	}

	dbFeedFollow, err := s.DB.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:     uuid.New(),
		UserID: dbUser.ID,
		FeedID: dbFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("creating feed following: %w", err)
	}

	fmt.Printf("User %s is now following %q", dbFeedFollow.UserName, dbFeedFollow.FeedName)

	return nil
}

func HandlerFollowing(s *state.State, cmd cli.Command) error {
	dbUser, err := s.DB.GetUserByName(context.Background(), s.Cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("getting user: %w", err)
	}

	dbFeedsFollowed, err := s.DB.GetFeedFollowsForUser(context.Background(), dbUser.ID)
	if err != nil {
		return fmt.Errorf("getting feeds followed by user: %w", err)
	}

	if len(dbFeedsFollowed) == 0 {
		fmt.Println("You are currently not following any RSS feed")
		return nil
	}

	fmt.Printf("You are currently following:\n")
	for _, feed := range dbFeedsFollowed {
		fmt.Printf("  * %q\n", feed.FeedName)
	}

	return nil
}
