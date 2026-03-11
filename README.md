# Gator

A command-line RSS feed aggregator built in Go. Register users, subscribe to RSS feeds, automatically scrape new posts on a schedule, and browse them from your terminal.

## Features

- Multi-user support with simple username-based sessions
- Add and manage RSS feeds
- Follow/unfollow feeds per user
- Background aggregation with configurable intervals
- Browse posts from subscribed feeds
- PostgreSQL-backed persistent storage

## Prerequisites

- **Go** 1.25.5+
- **PostgreSQL** running and accessible
- **Goose** for database migrations (`go install github.com/pressly/goose/v3/cmd/goose@latest`)

## Installation

```bash
git clone https://github.com/pgrigorakis/gator.git
cd gator
go install
```

## Database Setup

1. Create a PostgreSQL database:

```sql
CREATE DATABASE gator;
```

2. Run the migrations with goose:

```bash
goose -dir sql/schema postgres "postgres://username:password@localhost:5432/gator?sslmode=disable" up
```

## Configuration

Gator reads its config from `~/.gatorconfig.json`. Create it with your database connection string:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

| Field               | Description                              |
|---------------------|------------------------------------------|
| `db_url`            | PostgreSQL connection string             |
| `current_user_name` | Currently logged-in user (set via CLI)   |

## Usage

### User Management

```bash
# Register a new user (sets as current user)
gator register alice

# Switch to an existing user
gator login alice

# List all users (* marks the current user)
gator users
```

### Feed Management

```bash
# Add a feed (automatically follows it)
gator addfeed "Go Blog" "https://go.dev/blog/feed.atom"

# List all feeds in the system
gator feeds
```

### Following Feeds

```bash
# Follow an existing feed
gator follow "https://go.dev/blog/feed.atom"

# List feeds you follow
gator following

# Unfollow a feed
gator unfollow "https://go.dev/blog/feed.atom"
```

### Aggregation

```bash
# Start the aggregator (runs continuously)
# Accepts Go duration strings: 30s, 1m, 5m, 1h, etc.
gator agg 1m
```

The aggregator fetches one feed per tick in round-robin order, prioritizing feeds that haven't been fetched yet or were fetched longest ago.

### Browsing Posts

```bash
# Browse posts from your followed feeds (default: 2 posts)
gator browse

# Specify a limit
gator browse 10
```

Posts are displayed in reverse chronological order.

### Reset

```bash
# Delete all data (users, feeds, follows, posts)
gator reset
```

## Commands Reference

| Command    | Arguments       | Auth | Description                              |
|------------|-----------------|------|------------------------------------------|
| `register` | `<username>`    | No   | Create a new user                        |
| `login`    | `<username>`    | No   | Switch to an existing user               |
| `users`    | —               | No   | List all registered users                |
| `addfeed`  | `<name> <url>`  | Yes  | Add a feed and auto-follow it            |
| `feeds`    | —               | No   | List all feeds                           |
| `follow`   | `<url>`         | Yes  | Follow an existing feed                  |
| `following`| —               | Yes  | List your followed feeds                 |
| `unfollow` | `<url>`         | Yes  | Unfollow a feed                          |
| `agg`      | `<interval>`    | No   | Start the feed aggregator                |
| `browse`   | `[limit]`       | Yes  | Browse posts (default limit: 2)          |
| `reset`    | —               | No   | Delete all data                          |

**Auth** = requires a logged-in user via `login` or `register`.

## Project Structure

```
gator/
├── main.go                 # Entry point, command registration
├── commands.go             # Command dispatcher
├── middleware.go            # Authentication middleware
├── handler_user.go          # register, login, users
├── handler_feed.go          # addfeed, feeds
├── handler_feed_follows.go  # follow, following, unfollow
├── handler_agg.go           # agg, browse, feed scraping
├── handler_reset.go         # reset
├── internal/
│   ├── config/
│   │   └── config.go       # Config file read/write
│   ├── database/            # sqlc-generated database layer
│   │   ├── db.go
│   │   ├── models.go
│   │   ├── users.sql.go
│   │   ├── feeds.sql.go
│   │   ├── feed_follows.sql.go
│   │   └── posts.sql.go
│   └── rss/
│       └── rss.go           # RSS feed fetching and parsing
└── sql/
    ├── schema/              # Goose migration files
    └── queries/             # sqlc query definitions
```

## Tech Stack

- **Go** — application language
- **PostgreSQL** — data storage
- **sqlc** — type-safe SQL code generation
- **goose** — database migrations
- **lib/pq** — PostgreSQL driver
- **google/uuid** — UUID generation
