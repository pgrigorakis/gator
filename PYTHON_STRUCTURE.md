# Gator RSS Feed Aggregator - Python Implementation

This document outlines how the Gator CLI RSS feed aggregator could be structured in Python, demonstrating idiomatic Python patterns and architecture.

## Project Structure

```
gator/
├── main.py                          # Application entry point
├── requirements.txt                 # Python dependencies
├── pyproject.toml                   # Modern Python project config
├── commands/
│   ├── __init__.py
│   ├── registry.py                  # Command pattern implementation
│   ├── user.py                      # User login/register/list handlers
│   ├── feed.py                      # Feed creation and listing handlers
│   ├── feed_follows.py              # Feed following/unfollowing handlers
│   ├── agg.py                       # Aggregation, scraping, and browse handlers
│   └── reset.py                     # Database reset handler
├── internal/
│   ├── __init__.py
│   ├── config.py                    # Configuration management
│   ├── database/
│   │   ├── __init__.py
│   │   ├── models.py                # SQLAlchemy models or dataclasses
│   │   ├── queries.py               # Database query layer
│   │   └── connection.py            # Database connection management
│   ├── rss/
│   │   ├── __init__.py
│   │   └── parser.py                # RSS feed parsing
│   └── middleware.py                # Authentication middleware
└── migrations/
    ├── 001_users.sql
    ├── 002_feeds.sql
    ├── 003_feed_follows.sql
    ├── 004_feeds_add_col.sql         # Adds last_fetched_at to feeds
    └── 005_posts.sql                 # Creates posts table
```

## Key Dependencies

```txt
# requirements.txt
psycopg2-binary>=2.9.0    # PostgreSQL driver
feedparser>=6.0.0          # RSS/Atom feed parsing
python-dotenv>=1.0.0       # Environment variable management
click>=8.1.0               # Alternative: argparse for CLI
sqlalchemy>=2.0.0          # ORM (optional, could use raw SQL)
```

## Core Implementation

### 1. Main Entry Point (`main.py`)

```python
#!/usr/bin/env python3
import sys
from pathlib import Path
from typing import Optional

from internal.config import Config
from internal.database.connection import Database
from commands.registry import CommandRegistry
from commands import user, feed, feed_follows, agg, reset
from internal.middleware import require_login


class State:
    """Application state container"""
    def __init__(self, db: Database, config: Config):
        self.db = db
        self.config = config


def main():
    # Load configuration
    try:
        config = Config.load()
    except Exception as e:
        print(f"Error reading config: {e}", file=sys.stderr)
        sys.exit(1)

    # Initialize database connection
    try:
        db = Database(config.db_url)
    except Exception as e:
        print(f"Error connecting to database: {e}", file=sys.stderr)
        sys.exit(1)

    # Create application state
    state = State(db=db, config=config)

    # Register all commands
    registry = CommandRegistry()
    registry.register("login", user.handler_login)
    registry.register("register", user.handler_register)
    registry.register("reset", reset.handler_reset_db)
    registry.register("users", user.handler_list_users)
    registry.register("agg", agg.handler_agg)
    registry.register("addfeed", require_login(feed.handler_add_feed))
    registry.register("feeds", feed.handler_list_feeds)
    registry.register("follow", require_login(feed_follows.handler_follow_feed))
    registry.register("following", require_login(feed_follows.handler_list_followed_feeds))
    registry.register("unfollow", require_login(feed_follows.handler_unfollow_feed))
    registry.register("browse", require_login(agg.handler_browse_feed))

    # Parse command line arguments
    if len(sys.argv) < 2:
        print("Error: no command entered", file=sys.stderr)
        sys.exit(1)

    command_name = sys.argv[1]
    command_args = sys.argv[2:]

    # Execute command
    try:
        registry.run(state, command_name, command_args)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)

    # Close database connection
    db.close()


if __name__ == "__main__":
    main()
```

### 2. Command Registry (`commands/registry.py`)

