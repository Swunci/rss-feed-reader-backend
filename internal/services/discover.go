package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/Swunci/rss-feed-backend/internal/models"
	"github.com/Swunci/rss-feed-backend/internal/repositories"
)

type DiscoverService struct {
	feedRepo    *repositories.FeedRepo
	itemRepo    *repositories.ItemRepo
	feedService *FeedService
	logger      *slog.Logger
}

func NewDiscoverService(fr *repositories.FeedRepo, ir *repositories.ItemRepo, is *FeedService, logger *slog.Logger) *DiscoverService {
	return &DiscoverService{feedRepo: fr, itemRepo: ir, feedService: is, logger: logger}
}

func (s *DiscoverService) DiscoverFeeds(url string) (models.DiscoverResponse, error) {
	result := models.DiscoverResponse{}
	feed, err := s.feedService.CreateFeed(url)
	if err == nil {
		result.Feed = &feed
		return result, err
	}

	if isYouTubeURL(url) {
		result.Feeds, err = getYouTubeFeeds(url)
		return result, err
	}
	if isRedditURL(url) {
		result.Feeds, err = getRedditFeeds(url)
		return result, err
	}
	return result, nil

}

func isYouTubeURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	return host == "youtube.com" || host == "www.youtube.com"
}

func isRedditURL(u string) bool {
	parsed, err := url.Parse(u)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	return host == "reddit.com" || host == "www.reddit.com"
}

func extractChannelID(channelURL string) (string, error) {
	re := regexp.MustCompile(`(?:www\.)?youtube\.com/channel/(UC[\w-]+)`)
	if match := re.FindStringSubmatch(channelURL); len(match) > 1 {
		return match[1], nil
	}

	resp, err := http.Get(channelURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	match := re.FindSubmatch(body)
	if len(match) < 2 {
		return "", fmt.Errorf("channel ID not found")
	}
	return string(match[1]), nil
}

func getYouTubeFeeds(channelURL string) ([]models.DiscoverFeed, error) {
	channelID, err := extractChannelID(channelURL)
	if err != nil {
		return nil, err
	}

	id := channelID[2:]
	base := "https://www.youtube.com/feeds/videos.xml?playlist_id="

	feeds := []models.DiscoverFeed{
		{Title: "All", URL: fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", channelID)},
		{Title: "Videos", URL: fmt.Sprintf("%sUULF%s", base, id)},
		{Title: "Shorts", URL: fmt.Sprintf("%sUUSH%s", base, id)},
		{Title: "Live", URL: fmt.Sprintf("%sUULV%s", base, id)},
	}

	return feeds, nil
}

func getRedditFeeds(redditURL string) ([]models.DiscoverFeed, error) {
	parsed, err := url.Parse(redditURL)
	if err != nil {
		return nil, fmt.Errorf("invalid reddit URL")
	}

	path := strings.TrimSuffix(parsed.Path, "/")

	var name string

	var options []struct {
		label string
		path  string
	}
	switch {
	case strings.HasPrefix(path, "/r/"):
		name = path[3:]
		options = []struct {
			label string
			path  string
		}{
			{"Best", ""},
			{"Hot", "/hot"},
			{"New", "/new"},
			{"Top", "/top"},
			{"Rising", "/rising"},
		}
	case strings.HasPrefix(path, "/u/"), strings.HasPrefix(path, "/user/"):
		parts := strings.Split(path, "/")
		name = parts[len(parts)-1]
		options = []struct {
			label string
			path  string
		}{
			{"Overview", ""},
			{"Submitted", "/submitted"},
			{"Comments", "/comments"},
		}
	default:
		return nil, fmt.Errorf("unsupported reddit URL")
	}

	base := fmt.Sprintf("https://www.reddit.com%s", path)

	var feeds []models.DiscoverFeed
	for _, s := range options {
		feeds = append(feeds, models.DiscoverFeed{
			Title: fmt.Sprintf("%s - %s", name, s.label),
			URL:   fmt.Sprintf("%s%s.rss", base, s.path),
		})
	}

	return feeds, nil
}
