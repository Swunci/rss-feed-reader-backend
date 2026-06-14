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
	"github.com/mmcdole/gofeed"
)

type DiscoverService struct {
	feedRepo    *repositories.FeedRepo
	itemRepo    *repositories.ItemRepo
	feedService *FeedService
}

func NewDiscoverService(fr *repositories.FeedRepo, ir *repositories.ItemRepo, is *FeedService) *DiscoverService {
	return &DiscoverService{feedRepo: fr, itemRepo: ir, feedService: is}
}

func (s *DiscoverService) DiscoverFeeds(url string) ([]models.DiscoverFeed, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	if isValidRSSLink(url) {
		slog.Debug("Valid RSS link, skipping discovery", "url", url)
		return []models.DiscoverFeed{}, nil
	}
	discovered_feeds := []models.DiscoverFeed{}
	if isYouTubeURL(url) {
		discovered_feeds, err := getYouTubeFeeds(url)
		return discovered_feeds, err
	}
	if isRedditURL(url) {
		discovered_feeds, err := getRedditFeeds(url)
		return discovered_feeds, err
	}
	return discovered_feeds, nil
}

func isValidRSSLink(url string) bool {
	if isRedditURL(url) {
		url = url + "?user=Positive_Ear1287&feed=fa6c8aa5fdc3af2f011b2cdc6cec7be7ec664436"
	}
	fp := gofeed.NewParser()
	_, err := fp.ParseURL(url)
	return err == nil
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

type YouTubeChannel struct {
	ID   string
	Name string
}

func extractChannelInfo(channelURL string) (YouTubeChannel, error) {
	re := regexp.MustCompile(`(?:www\.)?youtube\.com/channel/(UC[\w-]+)`)
	resp, err := http.Get(channelURL)
	if err != nil {
		slog.Error("Fetch YouTube channel", "url", channelURL, "err", err)
		return YouTubeChannel{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return YouTubeChannel{}, err
	}

	idMatch := re.FindSubmatch(body)
	if len(idMatch) < 2 {
		slog.Error("YouTube channel ID not found", "url", channelURL)
		return YouTubeChannel{}, fmt.Errorf("channel ID not found")
	}

	nameRe := regexp.MustCompile(`<meta property="og:title" content="([^"]+)"`)
	nameMatch := nameRe.FindSubmatch(body)
	name := ""
	if len(nameMatch) > 1 {
		name = string(nameMatch[1])
	}

	return YouTubeChannel{
		ID:   string(idMatch[1]),
		Name: name,
	}, nil
}

func getYouTubeFeeds(channelURL string) ([]models.DiscoverFeed, error) {
	youtube_channel, err := extractChannelInfo(channelURL)
	if err != nil {
		return nil, err
	}

	id := youtube_channel.ID[2:]
	base := "https://www.youtube.com/feeds/videos.xml?playlist_id="

	options := []models.DiscoverFeed{
		{Name: youtube_channel.Name, URL: fmt.Sprintf("https://www.youtube.com/feeds/videos.xml?channel_id=%s", youtube_channel.ID)},
		{Name: fmt.Sprintf("%s - Videos", youtube_channel.Name), URL: fmt.Sprintf("%sUULF%s", base, id)},
		{Name: fmt.Sprintf("%s - Shorts", youtube_channel.Name), URL: fmt.Sprintf("%sUUSH%s", base, id)},
		{Name: fmt.Sprintf("%s - Live", youtube_channel.Name), URL: fmt.Sprintf("%sUULV%s", base, id)},
	}

	var feeds []models.DiscoverFeed
	for _, o := range options {
		if isValidRSSLink(o.URL) {
			feeds = append(feeds, o)
		}
	}
	slog.Debug("YouTube feeds discovered", "channel", youtube_channel.Name, "count", len(feeds))
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
			{"", ""},
			{"Posts", "/submitted"},
			{"Comments", "/comments"},
		}
	default:
		return nil, fmt.Errorf("unsupported reddit URL")
	}

	base := fmt.Sprintf("https://www.reddit.com%s", path)

	var feeds []models.DiscoverFeed
	for _, s := range options {
		url := fmt.Sprintf("%s%s.rss", base, s.path)
		if isValidRSSLink(url) {
			feeds = append(feeds, models.DiscoverFeed{
				Name: fmt.Sprintf("%s - %s", name, s.label),
				URL:  url,
			})
		}
	}
	slog.Debug("Reddit feeds discovered", "url", redditURL, "count", len(feeds))
	return feeds, nil
}