```python
from typing import Callable, Dict, List, Any
from dataclasses import dataclass


@dataclass
class Command:
    """Command data structure"""
    name: str
    args: List[str]


# Type alias for handler functions
HandlerFunc = Callable[[Any, Command], None]


class CommandRegistry:
    """Command pattern implementation with registration"""

    def __init__(self):
        self._commands: Dict[str, HandlerFunc] = {}

    def register(self, name: str, handler: HandlerFunc):
        """Register a command handler"""
        if name in self._commands:
            raise ValueError(f"Command '{name}' already registered")
        self._commands[name] = handler

    def run(self, state: Any, command_name: str, args: List[str]):
        """Execute a registered command"""
        if command_name not in self._commands:
            raise ValueError(f"Unknown command: {command_name}")

        cmd = Command(name=command_name, args=args)
        handler = self._commands[command_name]
        handler(state, cmd)
```

### 3. Configuration Management (`internal/config.py`)

```python
import json
from pathlib import Path
from dataclasses import dataclass, asdict
from typing import Optional


@dataclass
class Config:
    """Application configuration"""
    db_url: str
    current_user_name: Optional[str] = None

    @staticmethod
    def _get_config_path() -> Path:
        """Get the path to the config file"""
        home = Path.home()
        return home / ".gatorconfig.json"

    @classmethod
    def load(cls) -> "Config":
        """Load configuration from file"""
        config_path = cls._get_config_path()

        if not config_path.exists():
            raise FileNotFoundError(f"Config file not found: {config_path}")

        with open(config_path, 'r') as f:
            data = json.load(f)

        return cls(
            db_url=data.get("db_url", ""),
            current_user_name=data.get("current_user_name")
        )

    def save(self):
        """Save configuration to file"""
        config_path = self._get_config_path()

        with open(config_path, 'w') as f:
            json.dump(asdict(self), f, indent=2)

    def set_user(self, username: str):
        """Set the current user and persist to config"""
        if not username:
            raise ValueError("No username provided")

        self.current_user_name = username
        self.save()
```

### 4. Database Layer (`internal/database/connection.py`)

