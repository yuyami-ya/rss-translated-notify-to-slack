# GitHub Actions セットアップガイド

このドキュメントでは、RSS通知システムをGitHub Actionsで実行するための設定方法と手動実行方法を説明します。

## 📋 目次

1. [GitHub Secretsの設定](#github-secretsの設定)
2. [手動実行方法](#手動実行方法)
3. [実行ログの確認](#実行ログの確認)
4. [トラブルシューティング](#トラブルシューティング)
5. [スケジュール設定の変更](#スケジュール設定の変更)

## 🔐 GitHub Secretsの設定

### 1. Secretsページへのアクセス

1. GitHubリポジトリのページにアクセス
2. **Settings** タブをクリック
3. 左サイドバーの **Secrets and variables** をクリック
4. **Actions** をクリック

### 2. 必須Secretsの設定

以下のSecretsを **New repository secret** ボタンで順次設定してください：

#### 🔴 必須項目（これらがないとアプリケーションが起動しません）

| Secret名 | 説明 | 設定例 |
|----------|------|--------|
| `FEED_URLS` | 監視するRSSフィードURL | `https://blog.bytebytego.com/feed` |
| `DEEPL_API_KEY` | DeepL翻訳APIキー | `your-deepl-api-key-here` |
| `OPENAI_API_KEY` | OpenAI APIキー | `sk-proj-...` |
| `SLACK_WEBHOOK_URL` | Slack Webhook URL | `https://hooks.slack.com/services/...` |

#### 🟡 推奨項目（デフォルト値がありますが、設定を推奨）

| Secret名 | 説明 | デフォルト値 | 推奨設定値 |
|----------|------|-------------|------------|
| `DEEPL_API_URL` | DeepL API エンドポイント | `https://api-free.deepl.com/v2/translate` | 無料プラン: `https://api-free.deepl.com/v2/translate`<br>有料プラン: `https://api.deepl.com/v2/translate` |
| `OPENAI_MODEL` | 使用するOpenAIモデル | `gpt-3.5-turbo` | `gpt-3.5-turbo` または `gpt-4` |
| `SLACK_CHANNEL` | 通知先Slackチャンネル | `#general` | `#rss-notifications` など |
| `SLACK_USE_THREADS` | スレッド形式での通知 | `true` | `true` または `false` |

### 3. 複数フィード設定（オプション）

複数のRSSフィードを監視したい場合は、`FEED_URLS`にカンマ区切りで設定：

```
https://blog.bytebytego.com/feed,https://example.com/rss,https://another-blog.com/feed
```

### 4. Secretsの設定手順（詳細）

1. **New repository secret** をクリック
2. **Name** フィールドにSecret名を入力（例：`DEEPL_API_KEY`）
3. **Secret** フィールドに実際の値を入力
4. **Add secret** をクリック
5. 全てのSecretsについて繰り返し

## 🚀 手動実行方法

### 1. GitHub Actionsページにアクセス

1. GitHubリポジトリのページで **Actions** タブをクリック
2. 左サイドバーから **RSS Notification** ワークフローを選択

### 2. ワークフローの手動実行

1. **Run workflow** ボタンをクリック（画面右上）
2. ブランチを選択（通常は `main`）
3. **Run workflow** ボタンをクリックして実行開始

### 3. 実行状況の確認

- 実行中のワークフローが一覧に表示されます
- ステータス：
  - 🟡 実行中（オレンジ色の円）
  - ✅ 成功（緑色のチェックマーク）
  - ❌ 失敗（赤色のX）

## 📊 実行ログの確認

### 1. ワークフロー実行詳細の表示

1. 確認したいワークフロー実行をクリック
2. **rss-check** ジョブをクリック

### 2. ステップごとのログ確認

各ステップをクリックするとログが表示されます：

- **Set up Go**: Go環境のセットアップ
- **Cache Go modules**: Go依存関係のキャッシュ
- **Download dependencies**: 依存パッケージのダウンロード
- **Run RSS notification**: メイン処理の実行

### 3. 主要なログメッセージ

正常実行時のログ例：
```
RSS通知システムを開始します...
設定読み込み完了: フィードURL数=1, 最大記事数/フィード=10
外部サービスの接続をテストしています...
DeepL API接続成功
OpenAI API接続成功
Slack Webhook接続成功
過去24時間以内の新しい記事をチェックしています...
```

## 🔧 トラブルシューティング

### よくあるエラーと対処法

#### 1. 環境変数設定エラー
```
Environment variable DEEPL_API_KEY is required but not set
```
**対処法**: 該当するSecretが正しく設定されているか確認

#### 2. API認証エラー
```
Failed to connect to DeepL API: 403 Forbidden
```
**対処法**: APIキーが正しいか、使用量制限に達していないか確認

#### 3. Slack通知エラー
```
Failed to send Slack notification: invalid_payload
```
**対処法**: Slack Webhook URLが正しいか、チャンネルが存在するか確認

#### 4. タイムアウトエラー
```
Process timeout after 300 seconds
```
**対処法**: 処理する記事数が多すぎる可能性。`MAX_ARTICLES_PER_FEED`の値を下げることを検討

### デバッグ手順

1. **Secretsの確認**
   - 必須Secretsが全て設定されているか
   - Secret名にタイポがないか
   - APIキーが有効期限内か

2. **外部サービスの確認**
   - DeepL APIの使用量制限
   - OpenAI APIの使用量制限
   - Slack Webhookの有効性

3. **ログの詳細確認**
   - エラーメッセージの詳細を確認
   - どのステップで失敗しているかを特定

## ⏰ スケジュール設定の変更

現在の設定：毎日18時JST（UTCで9時）に実行

### 実行時刻の変更

`.github/workflows/rss-notification.yml`の`cron`設定を変更：

```yaml
schedule:
  # 毎日12時JST（UTCで3時）に実行
  - cron: "0 3 * * *"
  
  # 毎日6時と18時JST（UTCで21時と9時）に実行
  - cron: "0 21,9 * * *"
  
  # 平日のみ9時JST（UTCで0時）に実行
  - cron: "0 0 * * 1-5"
```

### cron形式の説明
```
* * * * *
│ │ │ │ │
│ │ │ │ └── 曜日 (0-7, 0と7は日曜日)
│ │ │ └──── 月 (1-12)
│ │ └────── 日 (1-31)
│ └──────── 時 (0-23) ※UTC時間
└────────── 分 (0-59)
```

## 💡 運用のベストプラクティス

### 1. 初回設定時
- 最小限のSecretsで動作確認
- 手動実行でテスト
- ログを確認して正常動作を確認

### 2. 本格運用時
- APIキーの使用量監視
- 定期的なログ確認
- エラー通知の設定検討

### 3. セキュリティ
- APIキーの定期的な更新
- Secretsへの不要なアクセス権限の削除
- 実行ログに機密情報が含まれていないことの確認

## 📞 サポート

問題が解決しない場合は、以下の情報を含めてIssueを作成してください：

- エラーメッセージの全文
- 実行ログの該当箇所
- 設定した環境変数（値は含めない）
- 実行環境の情報

---

**参考資料**:
- [GitHub Actions Documentation](https://docs.github.com/ja/actions)
- [GitHub Secrets管理](https://docs.github.com/ja/actions/security-guides/encrypted-secrets)
- [Cron式ジェネレーター](https://crontab.guru/)
