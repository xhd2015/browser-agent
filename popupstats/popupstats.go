// Package popupstats builds pure popup recording stats for API Capture.
// Rules mirror Chrome-Ext-Capture-API getState enrichment (see popupStats.js).
package popupstats

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// OpaqueHost is the single bucket name for invalid/unparseable request URLs.
const OpaqueHost = "opaque"

// EntryValue is one capture buffer entry (request.url only for stats).
type EntryValue struct {
	URL string // request.url
}

// TabMeta is chrome.tabs-style metadata for a tab id.
type TabMeta struct {
	Title  string
	URL    string
	Active bool
}

// DomainCount is a host and how many requests hit it.
type DomainCount struct {
	Host  string
	Count int
}

// TabStats is one row in the popup "By tab" list.
type TabStats struct {
	TabID        int
	Title        string
	URL          string
	Active       bool
	Attached     bool
	RequestCount int
	DomainCount  int
	Domains      []DomainCount // top 3 only
}

// PopupStats is the enriched getState payload shape for recording chips + list.
type PopupStats struct {
	Count        int
	TabsWatching int
	DomainCount  int // unique hosts across all entries
	Tabs         []TabStats
}

// Input is the pure builder input (in-memory capture state).
type Input struct {
	Entries        map[string]EntryValue // key "${tabId}:${requestId}" preferred
	AttachedTabIDs []int
	TabMeta        map[int]TabMeta
}

type tabAccum struct {
	tabID    int
	hosts    map[string]int
	reqCount int
}

// BuildPopupStats turns capture buffer + attached set + tab meta into popup stats.
func BuildPopupStats(in Input) PopupStats {
	entries := in.Entries
	if entries == nil {
		entries = map[string]EntryValue{}
	}
	tabMeta := in.TabMeta
	if tabMeta == nil {
		tabMeta = map[int]TabMeta{}
	}

	attachedSet := make(map[int]struct{}, len(in.AttachedTabIDs))
	for _, id := range in.AttachedTabIDs {
		attachedSet[id] = struct{}{}
	}

	// Per-tab host counts + membership from entries.
	byTab := make(map[int]*tabAccum)
	globalHosts := make(map[string]struct{})

	for key, ev := range entries {
		tabID, ok := tabIDFromKey(key)
		if !ok {
			// Still count globally so chips reflect buffer size; skip per-tab
			// attribution for unparseable keys (production always uses tabId:...).
			host := hostFromURL(ev.URL)
			globalHosts[host] = struct{}{}
			continue
		}
		acc := byTab[tabID]
		if acc == nil {
			acc = &tabAccum{tabID: tabID, hosts: make(map[string]int)}
			byTab[tabID] = acc
		}
		acc.reqCount++
		host := hostFromURL(ev.URL)
		acc.hosts[host]++
		globalHosts[host] = struct{}{}
	}

	// Union: attached OR has ≥1 entry.
	memberIDs := make(map[int]struct{}, len(attachedSet)+len(byTab))
	for id := range attachedSet {
		memberIDs[id] = struct{}{}
	}
	for id := range byTab {
		memberIDs[id] = struct{}{}
	}

	tabs := make([]TabStats, 0, len(memberIDs))
	for id := range memberIDs {
		_, attached := attachedSet[id]
		meta := tabMeta[id]
		acc := byTab[id]

		reqCount := 0
		var hosts map[string]int
		if acc != nil {
			reqCount = acc.reqCount
			hosts = acc.hosts
		}
		if hosts == nil {
			hosts = map[string]int{}
		}

		domainCount := len(hosts)
		domains := topDomains(hosts, 3)

		title := meta.Title
		if title == "" {
			if attached {
				title = fmt.Sprintf("Tab %d", id)
			} else {
				title = "Closed tab"
			}
		}

		tabs = append(tabs, TabStats{
			TabID:        id,
			Title:        title,
			URL:          meta.URL,
			Active:       meta.Active,
			Attached:     attached,
			RequestCount: reqCount,
			DomainCount:  domainCount,
			Domains:      domains,
		})
	}

	sort.SliceStable(tabs, func(i, j int) bool {
		if tabs[i].RequestCount != tabs[j].RequestCount {
			return tabs[i].RequestCount > tabs[j].RequestCount
		}
		return tabs[i].TabID < tabs[j].TabID
	})

	// Entries that failed tab attribution still count toward global count.
	count := len(entries)

	return PopupStats{
		Count:        count,
		TabsWatching: len(attachedSet),
		DomainCount:  len(globalHosts),
		Tabs:         tabs,
	}
}

// tabIDFromKey parses the integer tab id before the first ':' in entry key.
func tabIDFromKey(key string) (int, bool) {
	idx := strings.IndexByte(key, ':')
	if idx <= 0 {
		return 0, false
	}
	id, err := strconv.Atoi(key[:idx])
	if err != nil {
		return 0, false
	}
	return id, true
}

// hostFromURL returns the URL hostname, or OpaqueHost when unparseable/empty.
func hostFromURL(raw string) string {
	if raw == "" {
		return OpaqueHost
	}
	u, err := url.Parse(raw)
	if err != nil {
		return OpaqueHost
	}
	host := u.Hostname()
	if host == "" {
		return OpaqueHost
	}
	return host
}

// topDomains returns up to n hosts sorted by count desc, then host asc.
func topDomains(hosts map[string]int, n int) []DomainCount {
	if len(hosts) == 0 {
		return nil
	}
	list := make([]DomainCount, 0, len(hosts))
	for h, c := range hosts {
		list = append(list, DomainCount{Host: h, Count: c})
	}
	sort.SliceStable(list, func(i, j int) bool {
		if list[i].Count != list[j].Count {
			return list[i].Count > list[j].Count
		}
		return list[i].Host < list[j].Host
	})
	if n > 0 && len(list) > n {
		list = list[:n]
	}
	return list
}