```python
import psycopg2
from psycopg2.extras import RealDictCursor
from typing import Optional, List, Dict, Any
from contextlib import contextmanager


class Database:
    """Database connection and query management"""

    def __init__(self, connection_string: str):
        self.connection_string = connection_string
        self._conn = psycopg2.connect(connection_string)

    @contextmanager
    def cursor(self):
        """Context manager for database cursor"""
        cur = self._conn.cursor(cursor_factory=RealDictCursor)
        try:
            yield cur
            self._conn.commit()
        except Exception:
            self._conn.rollback()
            raise
        finally:
            cur.close()

    def close(self):
        """Close database connection"""
        if self._conn:
            self._conn.close()

    # User queries
    def create_user(self, user_id: str, name: str, created_at, updated_at) -> Dict[str, Any]:
        """Create a new user"""
        with self.cursor() as cur:
            cur.execute(
                """
                INSERT INTO users (id, created_at, updated_at, name)
                VALUES (%s, %s, %s, %s)
                RETURNING *
                """,
                (user_id, created_at, updated_at, name)
            )
            return dict(cur.fetchone())

    def get_user(self, name: str) -> Optional[Dict[str, Any]]:
        """Get user by name"""
        with self.cursor() as cur:
            cur.execute("SELECT * FROM users WHERE name = %s", (name,))
            result = cur.fetchone()
            return dict(result) if result else None

    def get_all_users(self) -> List[Dict[str, Any]]:
        """Get all users"""
        with self.cursor() as cur:
            cur.execute("SELECT * FROM users ORDER BY name")
            return [dict(row) for row in cur.fetchall()]

    # Feed queries
    def create_feed(self, feed_id: str, name: str, url: str, user_id: str,
                    created_at, updated_at) -> Dict[str, Any]:
        """Create a new feed (last_fetched_at defaults to NULL)"""
        with self.cursor() as cur:
            cur.execute(
                """
                INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
                VALUES (%s, %s, %s, %s, %s, %s)
                RETURNING *
                """,
                (feed_id, created_at, updated_at, name, url, user_id)
            )
            return dict(cur.fetchone())

    def get_feed_by_url(self, url: str) -> Optional[Dict[str, Any]]:
        """Get feed by URL"""
        with self.cursor() as cur:
            cur.execute("SELECT * FROM feeds WHERE url = %s", (url,))
            result = cur.fetchone()
            return dict(result) if result else None

    def get_all_feeds(self) -> List[Dict[str, Any]]:
        """Get all feeds with user information"""
        with self.cursor() as cur:
            cur.execute(
                """
                SELECT f.*, u.name as user_name
                FROM feeds f
                JOIN users u ON f.user_id = u.id
                ORDER BY f.created_at
                """
            )
            return [dict(row) for row in cur.fetchall()]

    def mark_feed_fetched(self, feed_id: str) -> Dict[str, Any]:
        """Mark a feed as fetched, setting last_fetched_at to NOW()"""
        with self.cursor() as cur:
            cur.execute(
                """
                UPDATE feeds
                SET last_fetched_at = NOW(), updated_at = NOW()
                WHERE id = %s
                RETURNING *
                """,
                (feed_id,)
            )
            return dict(cur.fetchone())

    def get_next_feed_to_fetch(self) -> Optional[Dict[str, Any]]:
        """Get the feed that was least recently fetched (NULLs first)"""
        with self.cursor() as cur:
            cur.execute(
                """
                SELECT * FROM feeds
                ORDER BY last_fetched_at NULLS FIRST
                LIMIT 1
                """
            )
            result = cur.fetchone()
            return dict(result) if result else None

    # Post queries
    def create_post(self, post_id: str, title: str, url: str, feed_id: str,
                    created_at, updated_at,
                    description: Optional[str] = None,
                    published_at=None) -> Dict[str, Any]:
        """Create a new post"""
        with self.cursor() as cur:
            cur.execute(
                """
                INSERT INTO posts (id, created_at, updated_at, title, url,
                                   description, published_at, feed_id)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                RETURNING *
                """,
                (post_id, created_at, updated_at, title, url,
                 description, published_at, feed_id)
            )
            return dict(cur.fetchone())

    def get_posts_for_user(self, user_id: str, limit: int = 2) -> List[Dict[str, Any]]:
        """Get posts from feeds the user follows, ordered by newest first"""
        with self.cursor() as cur:
            cur.execute(
                """
                SELECT posts.*
                FROM posts
                INNER JOIN feed_follows ON posts.feed_id = feed_follows.feed_id
                WHERE feed_follows.user_id = %s
                ORDER BY published_at DESC
                LIMIT %s
                """,
                (user_id, limit)
            )
            return [dict(row) for row in cur.fetchall()]

    # Feed follow queries
    def create_feed_follow(self, follow_id: str, user_id: str, feed_id: str,
                          created_at, updated_at) -> Dict[str, Any]:
        """Create a feed follow relationship"""
        with self.cursor() as cur:
            cur.execute(
                """
                INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
                VALUES (%s, %s, %s, %s, %s)
                RETURNING *
                """,
                (follow_id, created_at, updated_at, user_id, feed_id)
            )
            return dict(cur.fetchone())

    def get_feed_follows_by_user(self, user_id: str) -> List[Dict[str, Any]]:
        """Get all feeds followed by a user"""
        with self.cursor() as cur:
            cur.execute(
                """
                SELECT ff.*, f.name as feed_name, f.url as feed_url
                FROM feed_follows ff
                JOIN feeds f ON ff.feed_id = f.id
                WHERE ff.user_id = %s
                ORDER BY ff.created_at
                """,
                (user_id,)
            )
            return [dict(row) for row in cur.fetchall()]

    def delete_feed_follow(self, user_id: str, feed_id: str):
        """Delete a feed follow relationship"""
        with self.cursor() as cur:
            cur.execute(
                """
                DELETE FROM feed_follows
                WHERE user_id = %s AND feed_id = %s
                """,
                (user_id, feed_id)
            )

    def reset_database(self):
        """Delete all users and feeds (cascades to feed_follows)"""
        with self.cursor() as cur:
            cur.execute("DELETE FROM users")
            cur.execute("DELETE FROM feeds")
```

