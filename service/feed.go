package service

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

// FeedService はRSSフィードの監視を管理する
type FeedService struct {
	feedURL     string
	parser      *gofeed.Parser
	lastChecked map[string]bool // フィードアイテムの重複チェック用
	stateFile   string          // 最後にチェックした記事の状態を保存するファイル
}

// FeedItem は処理対象のフィードアイテム
type FeedItem struct {
	Title       string
	Description string
	Link        string
	Published   time.Time
	GUID        string
}

// NewFeedService は新しいFeedServiceを作成する
func NewFeedService(feedURL string) *FeedService {
	return &FeedService{
		feedURL:     feedURL,
		parser:      gofeed.NewParser(),
		lastChecked: make(map[string]bool),
		stateFile:   "last_checked_state.txt",
	}
}

// CheckForNewItems は新しいRSSアイテムをチェックする
func (fs *FeedService) CheckForNewItems() ([]*FeedItem, error) {
	log.Printf("Checking RSS feed: %s", fs.feedURL)
	
	// RSSフィードを取得
	feed, err := fs.parser.ParseURL(fs.feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS feed: %w", err)
	}

	log.Printf("Found %d items in RSS feed", len(feed.Items))

	// 前回の状態を読み込み
	if err := fs.loadState(); err != nil {
		log.Printf("Warning: failed to load state: %v", err)
	}

	var newItems []*FeedItem
	
	// 各アイテムをチェック
	for _, item := range feed.Items {
		if item == nil {
			continue
		}

		// アイテムのユニークIDを生成（GUID or Link）
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		// 既にチェック済みのアイテムはスキップ
		if fs.lastChecked[guid] {
			continue
		}

		// 新しいアイテムとして追加
		feedItem := &FeedItem{
			Title:       cleanText(item.Title),
			Description: cleanText(item.Description),
			Link:        item.Link,
			GUID:        guid,
		}

		// 公開日時を解析
		if item.PublishedParsed != nil {
			feedItem.Published = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			feedItem.Published = *item.UpdatedParsed
		} else {
			feedItem.Published = time.Now()
		}

		newItems = append(newItems, feedItem)
		fs.lastChecked[guid] = true

		log.Printf("New item found: %s", feedItem.Title)
	}

	// 状態を保存
	if err := fs.saveState(); err != nil {
		log.Printf("Warning: failed to save state: %v", err)
	}

	log.Printf("Found %d new items", len(newItems))
	return newItems, nil
}

// loadState は前回チェック済みのアイテムの状態を読み込む
func (fs *FeedService) loadState() error {
	if _, err := os.Stat(fs.stateFile); os.IsNotExist(err) {
		// ファイルが存在しない場合は初回実行として処理
		return nil
	}

	data, err := os.ReadFile(fs.stateFile)
	if err != nil {
		return fmt.Errorf("failed to read state file: %w", err)
	}

	// 改行で分割してGUIDのリストとして読み込み
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			fs.lastChecked[line] = true
		}
	}

	log.Printf("Loaded %d checked items from state file", len(fs.lastChecked))
	return nil
}

// saveState は現在のチェック状態を保存する
func (fs *FeedService) saveState() error {
	// ディレクトリが存在しない場合は作成
	if err := os.MkdirAll(filepath.Dir(fs.stateFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	var guids []string
	for guid := range fs.lastChecked {
		guids = append(guids, guid)
	}

	// 最新1000件のみ保持（メモリ効率化）
	if len(guids) > 1000 {
		// 新しいマップを作成
		newLastChecked := make(map[string]bool)
		for i := len(guids) - 1000; i < len(guids); i++ {
			newLastChecked[guids[i]] = true
		}
		fs.lastChecked = newLastChecked
		guids = guids[len(guids)-1000:]
	}

	data := strings.Join(guids, "\n")
	if err := os.WriteFile(fs.stateFile, []byte(data), 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
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
func (fs *FeedService) GetFeedInfo() (*gofeed.Feed, error) {
	return fs.parser.ParseURL(fs.feedURL)
}