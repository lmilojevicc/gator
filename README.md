# Gator

A CLI-based RSS feed aggregator written in Go. Subscribe to your favorite RSS feeds, aggregate them, and browse posts from the command line.

## Features

- User registration and authentication with local config
- Add and manage RSS feeds
- Follow/unfollow feeds to curate your reading list
- Aggregate feeds on a configurable schedule
- Browse posts from feeds you follow
- Transaction-safe feed scraping with duplicate detection
- PostgreSQL backend with migrations

## Prerequisites

- [Go](https://golang.org/) 1.25.6 or later
- [PostgreSQL](https://www.postgresql.org/) 12 or later
- [mise](https://mise.jdx.dev/) (optional, for task running)
- [goose](https://github.com/pressly/goose) (for database migrations)
- [sqlc](https://sqlc.dev/) (for code generation)

## Installation

```bash
# Clone the repository
git clone <repository-url>
cd gator

# Build the binary
go build -o gator

# Or install directly
go install
```

## Configuration

Gator uses a JSON config file located at `~/.config/gator/config.json` (or `$XDG_CONFIG_HOME/gator/config.json`).

The config file stores:

- `db_url`: PostgreSQL connection string
- `current_user_name`: Currently logged-in user

### Environment Variables

Set your database URL via environment variable:

```bash
export POSTGRES_URL="postgres://user:password@localhost:5432/gator?sslmode=disable"
```

Or create a `.env` file:

```bash
POSTGRES_URL=postgres://localhost:5432/gator?sslmode=disable
```

## Database Setup

### Using Docker (recommended)

```bash
docker-compose up -d postgres
```

### Manual Setup

1. Create a PostgreSQL database:

```bash
createdb gator
```

2. Run migrations:

Using mise:

```bash
mise run goose:up
```

Or manually:

```bash
cd sql/schema
goose postgres "$POSTGRES_URL" up
```

## Usage

### User Management

```bash
# Register a new user
./gator register alice

# Login as existing user
./gator login alice

# List all users
./gator users

# Reset database (delete all users)
./gator reset
```

### Feed Management

```bash
# Add a new feed
./gator addfeed "Y Combinator" https://news.ycombinator.com/rss

# List all feeds in the database
./gator feeds

# Follow an existing feed
./gator follow https://news.ycombinator.com/rss

# List feeds you're following
./gator following

# Unfollow a feed
./gator unfollow https://news.ycombinator.com/rss
```

### Aggregating Feeds

Start the aggregator to fetch posts on a schedule:

```bash
# Fetch feeds every 1 hour
./gator agg 1h

# Fetch feeds every 30 minutes
./gator agg 30m

# Fetch feeds every 5 seconds (for testing)
./gator agg 5s
```

The aggregator will:

1. Fetch the next feed that hasn't been updated recently
2. Parse all posts from the feed
3. Store new posts in the database (duplicates are ignored)
4. Mark the feed as fetched
5. Repeat on the configured interval

### Browsing Posts

```bash
# Browse last 2 posts from followed feeds
./gator browse

# Browse last 10 posts
./gator browse 10

# Browse last 50 posts
./gator browse 50
```

## Project Structure

```
gator/
├── main.go                     # Application entry point
├── internal/
│   ├── handlers/              # CLI command handlers
│   │   ├── handler_rss.go     # Feed aggregation & browsing
│   │   ├── handler_following.go # Follow/unfollow commands
│   │   └── handler_user.go    # User management commands
│   ├── middleware/            # Authentication middleware
│   │   └── middleware.go      # LoggedIn middleware
│   ├── database/              # sqlc-generated code
│   │   ├── db.go             # Database connection
│   │   ├── models.go         # Data models
│   │   └── *.sql.go          # Generated query functions
│   ├── cli/                   # Command-line interface
│   │   └── commands.go       # Command registry
│   ├── state/                 # Application state
│   │   └── state.go          # State struct (DB, Config)
│   ├── config/                # Configuration management
│   │   └── config.go         # Config file handling
│   └── rss/                   # RSS feed fetching
│       └── rss.go            # HTTP client & XML parsing
├── sql/
│   ├── schema/               # Database migrations
│   │   ├── 001_user.sql
│   │   ├── 002_feeds.sql
│   │   ├── 003_feed_follow.sql
│   │   ├── 004_last_fetched_at.sql
│   │   └── 005_posts.sql
│   └── queries/              # SQL queries for sqlc
│       ├── users.sql
│       ├── feeds.sql
│       ├── follows.sql
│       └── posts.sql
├── docker-compose.yml        # PostgreSQL container
├── mise.toml                 # Task definitions
└── sqlc.yaml                 # sqlc configuration
```

## Technical Details

### Architecture

- **CLI Layer**: Simple command registry pattern with `internal/cli`
- **Handler Layer**: Command handlers separated by domain (user, feed, rss)
- **Middleware**: Authentication wrapper that injects current user
- **Database Layer**: sqlc generates type-safe Go code from SQL queries
- **State Management**: Centralized state struct passed to all handlers

### Database Schema

```mermaid
erDiagram
    users ||--o{ feeds : "creates"
    users ||--o{ feed_follows : "follows"
    feeds ||--o{ feed_follows : "followed_by"
    feeds ||--o{ posts : "contains"

    users {
        uuid id PK
        timestamp created_at
        timestamp updated_at
        text name UK
    }

    feeds {
        uuid id PK
        text name
        text url UK
        uuid user_id FK
        timestamp created_at
        timestamp updated_at
        timestamp last_fetched_at
    }

    feed_follows {
        uuid id PK
        uuid user_id FK
        uuid feed_id FK
        timestamp created_at
        timestamp updated_at
        unique(user_id, feed_id)
    }

    posts {
        uuid id PK
        text title
        text url UK
        text description
        timestamp published_at
        uuid feed_id FK
        timestamp created_at
        timestamp updated_at
    }
```
