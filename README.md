
## Description

A REST API backend for storing and organizing RSS feeds, their items (articles), and grouping feeds into collections. Can be run as a standalone Docker container, paired with the [rss-feed-reader-frontend](https://github.com/Swunci/rss-feed-reader-frontend), or bundled with the frontend as a single Windows executable.

## Features

- RSS feed management (add, update, delete)
- Background polling per feed at a configurable interval (default: 15 minutes)
- Feed discovery for YouTube channels and Reddit URLs
- Organize feeds into collections
- Filter items by all, unread, and favorite status
- New item notifications pushed to the client automatically via SSE
- Cursor-based pagination with timestamp and limit size

## How to Run

### Development

1. Clone the repository:
```bash
git clone https://github.com/Swunci/rss-feed-reader-backend.git
cd rss-feed-backend
```

2. Copy the example env file and configure it:
```bash
cp .env-example .env
```

3. Run the server:
```bash
go run ./cmd/rss-feed-backend/server/main.go
```

The server will be available at `http://localhost:8082`.

---

### Docker

**Backend only**

```bash
docker compose up --build -d
```

The backend will be available at `http://localhost:8081`.

**[Frontend](https://github.com/Swunci/rss-feed-reader-frontend) + Backend**

1. Create the production env file:

```bash
cp .env.production-example .env.production
```

2. (Linux only) Create the database directory and set permissions:

```bash
mkdir -p db-data && chown -R 10001:10001 db-data
```

3. Start all services:
```bash
docker compose -p rss-feed-reader -f compose-fullstack-app.yaml up -d
```

The app will be available at `http://localhost` or at your server's IP address.

---

### Windows Executable

1. Clone both repositories into the same parent directory:
```bash
git clone https://github.com/Swunci/rss-feed-reader-backend.git
git clone https://github.com/Swunci/rss-feed-reader-frontend.git
```

2. From the backend directory, build the `.exe`:
```bash
cd rss-feed-backend
make clean && make build-desktop
```

3. Move `rss-reader.exe` to any folder you want the app to live in. The SQLite database files will be created in the same directory as the executable when it runs.

4. Run `rss-reader.exe`. The app will start in the system tray and open `http://localhost:7721` in your browser.


## Background Polling

Each feed runs a background goroutine that polls for new items at a configurable interval (default: 15 minutes), controlled by the `POLLING_INTERVAL_MINUTES` environment variable. When a feed is added polling starts immediately and performs an initial fetch right away. When a feed is deleted its goroutine is stopped. New items are deduplicated on insert, and connected clients are notified via the `/items/events` SSE endpoint when new items arrive.



## API Endpoints

### Feeds

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/feeds` | Get all feeds with item counts |
| GET | `/feeds?filter=unread` | Get feeds that have unread items |
| GET | `/feeds?filter=favorite` | Get feeds that have favorited items |
| GET | `/feeds/{feed_id}` | Get a single feed |
| GET | `/feeds/{feed_id}/items` | Get items by feed |
| POST | `/feeds` | Add a new feed |
| POST | `/feeds/discover` | Discover feed options from a URL |
| POST | `/feeds/refresh` | Refresh all feeds (rate limited: 1 per 5 minutes) |
| POST | `/feeds/{feed_id}/refresh` | Refresh a single feed |
| PATCH | `/feeds/{feed_id}` | Update a feed |
| DELETE | `/feeds/{feed_id}` | Delete a feed |
| DELETE | `/feeds/{feed_id}/unassign` | Remove feed from its collection |

**POST** `/feeds`
```json
{
  "url": "https://example.com/feed.xml",
  "name": "Optional custom name"
}
```

**POST** `/feeds/discover`
```json
{
  "url": "https://www.youtube.com/@channelname"
}
```

**PATCH** `/feeds/{feed_id}`
```json
{
  "name": "New name",
  "url": "https://example.com/feed.xml",
  "collection_id": 1
}
```
All fields are optional.

---

### Items

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/items` | Get all items |
| GET | `/items/{item_id}` | Get a single item |
| GET | `/items/events` | SSE stream for new item events |
| PATCH | `/items/{item_id}/read` | Mark item as read/unread |
| PATCH | `/items/read` | Mark multiple items as read/unread |
| PATCH | `/items/{item_id}/favorite` | Toggle item favorite |

**PATCH** `/items/{item_id}/read`
```json
{
  "is_read": true
}
```

**PATCH** `/items/read`
```json
{
  "item_ids": [1, 2, 3],
  "is_read": true
}
```

**PATCH** `/items/{item_id}/favorite`
```json
{
  "is_favorite": true
}
```

---

### Collections

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/collections` | Get all collections |
| GET | `/collections/{collection_id}/items` | Get items by collection |
| POST | `/collections` | Create a collection |
| PUT | `/collections/{collection_id}` | Rename a collection |
| DELETE | `/collections/{collection_id}` | Delete a collection |

**POST** `/collections`
```json
{
  "name": "Tech News"
}
```

**PUT** `/collections/{collection_id}`
```json
{
  "name": "New Name"
}
```

---

### Query Parameters

All item listing endpoints (`/items`, `/feeds/{feed_id}/items`, `/collections/{collection_id}/items`) support:

| Parameter | Values | Description |
|-----------|--------|-------------|
| `cursor` | timestamp string | Pagination cursor (published_at of last item) |
| `limit` | integer | Number of items to return per page. Omit or set to 0 to return all items |
| `read` | `true` / `false` | Filter by read status |
| `favorite` | `true` / `false` | Filter by favorite status |