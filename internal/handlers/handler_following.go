package handlers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/lmilojevicc/gator/internal/cli"
	"github.com/lmilojevicc/gator/internal/database"
	"github.com/lmilojevicc/gator/internal/state"
)

func HandlerFollow(s *state.State, cmd cli.Command, dbUser database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <url>", cmd.Name)
	}

	feedURL := cmd.Arguments[0]

	dbFeed, err := s.Queries.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("getting feed: %w", err)
	}

	dbFeedFollow, err := s.Queries.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
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

func HandlerFollowing(s *state.State, cmd cli.Command, dbUser database.User) error {
	dbFeedsFollowed, err := s.Queries.GetFeedFollowsForUser(context.Background(), dbUser.ID)
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

func HandlerUnfollow(s *state.State, cmd cli.Command, dbUser database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("usage: %s <feed_name>", cmd.Name)
	}

	feedURL := cmd.Arguments[0]
	dbFeed, err := s.Queries.GetFeedByURL(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("getting feed by url: %w", err)
	}

	_, err = s.Queries.Unfollow(context.Background(), database.UnfollowParams{
		FeedID: dbFeed.ID,
		UserID: dbUser.ID,
	})
	if err == sql.ErrNoRows {
		return fmt.Errorf("no feed with url: %s", feedURL)
	}
	if err != nil {
		return fmt.Errorf("unfollowing feed: %w", err)
	}

	fmt.Printf("You have successfully unfollowed %q", dbFeed.Name)

	return nil
}