### 5. Middleware (`internal/middleware.py`)

```python
from functools import wraps
from typing import Callable


def require_login(handler: Callable) -> Callable:
    """
    Middleware decorator that requires a logged-in user.
    Injects the user object into the handler.
    """
    @wraps(handler)
    def wrapper(state, cmd):
        # Check if user is logged in
        if not state.config.current_user_name:
            raise ValueError("No user is currently logged in. Use 'login' or 'register' first.")

        # Fetch user from database
        user = state.db.get_user(state.config.current_user_name)
        if not user:
            raise ValueError(f"Current user '{state.config.current_user_name}' not found in database")

        # Call handler with user injected
        return handler(state, cmd, user)

    return wrapper
```

### 6. User Handlers (`commands/user.py`)

```python
from datetime import datetime, timezone
import uuid


def handler_login(state, cmd):
    """Log in as an existing user"""
    if len(cmd.args) != 1:
        raise ValueError(f"Usage: {cmd.name} <name>")

    username = cmd.args[0]

    # Check if user exists
    user = state.db.get_user(username)
    if not user:
        raise ValueError(f"User '{username}' not found")

    # Set current user in config
    state.config.set_user(username)

    print(f"User has been set to: {username}")


def handler_register(state, cmd):
    """Register a new user"""
    if len(cmd.args) != 1:
        raise ValueError(f"Usage: {cmd.name} <name>")

    username = cmd.args[0]

    # Create user
    now = datetime.now(timezone.utc)
    user = state.db.create_user(
        user_id=str(uuid.uuid4()),
        name=username,
        created_at=now,
        updated_at=now
    )

    # Set current user in config
    state.config.set_user(username)

    print(f"User created: {username}")
    _print_user(user)


def handler_list_users(state, cmd):
    """List all registered users"""
    users = state.db.get_all_users()

    for user in users:
        marker = " (current)" if user['name'] == state.config.current_user_name else ""
        print(f"* {user['name']}{marker}")


def _print_user(user):
    """Helper to print user details"""
    print(f" * ID:      {user['id']}")
    print(f" * Name:    {user['name']}")
```

### 7. Feed Handlers (`commands/feed.py`)

```python
from datetime import datetime, timezone
import uuid
from internal.rss.parser import fetch_rss


def handler_add_feed(state, cmd, user):
    """Add a new RSS feed (requires login)"""
    if len(cmd.args) != 2:
        raise ValueError(f"Usage: {cmd.name} <name> <url>")

    feed_name = cmd.args[0]
    feed_url = cmd.args[1]

    # Create feed
    now = datetime.now(timezone.utc)
    feed = state.db.create_feed(
        feed_id=str(uuid.uuid4()),
        name=feed_name,
        url=feed_url,
        user_id=user['id'],
        created_at=now,
        updated_at=now
    )

    print(f"Feed created: {feed['name']}")
    print(f"  URL: {feed['url']}")


def handler_list_feeds(state, cmd):
    """List all feeds"""
    feeds = state.db.get_all_feeds()

    for feed in feeds:
        print(f"* {feed['name']}")
        print(f"  URL: {feed['url']}")
        print(f"  Created by: {feed['user_name']}")


```

### 7b. Aggregation and Browse Handlers (`commands/agg.py`)

