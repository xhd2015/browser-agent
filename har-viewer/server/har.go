package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
)

var harDir string

func SetHARDir(dir string) {
	harDir = dir
}

func ListHARFiles() ([]string, error) {
	dir := harDir
	if dir == "" {
		dir = "."
	}
	matches, err := filepath.Glob(filepath.Join(dir, "*.har"))
	if err != nil {
		return nil, fmt.Errorf("failed to search for .har files: %v", err)
	}
	names := make([]string, 0, len(matches))
	for _, m := range matches {
		names = append(names, filepath.Base(m))
	}
	sort.Strings(names)
	return names, nil
}

type HARFile struct {
	Log HARLog `json:"log"`
}

type HARLog struct {
	Entries []json.RawMessage `json:"entries"`
}

type HAREntry struct {
	StartedDateTime string      `json:"startedDateTime"`
	Time            float64     `json:"time"`
	Request         HARRequest  `json:"request"`
	Response        HARResponse `json:"response"`
	Timings         HARTimings  `json:"timings"`
	ResourceType    string      `json:"_resourceType"`
	ServerIP        string      `json:"serverIPAddress"`
}

type HARRequest struct {
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	HTTPVersion string            `json:"httpVersion"`
	Headers     []HARNameValue    `json:"headers"`
	QueryString []HARNameValue    `json:"queryString"`
	PostData    *HARPostData      `json:"postData,omitempty"`
	HeadersSize int               `json:"headersSize"`
	BodySize    int               `json:"bodySize"`
	Cookies     []json.RawMessage `json:"cookies"`
}

type HARResponse struct {
	Status      int               `json:"status"`
	StatusText  string            `json:"statusText"`
	HTTPVersion string            `json:"httpVersion"`
	Headers     []HARNameValue    `json:"headers"`
	Content     HARContent        `json:"content"`
	HeadersSize int               `json:"headersSize"`
	BodySize    int               `json:"bodySize"`
	Cookies     []json.RawMessage `json:"cookies"`
}

type HARContent struct {
	Size     int    `json:"size"`
	MimeType string `json:"mimeType"`
	Text     string `json:"text,omitempty"`
	Encoding string `json:"encoding,omitempty"`
}

type HARPostData struct {
	MimeType string `json:"mimeType"`
	Text     string `json:"text"`
}

type HARNameValue struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type HARTimings struct {
	Blocked float64 `json:"blocked"`
	DNS     float64 `json:"dns"`
	Connect float64 `json:"connect"`
	Send    float64 `json:"send"`
	Wait    float64 `json:"wait"`
	Receive float64 `json:"receive"`
	SSL     float64 `json:"ssl"`
}

type EntrySummary struct {
	Index           int     `json:"index"`
	StartedDateTime string  `json:"startedDateTime"`
	Time            float64 `json:"time"`
	Method          string  `json:"method"`
	URL             string  `json:"url"`
	Host            string  `json:"host"`
	Path            string  `json:"path"`
	Status          int     `json:"status"`
	StatusText      string  `json:"statusText"`
	MimeType        string  `json:"mimeType"`
	Size            int     `json:"size"`
	ResourceType    string  `json:"resourceType"`
}

var (
	harCacheMu sync.Mutex
	harCache   = map[string][]json.RawMessage{}
)

func resolveHARPath(filename string) string {
	dir := harDir
	if dir == "" {
		dir = "."
	}
	return filepath.Join(dir, filepath.Base(filename))
}

func loadHAREntries(filename string) ([]json.RawMessage, error) {
	harCacheMu.Lock()
	defer harCacheMu.Unlock()

	if entries, ok := harCache[filename]; ok {
		return entries, nil
	}

	path := resolveHARPath(filename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read HAR file %s: %v", path, err)
	}
	var harFile HARFile
	if err := json.Unmarshal(data, &harFile); err != nil {
		return nil, fmt.Errorf("failed to parse HAR file %s: %v", path, err)
	}
	harCache[filename] = harFile.Log.Entries
	return harFile.Log.Entries, nil
}

func GetEntrySummaries(filename string) ([]EntrySummary, error) {
	entries, err := loadHAREntries(filename)
	if err != nil {
		return nil, err
	}

	summaries := make([]EntrySummary, 0, len(entries))
	for i, raw := range entries {
		var entry HAREntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			continue
		}

		parsed, _ := url.Parse(entry.Request.URL)
		host := ""
		path := entry.Request.URL
		if parsed != nil {
			host = parsed.Host
			path = parsed.Path
			if parsed.RawQuery != "" {
				path += "?" + parsed.RawQuery
			}
		}

		summaries = append(summaries, EntrySummary{
			Index:           i,
			StartedDateTime: entry.StartedDateTime,
			Time:            entry.Time,
			Method:          entry.Request.Method,
			URL:             entry.Request.URL,
			Host:            host,
			Path:            path,
			Status:          entry.Response.Status,
			StatusText:      entry.Response.StatusText,
			MimeType:        entry.Response.Content.MimeType,
			Size:            entry.Response.Content.Size,
			ResourceType:    entry.ResourceType,
		})
	}
	return summaries, nil
}

func RegisterHARAPI(mux *http.ServeMux) {
	mux.HandleFunc("/api/har/files", handleHARFiles)
	mux.HandleFunc("/api/har/entries", handleHAREntries)
	mux.HandleFunc("/api/har/entry", handleHAREntry)
}

func handleHARFiles(w http.ResponseWriter, r *http.Request) {
	files, err := ListHARFiles()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

func handleHAREntries(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "missing 'file' query parameter", http.StatusBadRequest)
		return
	}
	summaries, err := GetEntrySummaries(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summaries)
}

func handleHAREntry(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "missing 'file' query parameter", http.StatusBadRequest)
		return
	}

	entries, err := loadHAREntries(filename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	indexStr := r.URL.Query().Get("index")
	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 || index >= len(entries) {
		http.Error(w, "invalid index", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(entries[index])
}
