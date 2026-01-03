package enricher

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/danielmerrison/learnd/internal/model"
)

var (
	// YouTube URL patterns
	youtubePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/shorts/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`),
	}

	// ISO 8601 duration pattern (PT#H#M#S)
	durationPattern = regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
)

// YouTubeEnricher extracts metadata from YouTube videos using the Data API v3
type YouTubeEnricher struct {
	apiKey string
	client *http.Client
}

// NewYouTubeEnricher creates a new YouTube enricher
func NewYouTubeEnricher(apiKey string) *YouTubeEnricher {
	return &YouTubeEnricher{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (e *YouTubeEnricher) Name() string  { return "youtube" }
func (e *YouTubeEnricher) Priority() int { return 10 }

func (e *YouTubeEnricher) CanHandle(rawURL string) bool {
	return extractVideoID(rawURL) != ""
}

func (e *YouTubeEnricher) Enrich(ctx context.Context, rawURL string) (*Result, error) {
	videoID := extractVideoID(rawURL)
	if videoID == "" {
		return nil, fmt.Errorf("could not extract video ID from URL")
	}

	// Build API request
	apiURL := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/videos?id=%s&part=snippet,contentDetails&key=%s",
		url.QueryEscape(videoID),
		url.QueryEscape(e.apiKey),
	)

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := e.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch YouTube API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTube API error: %d", resp.StatusCode)
	}

	var apiResp youtubeAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(apiResp.Items) == 0 {
		return nil, fmt.Errorf("video not found")
	}

	item := apiResp.Items[0]
	snippet := item.Snippet
	contentDetails := item.ContentDetails

	// Parse duration
	var runtimeSeconds *int
	if duration := parseDuration(contentDetails.Duration); duration > 0 {
		runtimeSeconds = &duration
	}

	// Parse published date
	var publishedAt *time.Time
	if t, err := time.Parse(time.RFC3339, snippet.PublishedAt); err == nil {
		publishedAt = &t
	}

	return &Result{
		CanonicalURL:   fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID),
		Domain:         "youtube.com",
		SourceType:     model.SourceTypeYouTube,
		Title:          snippet.Title,
		Description:    truncateDescription(snippet.Description),
		PublishedAt:    publishedAt,
		RuntimeSeconds: runtimeSeconds,
		Metadata: map[string]interface{}{
			"channel_title": snippet.ChannelTitle,
			"channel_id":    snippet.ChannelID,
			"video_id":      videoID,
		},
	}, nil
}

// extractVideoID extracts the video ID from various YouTube URL formats
func extractVideoID(rawURL string) string {
	for _, pattern := range youtubePatterns {
		if matches := pattern.FindStringSubmatch(rawURL); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}

// parseDuration converts ISO 8601 duration to seconds
func parseDuration(duration string) int {
	matches := durationPattern.FindStringSubmatch(duration)
	if len(matches) == 0 {
		return 0
	}

	var hours, minutes, seconds int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(matches[3])
	}

	return hours*3600 + minutes*60 + seconds
}

// truncateDescription limits description length for storage
func truncateDescription(desc string) string {
	// Take first 500 characters
	if len(desc) > 500 {
		// Try to break at a sentence boundary
		if idx := strings.LastIndex(desc[:500], ". "); idx > 200 {
			return desc[:idx+1]
		}
		return desc[:500] + "..."
	}
	return desc
}

// YouTube API response structures
type youtubeAPIResponse struct {
	Items []youtubeVideoItem `json:"items"`
}

type youtubeVideoItem struct {
	Snippet        youtubeSnippet        `json:"snippet"`
	ContentDetails youtubeContentDetails `json:"contentDetails"`
}

type youtubeSnippet struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	ChannelTitle string `json:"channelTitle"`
	ChannelID    string `json:"channelId"`
	PublishedAt  string `json:"publishedAt"`
}

type youtubeContentDetails struct {
	Duration string `json:"duration"`
}
