package bot

import (
	"bufio"
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	playlistRequestTimeout = 15 * time.Second
	playlistMaxEntries     = 100
	playlistResponseLimit  = 2 << 20
)

type PlaylistEntry struct {
	Title string
	Query string
}

func FetchPlaylistEntries(ctx context.Context, playlistURL string) ([]PlaylistEntry, error) {
	cleanedURL := strings.TrimSpace(playlistURL)
	if cleanedURL == "" {
		return nil, fmt.Errorf("playlist URL cannot be empty")
	}

	parsedURL, err := neturl.Parse(cleanedURL)
	if err != nil {
		return nil, fmt.Errorf("invalid playlist URL: %w", err)
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, fmt.Errorf("playlist URL must use http or https")
	}

	if ctx == nil {
		ctx = context.Background()
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, cleanedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create playlist request: %w", err)
	}

	client := &http.Client{Timeout: playlistRequestTimeout}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("fetch playlist: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		snippet, _ := io.ReadAll(io.LimitReader(response.Body, 1024))
		snippetText := strings.TrimSpace(string(snippet))
		if snippetText != "" {
			return nil, fmt.Errorf("playlist request failed: %s: %s", response.Status, snippetText)
		}

		return nil, fmt.Errorf("playlist request failed: %s", response.Status)
	}

	body, err := io.ReadAll(io.LimitReader(response.Body, playlistResponseLimit))
	if err != nil {
		return nil, fmt.Errorf("read playlist response: %w", err)
	}

	entries, err := parsePlaylistDocument(body, response.Header.Get("Content-Type"), parsedURL)
	if err != nil {
		return nil, err
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no tracks found at %s", cleanedURL)
	}

	return entries, nil
}

func parsePlaylistDocument(body []byte, contentType string, baseURL *neturl.URL) ([]PlaylistEntry, error) {
	normalizedContentType := strings.ToLower(contentType)
	extension := strings.ToLower(path.Ext(baseURL.Path))

	switch {
	case strings.Contains(normalizedContentType, "application/json") || extension == ".json":
		return parsePlaylistJSON(body)
	case strings.Contains(normalizedContentType, "text/csv") || extension == ".csv":
		return parsePlaylistCSV(body)
	case strings.Contains(normalizedContentType, "text/html"):
		if entries := parsePlaylistHTML(body, baseURL); len(entries) > 0 {
			return entries, nil
		}

		return parsePlaylistLines(body)
	}

	switch extension {
	case ".m3u", ".m3u8", ".txt", ".list", ".pls":
		return parsePlaylistLines(body)
	case ".csv":
		return parsePlaylistCSV(body)
	case ".json":
		return parsePlaylistJSON(body)
	}

	if looksLikeHTML(body) {
		if entries := parsePlaylistHTML(body, baseURL); len(entries) > 0 {
			return entries, nil
		}
	}

	entries, err := parsePlaylistLines(body)
	if err != nil {
		return nil, err
	}

	if len(entries) > 0 {
		return entries, nil
	}

	return parsePlaylistHTML(body, baseURL), nil
}

func parsePlaylistLines(body []byte) ([]PlaylistEntry, error) {
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	entries := make([]PlaylistEntry, 0)
	pendingTitle := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		upperLine := strings.ToUpper(line)
		if strings.HasPrefix(upperLine, "#EXTINF:") {
			if commaIndex := strings.Index(line, ","); commaIndex >= 0 && commaIndex+1 < len(line) {
				pendingTitle = strings.TrimSpace(line[commaIndex+1:])
			}

			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		entry := PlaylistEntry{Query: line}
		if pendingTitle != "" {
			entry.Title = pendingTitle
			pendingTitle = ""
		}

		entries = append(entries, entry)
		if len(entries) >= playlistMaxEntries {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan playlist lines: %w", err)
	}

	return entries, nil
}

func parsePlaylistCSV(body []byte) ([]PlaylistEntry, error) {
	reader := csv.NewReader(bytes.NewReader(body))
	reader.FieldsPerRecord = -1

	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv playlist: %w", err)
	}

	entries := make([]PlaylistEntry, 0, len(rows))
	for _, row := range rows {
		entry := parseCSVRow(row)
		if strings.TrimSpace(entry.Query) == "" {
			continue
		}

		entries = append(entries, entry)
		if len(entries) >= playlistMaxEntries {
			break
		}
	}

	return entries, nil
}