```python
import time
import uuid
import logging
from datetime import datetime, timezone
from typing import Optional

from internal.rss.parser import fetch_rss
from psycopg2.errors import UniqueViolation

logger = logging.getLogger(__name__)


def _parse_pub_date(pub_date: str) -> Optional[datetime]:
    """Attempt to parse a publication date string in multiple formats"""
    formats = [
        "%a, %d %b %Y %H:%M:%S %z",   # RFC 1123 with timezone
        "%a, %d %b %Y %H:%M:%S %Z",   # RFC 1123 with named timezone
    ]
    for fmt in formats:
        try:
            return datetime.strptime(pub_date, fmt)
        except ValueError:
            continue
    return None


def _add_post(state, feed: dict, item: dict):
    """Save an RSS item as a post, with nullable description and published_at"""
    now = datetime.now(timezone.utc)
    published_at = _parse_pub_date(item.get('published', ''))
    if published_at is None:
        logger.warning("Could not parse publication time for %s", item['title'])

    description = item.get('description') or None

    state.db.create_post(
        post_id=str(uuid.uuid4()),
        title=item['title'],
        url=item['link'],
        feed_id=feed['id'],
        created_at=now,
        updated_at=now,
        description=description,
        published_at=published_at,
    )


def _scrape_feeds(state):
    """Fetch the next unfetched feed, parse RSS, and save posts to DB"""
    next_feed = state.db.get_next_feed_to_fetch()
    if not next_feed:
        logger.info("No feeds to fetch")
        return

    state.db.mark_feed_fetched(next_feed['id'])

    try:
        items = fetch_rss(next_feed['url'])
    except Exception as e:
        logger.error("Could not fetch feed %s: %s", next_feed['url'], e)
        return

    for item in items:
        print(f"Found post: [{next_feed['name']}] {item['title']}")
        try:
            _add_post(state, next_feed, item)
        except UniqueViolation:
            continue  # Skip duplicate posts
        except Exception as e:
            logger.error("Could not create post %s: %s", item['title'], e)

    logger.info("Feed %s collected, %d posts found", next_feed['name'], len(items))


def handler_agg(state, cmd):
    """Continuously scrape feeds on a timed interval"""
    if len(cmd.args) != 1:
        raise ValueError(f"Usage: {cmd.name} <interval e.g. 30s, 5m>")

    interval_str = cmd.args[0]

    # Parse duration string (e.g. "30s", "5m", "1h30m")
    # Python doesn't have Go's time.ParseDuration, so we parse manually
    seconds = _parse_duration(interval_str)

    print(f"Collecting feeds every {interval_str}...")
    while True:
        _scrape_feeds(state)
        time.sleep(seconds)


def _parse_duration(duration_str: str) -> float:
    """Parse a Go-style duration string into seconds (e.g. '1m30s' -> 90.0)"""
    import re
    total = 0.0
    pattern = re.compile(r'(\d+)(h|m|s)')
    matches = pattern.findall(duration_str)
    if not matches:
        raise ValueError(f"Invalid duration format: {duration_str}")
    for value, unit in matches:
        if unit == 'h':
            total += int(value) * 3600
        elif unit == 'm':
            total += int(value) * 60
        elif unit == 's':
            total += int(value)
    return total


def handler_browse_feed(state, cmd, user):
    """Browse posts from subscribed feeds (requires login)"""
    if len(cmd.args) > 1:
        raise ValueError(f"Usage: {cmd.name} [limit, default=2]")

    limit = 2
    if len(cmd.args) == 1:
        try:
            limit = int(cmd.args[0])
        except ValueError:
            raise ValueError(f"Could not parse argument: {cmd.args[0]}")

    posts = state.db.get_posts_for_user(user['id'], limit=limit)

    for post in posts:
        if post.get('description'):
            print(f"     {post['description']}")
```

### 8. RSS Parser (`internal/rss/parser.py`)

```python
import feedparser
from typing import List, Dict
import html


def fetch_rss(url: str) -> List[Dict[str, str]]:
    """
    Fetch and parse an RSS feed from the given URL.
    Returns a list of feed items with title, link, and description.
    """
    # Parse the feed
    feed = feedparser.parse(url)

    if feed.bozo:  # feedparser sets this flag for malformed feeds
        raise ValueError(f"Error parsing feed: {feed.bozo_exception}")

    items = []
    for entry in feed.entries:
        # Unescape HTML entities
        title = html.unescape(entry.get('title', 'No title'))
        link = entry.get('link', '')
        description = html.unescape(entry.get('description', ''))

        items.append({
            'title': title,
            'link': link,
            'description': description,
            'published': entry.get('published', '')
        })

    return items
```

