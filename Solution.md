# Solution

**Problem:** `GET /?count=N&offset=M` must return N content items in the order defined by the config sequence, with fallback per slot; if main and fallback both fail, return only items up to that point. Fetches must be parallel.

**Approach:** Thin HTTP in `server.go` (parse, validate, call fetch, write JSON). All fetch logic in `aggregator.go`: one function that runs slots in parallel, tries main then fallback per slot, returns ordered list and stops at first failure.

**What was implemented**

- **aggregator.go:** `FetchContent(config, clients, userIP, count, offset)` — for each slot `(offset+slot)%len(config)`, call `fetchOneSlot` (main then fallback); goroutines + WaitGroup; build result in order, stop at first nil.
- **server.go:** `ServeHTTP` — GET only; parse `count`/`offset` (offset defaults to 0); call `FetchContent`; 200 + JSON. Helpers: `parseCountOffset`, `userIPFromRequest`.
- **main.go:** Use `Handler: &app` so `*App` is the handler.

**Sample response** (`curl -s 'http://127.0.0.1:8080/?count=2&offset=0'`)

```json
[
  {
    "id": "2763033652435159293",
    "title": "title",
    "source": "1",
    "summary": "",
    "link": "",
    "expiry": "2026-03-19T00:01:45.371424Z"
  },
  {
    "id": "3813514960796899517",
    "title": "title",
    "source": "1",
    "summary": "",
    "link": "",
    "expiry": "2026-03-19T00:01:45.371422Z"
  }
]
```

**Run**

```bash
go run .
curl 'http://127.0.0.1:8080/?count=5&offset=0'
go test -v .
```
