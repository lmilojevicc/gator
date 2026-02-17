package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
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
	defer ticker.Stop()
	for ; ; <-ticker.C {
		if err := scrapeFeeds(s); err != nil {
			fmt.Fprintf(os.Stderr, "Error scraping feeds: %v\n", err)
		}
	}
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
		publishedAt, err := parsePubDate(item.PubDate)
		if err != nil {
			fmt.Printf("Warning: skipping item %q with bad date: %v\n", item.Title, err)
			continue
		}

		_, err = qtx.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			Title:       sql.NullString{String: item.Title, Valid: item.Title != ""},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: sql.NullTime{Time: publishedAt, Valid: !publishedAt.IsZero()},
			FeedID:      nextFeedToFetch.ID,
		})

		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			return fmt.Errorf("creating post: %w", err)
		}
	}

	return tx.Commit()
}

func parsePubDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,  // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z, // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC3339,  // ISO 8601: "2006-01-02T15:04:05Z07:00"
		"2006-01-02T15:04:05Z",
		"Mon, 2 Jan 2006 15:04:05 MST",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func HandlerBrowse(s *state.State, cmd cli.Command, user database.User) error {
	limit := int32(2)

	if len(cmd.Arguments) == 1 {
		parsed, err := strconv.Atoi(cmd.Arguments[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsed)
	}

	dbPosts, err := s.Queries.GetPostsByUser(context.Background(), database.GetPostsByUserParams{
		UserID: user.ID,
		Limit:  limit,
	})
	if err != nil {
		return fmt.Errorf("getting posts: %w", err)
	}

	for _, post := range dbPosts {
		var date string
		if post.PublishedAt.Valid {
			date = post.PublishedAt.Time.Format("2006-01-02")
		} else {
			date = "unknown"
		}
		fmt.Printf("%q posted on %s\n", post.Title.String, date)
		fmt.Printf("Read at: %s\n\n", post.Url)
	}

	return nil
}
