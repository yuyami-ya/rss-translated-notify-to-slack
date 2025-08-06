package main

import (
	"fmt"
	"log"
	"time"

	"rss-en-to-jp-notification/config"
	"rss-en-to-jp-notification/service"
)

// App はアプリケーションのメイン構造体
type App struct {
	config              *config.Config
	feedService         *service.FeedService
	translatorService   *service.TranslatorService
	notificationService *service.NotificationService
}

func main() {
	log.Println("RSS通知システムを開始します...")

	// 設定を読み込み
	cfg := config.LoadConfig()
	log.Printf("設定読み込み完了: フィードURL数=%d, 最大記事数/フィード=%d", len(cfg.FeedURLs), cfg.MaxArticlesPerFeed)

	// アプリケーションを初期化
	app := NewApp(cfg)

	// 各サービスの接続テスト
	if err := app.TestConnections(); err != nil {
		log.Fatalf("接続テストに失敗しました: %v", err)
	}

	// メイン処理を実行（一回だけ）
	app.RunOnce()

	log.Println("RSS通知システムを終了します...")
}

// NewApp は新しいAppインスタンスを作成する
func NewApp(cfg *config.Config) *App {
	// サービスを初期化
	feedService := service.NewFeedService(cfg.FeedURLs, cfg.MaxArticlesPerFeed)
	translatorService := service.NewTranslatorService(
		cfg.DeepLAPIKey,
		cfg.DeepLAPIURL,
		cfg.OpenAIAPIKey,
		cfg.OpenAIModel,
	)
	notificationService := service.NewNotificationService(
		cfg.SlackWebhookURL,
		cfg.SlackChannel,
	)

	return &App{
		config:              cfg,
		feedService:         feedService,
		translatorService:   translatorService,
		notificationService: notificationService,
	}
}

// TestConnections は各外部サービスの接続をテストする
func (app *App) TestConnections() error {
	log.Println("外部サービスの接続をテストしています...")

	// DeepL API接続テスト
	log.Println("DeepL APIの接続をテスト中...")
	if err := app.translatorService.TestDeepLConnection(); err != nil {
		return err
	}
	log.Println("DeepL API接続成功")

	// OpenAI API接続テスト
	log.Println("OpenAI APIの接続をテスト中...")
	if err := app.translatorService.TestOpenAIConnection(); err != nil {
		return err
	}
	log.Println("OpenAI API接続成功")

	// Slack Webhook接続テスト
	log.Println("Slack Webhookの接続をテスト中...")
	if err := app.notificationService.TestSlackConnection(); err != nil {
		return err
	}
	log.Println("Slack Webhook接続成功")

	log.Println("全ての接続テストが完了しました")
	return nil
}

// RunOnce は一度だけRSSチェックと処理を実行する
func (app *App) RunOnce() {
	log.Println("過去24時間以内の新しい記事をチェックしています...")

	// 過去24時間以内の新しい記事をチェック
	recentItems, err := app.feedService.CheckForRecentItems()
	if err != nil {
		errMsg := "RSSフィードのチェックに失敗しました: " + err.Error()
		log.Printf("ERROR: %s", errMsg)
		
		// エラー通知を送信
		if notifyErr := app.notificationService.SendErrorNotification(errMsg); notifyErr != nil {
			log.Printf("WARNING: エラー通知の送信に失敗: %v", notifyErr)
		}
		return
	}

	if len(recentItems) == 0 {
		log.Println("過去24時間以内の新しい記事はありませんでした")
		return
	}

	log.Printf("%d件の新しい記事が見つかりました", len(recentItems))

	// 各記事を処理
	var results []*service.TranslationResult
	for i, item := range recentItems {
		log.Printf("記事 %d/%d を処理中: %s", i+1, len(recentItems), item.Title)

		// 翻訳と要約を実行
		result, err := app.translatorService.TranslateAndSummarize(item)
		if err != nil {
			errMsg := fmt.Sprintf("記事の翻訳・要約に失敗しました: %s - エラー: %v", item.Title, err)
			log.Printf("ERROR: %s", errMsg)
			continue
		}

		results = append(results, result)
		log.Printf("SUCCESS: 記事の処理完了: %s", result.TranslatedTitle)
		
		// API制限を考慮して記事間に間隔を設ける
		if i < len(recentItems)-1 {
			time.Sleep(2 * time.Second)
		}
	}

	// 通知を送信
	if len(results) > 0 {
		app.sendNotifications(results)
	}
}

// sendNotifications は処理結果に基づいて通知を送信する
func (app *App) sendNotifications(results []*service.TranslationResult) {
	log.Printf("%d件の記事通知を送信します", len(results))

	if len(results) == 1 {
		// 単一記事の通知（設定によってスレッド形式or通常形式を選択）
		if app.config.SlackUseThreads {
			if err := app.notificationService.SendNewArticleNotificationWithThread(results[0]); err != nil {
				errMsg := fmt.Sprintf("Slackスレッド通知の送信に失敗しました: %v", err)
				log.Printf("ERROR: %s", errMsg)
				
				// フォールバック: 通常の通知を試行
				log.Println("通常の通知形式にフォールバックします...")
				if err := app.notificationService.SendNewArticleNotification(results[0]); err != nil {
					log.Printf("ERROR: フォールバック通知も失敗しました: %v", err)
				} else {
					log.Println("SUCCESS: フォールバック通知を送信しました")
				}
			} else {
				log.Println("SUCCESS: スレッド形式の記事通知を送信しました")
			}
		} else {
			if err := app.notificationService.SendNewArticleNotification(results[0]); err != nil {
				errMsg := fmt.Sprintf("Slack通知の送信に失敗しました: %v", err)
				log.Printf("ERROR: %s", errMsg)
			} else {
				log.Println("SUCCESS: 記事通知を送信しました")
			}
		}
	} else {
		// 複数記事の通知
		if app.config.SlackUseThreads {
			log.Println("個別スレッド通知を送信します...")
			for i, result := range results {
				if err := app.notificationService.SendNewArticleNotificationWithThread(result); err != nil {
					log.Printf("ERROR: 記事 %d/%d のスレッド通知送信に失敗: %v", i+1, len(results), err)
					// フォールバック: 通常の通知
					if err := app.notificationService.SendNewArticleNotification(result); err != nil {
						log.Printf("ERROR: 記事 %d/%d の通常通知も失敗: %v", i+1, len(results), err)
					} else {
						log.Printf("SUCCESS: 記事 %d/%d の通常通知を送信しました", i+1, len(results))
					}
				} else {
					log.Printf("SUCCESS: 記事 %d/%d のスレッド通知を送信しました", i+1, len(results))
				}
				
				// レート制限を避けるため少し待機
				if i < len(results)-1 {
					time.Sleep(2 * time.Second)
				}
			}
		} else {
			log.Println("個別通知を送信します...")
			for i, result := range results {
				if err := app.notificationService.SendNewArticleNotification(result); err != nil {
					log.Printf("ERROR: 記事 %d/%d の通知送信に失敗: %v", i+1, len(results), err)
				} else {
					log.Printf("SUCCESS: 記事 %d/%d の通知を送信しました", i+1, len(results))
				}
				
				// レート制限を避けるため少し待機
				if i < len(results)-1 {
					time.Sleep(1 * time.Second)
				}
			}
		}
	}
}