## Key Differences from Go Implementation

### 1. **Type System**
- Go: Static typing with compile-time checks
- Python: Dynamic typing with optional type hints (using `typing` module)

### 2. **Error Handling**
- Go: Explicit error returns (`error` as return value)
- Python: Exception-based error handling (try/except)

### 3. **Database Access**
- Go: Using `sqlc` for compile-time SQL generation
- Python: Options include SQLAlchemy (ORM), raw `psycopg2`, or query builders

### 4. **Package Management**
- Go: Go modules (`go.mod`)
- Python: pip with `requirements.txt` or modern `pyproject.toml`

### 5. **Middleware Pattern**
- Go: Higher-order functions returning handler functions
- Python: Decorators (syntactic sugar for function wrapping)

### 6. **Context Management**
- Go: Explicit `context.Context` for cancellation
- Python: Context managers (`with` statements) for resource cleanup

### 7. **Concurrency**
- Go: Goroutines and channels (not used in this CLI app)
- Python: `asyncio`, threading, or multiprocessing (would use `asyncio` for I/O-bound RSS fetching)

## Alternative Python Approaches

### Using Click for CLI

```python
import click
from internal.config import Config
from internal.database.connection import Database

@click.group()
@click.pass_context
def cli(ctx):
    """Gator RSS Feed Aggregator"""
    ctx.obj = State(db=Database.connect(), config=Config.load())

@cli.command()
@click.argument('name')
@click.pass_obj
def login(state, name):
    """Log in as an existing user"""
    # Implementation here
    pass

@cli.command()
@click.argument('name')
@click.pass_obj
def register(state, name):
    """Register a new user"""
    # Implementation here
    pass

if __name__ == '__main__':
    cli()
```

### Using SQLAlchemy ORM

```python
from sqlalchemy import create_engine, Column, String, DateTime, ForeignKey
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker, relationship
from datetime import datetime
import uuid

Base = declarative_base()

class User(Base):
    __tablename__ = 'users'

    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    name = Column(String, unique=True, nullable=False)

    feeds = relationship('Feed', back_populates='user')
    feed_follows = relationship('FeedFollow', back_populates='user')

class Feed(Base):
    __tablename__ = 'feeds'

    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    name = Column(String, nullable=False)
    url = Column(String, unique=True, nullable=False)
    user_id = Column(String, ForeignKey('users.id'))
    last_fetched_at = Column(DateTime, nullable=True)

    user = relationship('User', back_populates='feeds')
    feed_follows = relationship('FeedFollow', back_populates='feed')
    posts = relationship('Post', back_populates='feed')

class FeedFollow(Base):
    __tablename__ = 'feed_follows'

    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    user_id = Column(String, ForeignKey('users.id'))
    feed_id = Column(String, ForeignKey('feeds.id'))

    user = relationship('User', back_populates='feed_follows')
    feed = relationship('Feed', back_populates='feed_follows')

class Post(Base):
    __tablename__ = 'posts'

    id = Column(String, primary_key=True, default=lambda: str(uuid.uuid4()))
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)
    title = Column(String, nullable=False)
    url = Column(String, unique=True, nullable=False)
    description = Column(String, nullable=True)
    published_at = Column(DateTime, nullable=True)
    feed_id = Column(String, ForeignKey('feeds.id'))

    feed = relationship('Feed', back_populates='posts')
```

## Conclusion

The Python implementation would be more concise due to:
- Dynamic typing (less boilerplate)
- Rich standard library (`json`, `pathlib`, `functools`)
- Mature third-party libraries (`feedparser`, `psycopg2`, `click`)
- Decorator syntax for middleware
- Context managers for resource cleanup

However, it would trade off:
- Compile-time type safety
- Performance (Python is slower than Go)
- Explicit error handling (exceptions can be hidden)
- Built-in concurrency primitives

Both approaches are valid, with Go excelling at performance and type safety, while Python offers rapid development and extensive library support.
