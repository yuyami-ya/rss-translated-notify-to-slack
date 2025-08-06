package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// NotificationService はSlack通知を管理する
type NotificationService struct {
	webhookURL string
	channel    string
	httpClient *http.Client
}

// SlackMessage はSlackに送信するメッセージの構造体
type SlackMessage struct {
	Channel     string       `json:"channel,omitempty"`
	Username    string       `json:"username,omitempty"`
	IconEmoji   string       `json:"icon_emoji,omitempty"`
	Text        string       `json:"text,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	ThreadTS    string       `json:"thread_ts,omitempty"` // スレッドタイムスタンプ
}

// SlackResponse はSlackからのレスポンス構造体
type SlackResponse struct {
	OK        bool   `json:"ok"`
	Error     string `json:"error,omitempty"`
	Timestamp string `json:"ts,omitempty"` // メッセージのタイムスタンプ
}

// Attachment はSlackメッセージの添付ファイル構造体
type Attachment struct {
	Color      string  `json:"color,omitempty"`
	Title      string  `json:"title,omitempty"`
	TitleLink  string  `json:"title_link,omitempty"`
	Text       string  `json:"text,omitempty"`
	Fields     []Field `json:"fields,omitempty"`
	Footer     string  `json:"footer,omitempty"`
	Timestamp  int64   `json:"ts,omitempty"`
	MarkdownIn []string `json:"mrkdwn_in,omitempty"`
}

// Field はSlackメッセージの添付フィールド構造体
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// NewNotificationService は新しいNotificationServiceを作成する
func NewNotificationService(webhookURL, channel string) *NotificationService {
	return &NotificationService{
		webhookURL: webhookURL,
		channel:    channel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SendNewArticleNotification は新記事の通知を送信する
func (ns *NotificationService) SendNewArticleNotification(result *TranslationResult) error {
	log.Printf("Sending Slack notification for article: %s", result.TranslatedTitle)

	// Slackメッセージを構築
	message := ns.buildArticleMessage(result)

	// Slackに送信
	if err := ns.sendToSlack(message); err != nil {
		return fmt.Errorf("failed to send Slack notification: %w", err)
	}

	log.Printf("Slack notification sent successfully for: %s", result.TranslatedTitle)
	return nil
}

// SendNewArticleNotificationWithThread は新記事の通知をスレッド形式で送信する
func (ns *NotificationService) SendNewArticleNotificationWithThread(result *TranslationResult) error {
	log.Printf("Sending threaded Slack notification for article: %s", result.TranslatedTitle)

	// 1. まずタイトルメッセージを送信
	titleMessage := ns.buildTitleMessage(result)
	timestamp, err := ns.sendToSlackWithResponse(titleMessage)
	if err != nil {
		return fmt.Errorf("failed to send title message: %w", err)
	}

	// 2. 要約をスレッドで返信
	summaryMessage := ns.buildSummaryMessage(result)
	summaryMessage.ThreadTS = timestamp // スレッドに関連付け

	if err := ns.sendToSlack(summaryMessage); err != nil {
		return fmt.Errorf("failed to send summary in thread: %w", err)
	}

	log.Printf("Threaded Slack notification sent successfully for: %s", result.TranslatedTitle)
	return nil
}

// SendErrorNotification はエラー通知を送信する
func (ns *NotificationService) SendErrorNotification(errorMsg string) error {
	log.Printf("Sending error notification to Slack: %s", errorMsg)

	message := &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":warning:",
		Attachments: []Attachment{
			{
				Color: "danger",
				Title: "RSS通知システムでエラーが発生しました",
				Text:  errorMsg,
				Fields: []Field{
					{
						Title: "発生日時",
						Value: time.Now().In(time.FixedZone("JST", 9*60*60)).Format("2006-01-02 15:04:05 JST"),
						Short: true,
					},
				},
				Footer:     "RSS通知システム",
				Timestamp:  time.Now().Unix(),
				MarkdownIn: []string{"text"},
			},
		},
	}

	if err := ns.sendToSlack(message); err != nil {
		return fmt.Errorf("failed to send error notification: %w", err)
	}

	return nil
}

// SendStartupNotification はシステム起動通知を送信する
func (ns *NotificationService) SendStartupNotification() error {
	log.Println("Sending startup notification to Slack")

	message := &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":rocket:",
		Attachments: []Attachment{
			{
				Color: "good",
				Title: "RSS通知システムが開始されました",
				Text:  "ByteByteGoのRSSフィード監視を開始します。",
				Fields: []Field{
					{
						Title: "開始日時",
						Value: time.Now().In(time.FixedZone("JST", 9*60*60)).Format("2006-01-02 15:04:05 JST"),
						Short: true,
					},
					{
						Title: "フィードURL",
						Value: "https://blog.bytebytego.com/feed",
						Short: true,
					},
				},
				Footer:     "RSS通知システム",
				Timestamp:  time.Now().Unix(),
				MarkdownIn: []string{"text"},
			},
		},
	}

	return ns.sendToSlack(message)
}

// buildArticleMessage は記事通知用のSlackメッセージを構築する
func (ns *NotificationService) buildArticleMessage(result *TranslationResult) *SlackMessage {
	// 説明文を短縮（Slackの制限に対応）
	description := result.TranslatedDescription
	if len(description) > 500 {
		description = description[:500] + "..."
	}

	// 要約文の整形
	summary := result.Summary
	if summary == "" {
		summary = "要約が利用できません。"
	}

	return &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":newspaper:",
		Text:      " *ByteByteGoの新しい記事が投稿されました！*",
		Attachments: []Attachment{
			{
				Color:     "#36a64f",
				Title:     result.TranslatedTitle,
				TitleLink: result.Link,
				Text:      fmt.Sprintf("* 要約*\n%s", summary),
				Fields: []Field{
					{
						Title: "原文タイトル",
						Value: result.OriginalTitle,
						Short: false,
					},
					{
						Title: "詳細",
						Value: ns.truncateText(description, 300),
						Short: false,
					},
				},
				Footer: "ByteByteGo RSS通知",
				Timestamp: time.Now().Unix(),
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}
}

// sendToSlack はSlackにメッセージを送信する
func (ns *NotificationService) sendToSlack(message *SlackMessage) error {
	// JSONにエンコード
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", ns.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Slack API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Slackからの "ok" レスポンスをチェック
	if strings.TrimSpace(string(body)) != "ok" {
		return fmt.Errorf("unexpected Slack response: %s", string(body))
	}

	return nil
}

// sendToSlackWithResponse はSlackにメッセージを送信し、タイムスタンプを返す
func (ns *NotificationService) sendToSlackWithResponse(message *SlackMessage) (string, error) {
	// JSONにエンコード
	jsonData, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", ns.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := ns.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// ステータスコードをチェック
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Slack API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// Webhook形式の場合は "ok" レスポンスなので、現在時刻をタイムスタンプとして使用
	// 注意: Webhook URLではメッセージタイムスタンプを取得できないため、
	// 実際のBot TokenベースのAPIが必要な場合は別実装が必要
	timestamp := fmt.Sprintf("%.6f", float64(time.Now().Unix())+float64(time.Now().Nanosecond())/1e9)
	
	// Slackからの "ok" レスポンスをチェック
	if strings.TrimSpace(string(body)) != "ok" {
		return "", fmt.Errorf("unexpected Slack response: %s", string(body))
	}

	return timestamp, nil
}

// TestSlackConnection はSlack Webhookの接続をテストする
func (ns *NotificationService) TestSlackConnection() error {
	log.Println("Testing Slack connection...")

	message := &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":white_check_mark:",
		Text:      " RSS通知システムの接続テストです。このメッセージが表示されていれば正常に動作しています。",
	}

	return ns.sendToSlack(message)
}

// truncateText は指定した長さでテキストを切り詰める
func (ns *NotificationService) truncateText(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	
	// 単語の境界で切り詰める
	truncated := text[:maxLen]
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}
	
	return truncated + "..."
}

// SendBatchNotification は複数の記事をまとめて通知する
func (ns *NotificationService) SendBatchNotification(results []*TranslationResult) error {
	if len(results) == 0 {
		return nil
	}

	log.Printf("Sending batch notification for %d articles", len(results))

	// バッチ通知のメッセージを構築
	var attachments []Attachment

	// ヘッダー添付
	headerAttachment := Attachment{
		Color: "#36a64f",
		Title: fmt.Sprintf(" ByteByteGoに %d 件の新しい記事が投稿されました！", len(results)),
		Footer: "ByteByteGo RSS通知",
		Timestamp: time.Now().Unix(),
		MarkdownIn: []string{"text"},
	}
	attachments = append(attachments, headerAttachment)

	// 各記事の添付（最大5件まで）
	maxArticles := 5
	if len(results) > maxArticles {
		results = results[:maxArticles]
	}

	for i, result := range results {
		attachment := Attachment{
			Color:     "#2196F3",
			Title:     result.TranslatedTitle,
			TitleLink: result.Link,
			Text:      result.Summary,
			Fields: []Field{
				{
					Title: "原文タイトル",
					Value: result.OriginalTitle,
					Short: false,
				},
			},
			MarkdownIn: []string{"text"},
		}

		// 最後の記事以外は区切り線を追加
		if i < len(results)-1 {
			attachment.Text += "\n---"
		}

		attachments = append(attachments, attachment)
	}

	message := &SlackMessage{
		Channel:     ns.channel,
		Username:    "RSS通知Bot",
		IconEmoji:   ":newspaper:",
		Attachments: attachments,
	}

	return ns.sendToSlack(message)
}

// buildTitleMessage はタイトル投稿用のSlackメッセージを構築する
func (ns *NotificationService) buildTitleMessage(result *TranslationResult) *SlackMessage {
	return &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":newspaper:",
		Text:      " *ByteByteGoの新しい記事が投稿されました！*",
		Attachments: []Attachment{
			{
				Color:     "#36a64f",
				Title:     result.TranslatedTitle,
				TitleLink: result.Link,
				Fields: []Field{
					{
						Title: "原文タイトル",
						Value: result.OriginalTitle,
						Short: false,
					},
				},
				Footer:     "ByteByteGo RSS通知 - 要約は下記スレッドをご確認ください 👇",
				Timestamp:  time.Now().Unix(),
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}
}

// buildSummaryMessage は要約投稿用のSlackメッセージを構築する
func (ns *NotificationService) buildSummaryMessage(result *TranslationResult) *SlackMessage {
	// 説明文を短縮（Slackの制限に対応）
	description := result.TranslatedDescription
	if len(description) > 800 {
		description = description[:800] + "..."
	}

	// 要約文の整形
	summary := result.Summary
	if summary == "" {
		summary = "要約が利用できません。"
	}

	return &SlackMessage{
		Channel:   ns.channel,
		Username:  "RSS通知Bot",
		IconEmoji: ":memo:",
		Text:      fmt.Sprintf(" **記事要約**\n%s", summary),
		Attachments: []Attachment{
			{
				Color: "#2196F3",
				Title: "詳細内容",
				Text:  ns.truncateText(description, 600),
				Fields: []Field{
					{
						Title: "記事リンク",
						Value: fmt.Sprintf("<%s|記事を読む>", result.Link),
						Short: true,
					},
				},
				Footer:     "ByteByteGo RSS通知",
				Timestamp:  time.Now().Unix(),
				MarkdownIn: []string{"text", "fields"},
			},
		},
	}
}