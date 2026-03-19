package main

import "sync"

// FetchContent returns up to `count` items from the config sequence starting at `offset`.
// For each slot: use main provider; on failure use fallback if set; if both fail, return
// only items collected so far. Fetches run in parallel to keep latency low.
func FetchContent(config ContentMix, clients map[Provider]Client, userIP string, count, offset int) []*ContentItem {
	if count <= 0 || len(config) == 0 {
		return nil
	}
	if offset < 0 {
		offset = 0
	}

	// One result per requested slot; each slot runs in its own goroutine.
	results := make([]*ContentItem, count)
	var wg sync.WaitGroup
	wg.Add(count)

	for slot := 0; slot < count; slot++ {
		go func(slot int) {
			defer wg.Done()
			cfg := config[(offset+slot)%len(config)]
			results[slot] = fetchOneSlot(cfg, clients, userIP)
		}(slot)
	}

	wg.Wait()

	// Return in order; stop at first slot that failed (main + fallback both failed).
	out := make([]*ContentItem, 0, count)
	for _, item := range results {
		if item == nil {
			break
		}
		out = append(out, item)
	}
	return out
}

// fetchOneSlot tries main provider, then fallback. Returns nil if both fail.
func fetchOneSlot(cfg ContentConfig, clients map[Provider]Client, userIP string) *ContentItem {
	if client, ok := clients[cfg.Type]; ok {
		items, err := client.GetContent(userIP, 1)
		if err == nil && len(items) > 0 {
			return items[0]
		}
	}
	if cfg.Fallback != nil {
		if client, ok := clients[*cfg.Fallback]; ok {
			items, err := client.GetContent(userIP, 1)
			if err == nil && len(items) > 0 {
				return items[0]
			}
		}
	}
	return nil
}
