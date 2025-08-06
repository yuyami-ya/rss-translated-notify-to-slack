package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	
	// アプリケーション状態
	running bool
	mutex   sync.RWMutex
}

func main() {
	log.Println("RSS通知システムを開始します...")

	// 設定を読み込み
	cfg := config.LoadConfig()
	log.Printf("設定読み込み完了: フィードURL=%s, チェック間隔=%v", cfg.FeedURL, cfg.CheckInterval)

	// アプリケーションを初期化
	app := NewApp(cfg)

	// 各サービスの接続テスト
	if err := app.TestConnections(); err != nil {
		log.Fatalf("接続テストに失敗しました: %v", err)
	}

	// 起動通知を送信
	if err := app.notificationService.SendStartupNotification(); err != nil {
		log.Printf("起動通知の送信に失敗しました: %v", err)
	}

	// メインループを開始
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// シグナルハンドリング
	go app.handleSignals(cancel)

	// メインループを実行
	app.Run(ctx)

	log.Println("RSS通知システムを終了します...")
}

// NewApp は新しいAppインスタンスを作成する
func NewApp(cfg *config.Config) *App {
	// サービスを初期化
	feedService := service.NewFeedService(cfg.FeedURL)
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
		running:             true,
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

// Run はメインのアプリケーションループを実行する
func (app *App) Run(ctx context.Context) {
	log.Printf("RSS監視を開始します (チェック間隔: %v)", app.config.CheckInterval)

	// 起動時に一度チェックを実行
	app.checkAndProcess()

	// 定期的なチェックを開始
	ticker := time.NewTicker(app.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("コンテキストがキャンセルされました。終了します...")
			return
		case <-ticker.C:
			if app.isRunning() {
				app.checkAndProcess()
			}
		}
	}
}

// checkAndProcess はRSSフィードをチェックし、新しい記事を処理する
func (app *App) checkAndProcess() {
	log.Println("RSSフィードをチェックしています...")

	// 新しい記事をチェック
	newItems, err := app.feedService.CheckForNewItems()
	if err != nil {
		errMsg := "RSSフィードのチェックに失敗しました: " + err.Error()
		log.Printf("ERROR: %s", errMsg)
		
		// エラー通知を送信
		if notifyErr := app.notificationService.SendErrorNotification(errMsg); notifyErr != nil {
			log.Printf("WARNING: エラー通知の送信に失敗: %v", notifyErr)
		}
		return
	}

	if len(newItems) == 0 {
		log.Println(" 新しい記事はありませんでした")
		return
	}

	log.Printf(" %d件の新しい記事が見つかりました", len(newItems))

	// 各記事を処理
	var results []*service.TranslationResult
	for i, item := range newItems {
		log.Printf(" 記事 %d/%d を処理中: %s", i+1, len(newItems), item.Title)

		// 翻訳と要約を実行
		result, err := app.translatorService.TranslateAndSummarize(item)
		if err != nil {
			errMsg := fmt.Sprintf("記事の翻訳・要約に失敗しました: %s - エラー: %v", item.Title, err)
			log.Printf("ERROR: %s", errMsg)
			
			// エラー通知を送信
			if notifyErr := app.notificationService.SendErrorNotification(errMsg); notifyErr != nil {
				log.Printf("WARNING: エラー通知の送信に失敗: %v", notifyErr)
			}
			continue
		}

		results = append(results, result)
		log.Printf("SUCCESS: 記事の処理完了: %s", result.TranslatedTitle)
	}

	// 通知を送信
	if len(results) > 0 {
		app.sendNotifications(results)
	}
}

// sendNotifications は処理結果に基づいて通知を送信する
func (app *App) sendNotifications(results []*service.TranslationResult) {
	log.Printf(" %d件の記事通知を送信します", len(results))

	if len(results) == 1 {
		// 単一記事の通知（設定によってスレッド形式or通常形式を選択）
		if app.config.SlackUseThreads {
			if err := app.notificationService.SendNewArticleNotificationWithThread(results[0]); err != nil {
				errMsg := fmt.Sprintf("Slackスレッド通知の送信に失敗しました: %v", err)
				log.Printf("ERROR: %s", errMsg)
				
				// フォールバック: 通常の通知を試行
				log.Println(" 通常の通知形式にフォールバックします...")
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
		// 複数記事のバッチ通知
		if err := app.notificationService.SendBatchNotification(results); err != nil {
			errMsg := fmt.Sprintf("バッチ通知の送信に失敗しました: %v", err)
			log.Printf("ERROR: %s", errMsg)
			
			// 個別通知にフォールバック（設定によってスレッド形式or通常形式）
			if app.config.SlackUseThreads {
				log.Println(" 個別スレッド通知にフォールバックします...")
				for i, result := range results {
					if err := app.notificationService.SendNewArticleNotificationWithThread(result); err != nil {
						log.Printf("ERROR: 記事 %d/%d のスレッド通知送信に失敗: %v", i+1, len(results), err)
						// さらにフォールバック: 通常の通知
						if err := app.notificationService.SendNewArticleNotification(result); err != nil {
							log.Printf("ERROR: 記事 %d/%d の通常通知も失敗: %v", i+1, len(results), err)
						} else {
							log.Printf("SUCCESS: 記事 %d/%d の通常通知を送信しました", i+1, len(results))
						}
					} else {
						log.Printf("SUCCESS: 記事 %d/%d のスレッド通知を送信しました", i+1, len(results))
					}
					
					// レート制限を避けるため少し待機
					time.Sleep(2 * time.Second)
				}
			} else {
				log.Println(" 個別通知にフォールバックします...")
				for i, result := range results {
					if err := app.notificationService.SendNewArticleNotification(result); err != nil {
						log.Printf("ERROR: 記事 %d/%d の通知送信に失敗: %v", i+1, len(results), err)
					} else {
						log.Printf("SUCCESS: 記事 %d/%d の通知を送信しました", i+1, len(results))
					}
					
					// レート制限を避けるため少し待機
					time.Sleep(1 * time.Second)
				}
			}
		} else {
			log.Println("SUCCESS: バッチ通知を送信しました")
		}
	}
}

// handleSignals はOSシグナルを処理する
func (app *App) handleSignals(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf(" シグナル %v を受信しました。グレースフルシャットダウンを開始します...", sig)
	
	app.setRunning(false)
	cancel()
}

// isRunning はアプリケーションが実行中かどうかを返す
func (app *App) isRunning() bool {
	app.mutex.RLock()
	defer app.mutex.RUnlock()
	return app.running
}

// setRunning はアプリケーションの実行状態を設定する
func (app *App) setRunning(running bool) {
	app.mutex.Lock()
	defer app.mutex.Unlock()
	app.running = running
}