func parseCSVRow(row []string) PlaylistEntry {
	cleaned := make([]string, 0, len(row))
	for _, cell := range row {
		value := strings.TrimSpace(cell)
		if value != "" {
			cleaned = append(cleaned, value)
		}
	}

	if len(cleaned) == 0 {
		return PlaylistEntry{}
	}

	if len(cleaned) == 1 {
		return PlaylistEntry{Query: cleaned[0]}
	}

	firstLooksLikeURL := looksLikeURL(cleaned[0])
	secondLooksLikeURL := looksLikeURL(cleaned[1])

	switch {
	case firstLooksLikeURL && !secondLooksLikeURL:
		return PlaylistEntry{Title: cleaned[1], Query: cleaned[0]}
	case !firstLooksLikeURL && secondLooksLikeURL:
		return PlaylistEntry{Title: cleaned[0], Query: cleaned[1]}
	default:
		return PlaylistEntry{Title: cleaned[0], Query: cleaned[1]}
	}
}

func parsePlaylistJSON(body []byte) ([]PlaylistEntry, error) {
	var payload any
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("parse json playlist: %w", err)
	}

	entries := flattenPlaylistJSON(payload)
	if len(entries) == 0 {
		return nil, fmt.Errorf("json playlist did not contain any tracks")
	}

	return limitPlaylistEntries(entries), nil
}

func flattenPlaylistJSON(payload any) []PlaylistEntry {
	switch value := payload.(type) {
	case []any:
		return flattenPlaylistJSONArray(value)
	case map[string]any:
		for _, key := range []string{"tracks", "items", "songs", "playlist"} {
			if nested, ok := value[key]; ok {
				if entries := flattenPlaylistJSON(nested); len(entries) > 0 {
					return entries
				}
			}
		}

		entry := playlistEntryFromMap(value)
		if strings.TrimSpace(entry.Query) == "" {
			return nil
		}

		return []PlaylistEntry{entry}
	case string:
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return nil
		}

		return []PlaylistEntry{{Query: trimmed}}
	default:
		return nil
	}
}

func flattenPlaylistJSONArray(values []any) []PlaylistEntry {
	entries := make([]PlaylistEntry, 0, len(values))

	for _, item := range values {
		switch value := item.(type) {
		case string:
			trimmed := strings.TrimSpace(value)
			if trimmed == "" {
				continue
			}

			entries = append(entries, PlaylistEntry{Query: trimmed})
		case map[string]any:
			entry := playlistEntryFromMap(value)
			if strings.TrimSpace(entry.Query) == "" {
				continue
			}

			entries = append(entries, entry)
		default:
			continue
		}

		if len(entries) >= playlistMaxEntries {
			break
		}
	}

	return entries
}

func playlistEntryFromMap(values map[string]any) PlaylistEntry {
	title := firstStringValue(values, "title", "name", "label")
	query := firstStringValue(values, "query", "url", "link", "href", "source")

	if query == "" {
		query = title
	}

	if title == "" {
		title = query
	}

	return PlaylistEntry{
		Title: strings.TrimSpace(title),
		Query: strings.TrimSpace(query),
	}
}

func firstStringValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		rawValue, ok := values[key]
		if !ok {
			continue
		}

		stringValue, ok := rawValue.(string)
		if !ok {
			continue
		}

		trimmed := strings.TrimSpace(stringValue)
		if trimmed != "" {
			return trimmed
		}
	}

	return ""
}

func parsePlaylistHTML(body []byte, baseURL *neturl.URL) []PlaylistEntry {
	matches := hrefPattern.FindAllSubmatch(body, -1)
	if len(matches) == 0 {
		return nil
	}

	entries := make([]PlaylistEntry, 0, len(matches))
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		resolvedURL := resolveReference(baseURL, string(match[1]))
		if resolvedURL == "" {
			continue
		}

		entries = append(entries, PlaylistEntry{Query: resolvedURL})
		if len(entries) >= playlistMaxEntries {
			break
		}
	}

	return entries
}

func resolveReference(baseURL *neturl.URL, rawReference string) string {
	trimmed := strings.TrimSpace(rawReference)
	if trimmed == "" {
		return ""
	}

	lower := strings.ToLower(trimmed)
	if strings.HasPrefix(lower, "#") || strings.HasPrefix(lower, "javascript:") || strings.HasPrefix(lower, "mailto:") {
		return ""
	}

	reference, err := neturl.Parse(trimmed)
	if err != nil {
		return ""
	}

	if baseURL == nil {
		return reference.String()
	}

	return baseURL.ResolveReference(reference).String()
}

func looksLikeHTML(body []byte) bool {
	lowered := strings.ToLower(string(body))
	return strings.Contains(lowered, "<html") || strings.Contains(lowered, "<a ") || strings.Contains(lowered, "<body")
}

func looksLikeURL(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	parsed, err := neturl.Parse(trimmed)
	if err != nil {
		return false
	}

	return parsed.Scheme != "" || strings.HasPrefix(trimmed, "www.")
}

func limitPlaylistEntries(entries []PlaylistEntry) []PlaylistEntry {
	if len(entries) <= playlistMaxEntries {
		return entries
	}

	return append([]PlaylistEntry(nil), entries[:playlistMaxEntries]...)
}

var hrefPattern = regexp.MustCompile(`(?i)href\s*=\s*["']([^"'#]+)["']`)
