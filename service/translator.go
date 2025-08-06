package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sashabaranov/go-openai"
)

// TranslatorService は翻訳サービスを管理する
type TranslatorService struct {
	deepLAPIKey   string
	deepLAPIURL   string
	openAIClient  *openai.Client
	openAIModel   string
	httpClient    *http.Client
}

// DeepLRequest はDeepL APIのリクエスト構造体
type DeepLRequest struct {
	Text       []string `json:"text"`
	TargetLang string   `json:"target_lang"`
	SourceLang string   `json:"source_lang,omitempty"`
}

// DeepLResponse はDeepL APIのレスポンス構造体
type DeepLResponse struct {
	Translations []struct {
		DetectedSourceLanguage string `json:"detected_source_language"`
		Text                   string `json:"text"`
	} `json:"translations"`
}

// TranslationResult は翻訳結果を表す構造体
type TranslationResult struct {
	OriginalTitle       string
	TranslatedTitle     string
	OriginalDescription string
	TranslatedDescription string
	Summary             string
	Link                string
}

// NewTranslatorService は新しいTranslatorServiceを作成する
func NewTranslatorService(deepLAPIKey, deepLAPIURL, openAIAPIKey, openAIModel string) *TranslatorService {
	return &TranslatorService{
		deepLAPIKey:  deepLAPIKey,
		deepLAPIURL:  deepLAPIURL,
		openAIClient: openai.NewClient(openAIAPIKey),
		openAIModel:  openAIModel,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TranslateAndSummarize は記事を翻訳し要約を生成する
func (ts *TranslatorService) TranslateAndSummarize(item *FeedItem) (*TranslationResult, error) {
	log.Printf("Translating and summarizing: %s", item.Title)

	// タイトルを翻訳
	translatedTitle, err := ts.translateWithDeepL(item.Title)
	if err != nil {
		log.Printf("Warning: Title translation failed, using original: %v", err)
		translatedTitle = item.Title
	}

	// 説明文を翻訳
	translatedDescription, err := ts.translateWithDeepL(item.Description)
	if err != nil {
		log.Printf("Warning: Description translation failed, using original: %v", err)
		translatedDescription = item.Description
	}

	// OpenAI APIで要約を生成
	summary, err := ts.generateSummaryWithOpenAI(translatedTitle, translatedDescription)
	if err != nil {
		log.Printf("Warning: Summary generation failed: %v", err)
		summary = "要約の生成に失敗しました。"
	}

	result := &TranslationResult{
		OriginalTitle:         item.Title,
		TranslatedTitle:       translatedTitle,
		OriginalDescription:   item.Description,
		TranslatedDescription: translatedDescription,
		Summary:               summary,
		Link:                  item.Link,
	}

	log.Printf("Translation and summarization completed for: %s", item.Title)
	return result, nil
}

// translateWithDeepL はDeepL APIを使用してテキストを翻訳する
func (ts *TranslatorService) translateWithDeepL(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	// リクエストボディを作成
	reqBody := DeepLRequest{
		Text:       []string{text},
		TargetLang: "JA",
		SourceLang: "EN",
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", ts.deepLAPIURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーを設定
	req.Header.Set("Authorization", "DeepL-Auth-Key "+ts.deepLAPIKey)
	req.Header.Set("Content-Type", "application/json")

	// リクエストを送信
	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepL API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// レスポンスをパース
	var deepLResp DeepLResponse
	if err := json.Unmarshal(body, &deepLResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepLResp.Translations) == 0 {
		return "", fmt.Errorf("no translations returned from DeepL")
	}

	return deepLResp.Translations[0].Text, nil
}

// translateWithDeepLFormData はDeepL APIをform-dataで呼び出す（代替実装）
func (ts *TranslatorService) translateWithDeepLFormData(text string) (string, error) {
	if strings.TrimSpace(text) == "" {
		return "", nil
	}

	// フォームデータを作成
	data := url.Values{}
	data.Set("text", text)
	data.Set("target_lang", "JA")
	data.Set("source_lang", "EN")

	// HTTPリクエストを作成
	req, err := http.NewRequest("POST", ts.deepLAPIURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// ヘッダーを設定
	req.Header.Set("Authorization", "DeepL-Auth-Key "+ts.deepLAPIKey)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// リクエストを送信
	resp, err := ts.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// レスポンスを読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepL API error: status=%d, body=%s", resp.StatusCode, string(body))
	}

	// レスポンスをパース
	var deepLResp DeepLResponse
	if err := json.Unmarshal(body, &deepLResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(deepLResp.Translations) == 0 {
		return "", fmt.Errorf("no translations returned from DeepL")
	}

	return deepLResp.Translations[0].Text, nil
}

// generateSummaryWithOpenAI はOpenAI APIを使用して要約を生成する
func (ts *TranslatorService) generateSummaryWithOpenAI(title, description string) (string, error) {
	// プロンプトを作成
	prompt := fmt.Sprintf(`以下の技術記事の内容を、日本語で3行以内で要約してください。重要なポイントと学べる内容を含めて簡潔にまとめてください。

タイトル: %s

内容: %s

要約:`, title, description)

	// OpenAI APIにリクエストを送信
	resp, err := ts.openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: ts.openAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "あなたは技術記事の要約を得意とするAIアシスタントです。与えられた記事の内容を日本語で3行以内で簡潔に要約してください。",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   200,
			Temperature: 0.3,
		},
	)

	if err != nil {
		return "", fmt.Errorf("failed to generate summary with OpenAI: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no summary generated by OpenAI")
	}

	summary := strings.TrimSpace(resp.Choices[0].Message.Content)
	
	// 要約の長さチェック（あまりに長い場合は切り詰める）
	lines := strings.Split(summary, "\n")
	if len(lines) > 3 {
		summary = strings.Join(lines[:3], "\n")
	}

	return summary, nil
}

// TestDeepLConnection はDeepL APIの接続をテストする
func (ts *TranslatorService) TestDeepLConnection() error {
	testText := "Hello, World!"
	_, err := ts.translateWithDeepL(testText)
	if err != nil {
		// JSON形式で失敗した場合はform-data形式を試す
		_, err2 := ts.translateWithDeepLFormData(testText)
		if err2 != nil {
			return fmt.Errorf("DeepL connection test failed (JSON: %v, FormData: %v)", err, err2)
		}
	}
	return nil
}

// TestOpenAIConnection はOpenAI APIの接続をテストする
func (ts *TranslatorService) TestOpenAIConnection() error {
	// 簡単なテストリクエストを送信
	_, err := ts.openAIClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: ts.openAIModel,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello, this is a connection test.",
				},
			},
			MaxTokens: 10,
		},
	)
	return err
}