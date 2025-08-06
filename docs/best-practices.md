# 運用ベストプラクティス

## セキュリティ

### API キー管理

- **定期ローテーション**: 3-6 ヶ月ごとに API キーを更新
- **最小権限**: 必要最小限の権限のみ付与
- **環境分離**: 開発・ステージング・本番で異なるキーを使用
- **バックアップ**: キーの更新前に動作確認を実施

### アクセス制御

- **Git 管理**: `.env` ファイルを `.gitignore` に追加
- **コンテナセキュリティ**: 定期的なベースイメージの更新
- **ログ管理**: 機密情報がログに出力されないよう注意

## モニタリング

### システム監視

```bash
# 定期的なヘルスチェック
*/15 * * * * docker ps | grep rss-notification-app || echo "Container down" | mail admin@company.com

# ログサイズの監視
*/30 * * * * find /var/lib/docker -name "*.log" -size +100M -delete
```

### 使用量監視

- **DeepL API**: 月間文字数の追跡
- **OpenAI API**: 月間トークン使用量とコスト
- **Slack API**: レート制限の監視

### アラート設定

```bash
# 環境変数でアラート閾値を設定
DEEPL_MONTHLY_LIMIT=450000      # 月間制限の90%
OPENAI_MONTHLY_COST_LIMIT=10    # 月額10ドル制限
```

## コスト最適化

### API 使用量の最適化

#### DeepL 翻訳

- **キャッシュ**: 同じテキストの翻訳結果をキャッシュ
- **前処理**: 不要な HTML タグを事前に除去
- **文字数制限**: 長すぎるテキストは要約してから翻訳

#### OpenAI 要約

- **モデル選択**: 用途に応じた適切なモデル選択
  - 日常運用: `gpt-3.5-turbo`
  - 高品質要約: `gpt-4-turbo`
- **プロンプト最適化**: 効率的なプロンプト設計
- **トークン制限**: `MaxTokens` パラメータでコスト制御

### 実行頻度の最適化

```bash
# 記事投稿頻度に合わせた調整
CHECK_INTERVAL_MINUTES=60  # 通常は1時間間隔
CHECK_INTERVAL_MINUTES=30  # 活発な期間は30分間隔
```

## 運用効率化

### 自動化設定

#### Systemd サービス（Linux）

```bash
# /etc/systemd/system/rss-notification.service
[Unit]
Description=RSS Notification Service
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/path/to/rss-en-to-jp-notification
ExecStart=/usr/bin/make up
ExecStop=/usr/bin/make down
User=deploy

[Install]
WantedBy=multi-user.target
```

#### Cron での定期チェック

```bash
# アプリケーションのヘルスチェック
*/10 * * * * /path/to/health-check.sh

# 週次のログローテーション
0 2 * * 0 /usr/bin/make logs > /var/log/rss-notification-$(date +\%Y\%m\%d).log
```

### バックアップ戦略

#### 設定のバックアップ

```bash
# 設定ファイルの定期バックアップ
#!/bin/bash
DATE=$(date +%Y%m%d)
cp .env .env.backup.$DATE
cp docker-compose.yml docker-compose.yml.backup.$DATE
```

#### 状態ファイルのバックアップ

```bash
# 状態ファイルの定期バックアップ
0 0 * * * cp /path/to/last_checked_state.txt /backup/last_checked_state.txt.$(date +\%Y\%m\%d)
```

## パフォーマンス最適化

### リソース配分

#### Docker リソース制限

```yaml
# docker-compose.yml
services:
  app:
    deploy:
      resources:
        limits:
          memory: 512M
          cpus: "0.5"
        reservations:
          memory: 256M
          cpus: "0.25"
```

#### 並行処理の最適化

- **記事処理**: 同時に処理する記事数を制限
- **API コール**: レート制限を考慮した並行処理
- **エラー処理**: 適切な再試行間隔の設定

### キャッシュ戦略

```go
// 翻訳結果のキャッシュ例
type TranslationCache struct {
    cache map[string]string
    mutex sync.RWMutex
}

func (tc *TranslationCache) Get(text string) (string, bool) {
    tc.mutex.RLock()
    defer tc.mutex.RUnlock()
    result, exists := tc.cache[text]
    return result, exists
}
```

## 品質保証

### テスト戦略

#### 統合テスト

```bash
# API接続テスト
make exec cmd="go test ./service/... -tags=integration"

# エンドツーエンドテスト
make exec cmd="go test ./... -tags=e2e"
```

#### 手動テスト手順

1. **設定変更後の動作確認**
2. **新機能の動作テスト**
3. **エラーケースの確認**
4. **パフォーマンステスト**

### コード品質

#### 静的解析

```bash
# 定期実行
make vet
make fmt

# セキュリティチェック
go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
gosec ./...
```

#### 依存関係の管理

```bash
# 脆弱性チェック
go install golang.org/x/vuln/cmd/govulncheck@latest
govulncheck ./...

# 依存関係の更新
go get -u ./...
make mod-tidy
```

## 災害復旧

### 障害対応手順

#### レベル 1: 軽微な障害

- ログ確認とエラー内容の把握
- 設定の見直し
- サービスの再起動

#### レベル 2: 重大な障害

- 緊急停止の実施
- バックアップからの復旧
- 根本原因の調査

#### レベル 3: 完全復旧

- システム全体の再構築
- データの整合性確認
- 再発防止策の実装

### 復旧手順書

```bash
# 1. 緊急停止
make down

# 2. バックアップからの復旧
cp .env.backup.YYYYMMDD .env
cp docker-compose.yml.backup.YYYYMMDD docker-compose.yml

# 3. システム再構築
make clean
make init

# 4. 動作確認
make test
```

## チーム連携

### ドキュメント管理

- **変更履歴**: 設定変更の記録
- **運用手順**: 標準作業手順書の整備
- **緊急連絡先**: 障害時の連絡体制

### ナレッジ共有

- **定期レビュー**: 月次の運用レビュー会議
- **改善提案**: チームメンバーからのフィードバック収集
- **ベストプラクティス**: 運用知識の文書化

## 継続的改善

### メトリクス収集

- **処理時間**: 各段階の処理時間測定
- **成功率**: API コールの成功率
- **エラー率**: カテゴリ別エラー分析

### 改善サイクル

1. **メトリクス分析**: 週次のパフォーマンス分析
2. **問題特定**: ボトルネックの特定
3. **改善実装**: 段階的な改善の実装
4. **効果測定**: 改善効果の定量的評価

### バージョン管理

- **セマンティックバージョニング**: 明確なバージョン管理
- **変更ログ**: 各バージョンの変更内容記録
- **ロールバック計画**: 問題発生時の巻き戻し手順
