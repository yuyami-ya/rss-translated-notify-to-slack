package service

import (
	"log"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedService はRSSフィードの監視を管理する
type FeedService struct {
	feedURLs           []string
	maxArticlesPerFeed int
	parser             *gofeed.Parser
}

// FeedItem は処理対象のフィードアイテム
type FeedItem struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
	GUID        string
	FeedURL     string // どのフィードからの記事かを識別
}

// NewFeedService は新しいFeedServiceを作成する
func NewFeedService(feedURLs []string, maxArticlesPerFeed int) *FeedService {
	return &FeedService{
		feedURLs:           feedURLs,
		maxArticlesPerFeed: maxArticlesPerFeed,
		parser:             gofeed.NewParser(),
	}
}

// CheckForRecentItems は過去24時間以内の新しいRSSアイテムをチェックする
func (fs *FeedService) CheckForRecentItems() ([]*FeedItem, error) {
	log.Printf("Checking %d RSS feeds for recent items", len(fs.feedURLs))
	
	since := time.Now().Add(-24 * time.Hour)
	var allRecentItems []*FeedItem
	
	for _, feedURL := range fs.feedURLs {
		log.Printf("Checking RSS feed: %s", feedURL)
		
		// RSSフィードを取得
		feed, err := fs.parser.ParseURL(feedURL)
		if err != nil {
			log.Printf("Failed to parse RSS feed %s: %v", feedURL, err)
			continue // エラーがあっても他のフィードは処理を続ける
		}

		log.Printf("Found %d items in RSS feed: %s", len(feed.Items), feedURL)

		var recentItems []*FeedItem
		
		// 各アイテムをチェック（最大件数まで）
		for i, item := range feed.Items {
			if i >= fs.maxArticlesPerFeed {
				log.Printf("Reached max articles limit (%d) for feed: %s", fs.maxArticlesPerFeed, feedURL)
				break
			}

			if item == nil {
				continue
			}

			// 記事の公開日時をチェック
			var publishedTime time.Time
			if item.PublishedParsed != nil {
				publishedTime = *item.PublishedParsed
			} else if item.UpdatedParsed != nil {
				publishedTime = *item.UpdatedParsed
			} else {
				// 日付情報がない場合は現在時刻を使用（安全側に倒す）
				publishedTime = time.Now()
			}

			// 過去24時間以内の記事のみ処理
			if publishedTime.After(since) {
				// アイテムのユニークIDを生成（GUID or Link）
				guid := item.GUID
				if guid == "" {
					guid = item.Link
				}

				feedItem := &FeedItem{
					Title:       cleanText(item.Title),
					Description: cleanText(item.Description),
					Link:        item.Link,
					Published:   publishedTime,
					GUID:        guid,
					FeedURL:     feedURL,
				}

				recentItems = append(recentItems, feedItem)
				log.Printf("Recent item found: %s (published: %s)", feedItem.Title, publishedTime.Format("2006-01-02 15:04:05"))
			}
		}

		log.Printf("Found %d recent items (within 24h) from feed: %s", len(recentItems), feedURL)
		allRecentItems = append(allRecentItems, recentItems...)
	}

	log.Printf("Total recent items found across all feeds: %d", len(allRecentItems))
	return allRecentItems, nil
}

// cleanText はテキストから不要な文字を除去する
func cleanText(text string) string {
	// HTMLタグを除去（簡易版）
	text = strings.ReplaceAll(text, "<br>", "\n")
	text = strings.ReplaceAll(text, "<br/>", "\n")
	text = strings.ReplaceAll(text, "<br />", "\n")
	
	// その他のHTMLタグを除去（より高度な処理が必要な場合は html.UnescapeString や goquery を使用）
	for strings.Contains(text, "<") && strings.Contains(text, ">") {
		start := strings.Index(text, "<")
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}

	// 余分な空白を除去
	text = strings.TrimSpace(text)
	lines := strings.Split(text, "\n")
	var cleanLines []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			cleanLines = append(cleanLines, line)
		}
	}

	return strings.Join(cleanLines, "\n")
}

// GetFeedInfo はフィードの基本情報を取得する（デバッグ用）
func (fs *FeedService) GetFeedInfo(feedURL string) (*gofeed.Feed, error) {
	return fs.parser.ParseURL(feedURL)
}