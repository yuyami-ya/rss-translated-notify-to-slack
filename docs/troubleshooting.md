# トラブルシューティング

## よくある問題と解決方法

### 1. API キー関連のエラー

#### DeepL API キーが無効

```
ERROR: DeepL connection test failed
```

**原因と解決方法:**

- `.env` ファイルの API キーが正しく設定されているか確認
- API キーにスペースや改行が含まれていないか確認
- DeepL のアカウント設定で API キーが有効か確認
- 無料プランの場合は月間制限（500,000 文字）を超えていないか確認

**設定例:**

```bash
# 正しい設定
DEEPL_API_KEY=your-actual-api-key-here

# 間違った設定例
DEEPL_API_KEY=" your-actual-api-key-here "  # スペースあり
DEEPL_API_KEY=                              # 空文字
```

#### OpenAI API キーが無効

```
ERROR: OpenAI connection test failed
```

**原因と解決方法:**

- OpenAI Platform で API キーが有効か確認
- 支払い方法が正しく設定されているか確認
- 使用量制限に達していないか確認
- API キーの権限設定を確認

### 2. Slack 通知が届かない

#### Webhook URL エラー

```
ERROR: Slack API error: status=400
```

**原因と解決方法:**

- Webhook URL が正しく設定されているか確認
- URL の形式が `https://hooks.slack.com/services/...` であることを確認
- Slack アプリの Incoming Webhooks が有効になっているか確認

#### チャンネル設定エラー

```
ERROR: channel not found
```

**原因と解決方法:**

- チャンネル名が正しいか確認（#を含める）
- Slack アプリがそのチャンネルに投稿権限を持っているか確認
- チャンネルが存在するか、アーカイブされていないか確認

### 3. RSS フィード関連のエラー

#### フィード読み込みエラー

```
ERROR: failed to parse RSS feed
```

**原因と解決方法:**

- インターネット接続を確認
- フィード URL が正しいか確認
- RSS フィードが一時的に利用不可の可能性
- DNS 設定を確認

#### 新しい記事が検出されない

```
新しい記事はありませんでした
```

**原因と解決方法:**

- RSS フィードに実際に新しい記事があるか確認
- `last_checked_state.txt` ファイルを削除して再初期化
- チェック間隔（`CHECK_INTERVAL_MINUTES`）が適切か確認

### 4. Docker 関連の問題

#### コンテナが起動しない

```
ERROR: container failed to start
```

**原因と解決方法:**

```bash
# ログを確認
make logs

# コンテナを再ビルド
make build

# クリーンアップして再作成
make clean
make init
```

#### 環境変数が読み込まれない

**解決方法:**

```bash
# .envファイルが存在するか確認
ls -la .env

# 正しいフォーマットか確認
cat .env

# コンテナを再起動
make restart
```

### 5. 翻訳・要約の品質問題

#### 翻訳が不自然

**解決方法:**

- DeepL API の無料プランから有料プランへのアップグレード
- 原文の文字数を確認（長すぎる場合は分割処理）

#### 要約が適切でない

**解決方法:**

```bash
# より高性能なモデルを使用
OPENAI_MODEL=gpt-4

# または
OPENAI_MODEL=gpt-4-turbo
```

## ログの確認方法

### 基本的なログ確認

```bash
# リアルタイムでログを表示
make logs

# 特定の時間範囲のログ
docker compose logs --since="2024-01-15T10:00:00" app

# エラーのみフィルタリング
make logs | grep "ERROR"
```

### 詳細なログ出力

```bash
# デバッグモードでの実行
LOG_LEVEL=debug make restart
```

### ログレベルの説明

- `debug`: 詳細な処理情報（開発・デバッグ用）
- `info`: 通常の処理情報（本番推奨）
- `warn`: 警告レベルの情報
- `error`: エラーのみ

## 設定の診断

### 環境変数の確認

```bash
# コンテナ内で環境変数を確認
make exec cmd="env | grep -E '(DEEPL|OPENAI|SLACK)'"
```

### 設定ファイルの検証

```bash
# 設定が正しく読み込まれているか確認
make exec cmd="go run -tags debug main.go --check-config"
```

## パフォーマンスの問題

### 処理が遅い

**原因と対策:**

- **DeepL API**: レート制限に達している可能性
- **OpenAI API**: モデルの変更を検討（gpt-3.5-turbo は高速）
- **ネットワーク**: 接続速度を確認

### メモリ使用量が多い

**対策:**

```bash
# メモリ使用量の確認
docker stats rss-notification-app

# 不要なファイルのクリーンアップ
make exec cmd="rm -f *.log *.tmp"
```

## よくある設定ミス

### 1. 環境変数の形式エラー

```bash
# 間違い
SLACK_USE_THREADS="true"  # 余分なクォート

# 正しい
SLACK_USE_THREADS=true
```

### 2. URL の設定エラー

```bash
# 間違い
DEEPL_API_URL=api-free.deepl.com/v2/translate  # https:// がない

# 正しい
DEEPL_API_URL=https://api-free.deepl.com/v2/translate
```

### 3. チャンネル名の設定エラー

```bash
# 間違い
SLACK_CHANNEL=general  # # がない

# 正しい
SLACK_CHANNEL=#general
```

## 緊急時の対応

### システムを一時停止する

```bash
# アプリケーションの停止
make down
```

### 設定をリセットする

```bash
# 状態ファイルの削除
rm -f last_checked_state.txt

# 環境設定のリセット
cp env.example .env
# 必要な値を再設定
```

### ログの緊急バックアップ

```bash
# ログのエクスポート
make logs > emergency_logs_$(date +%Y%m%d_%H%M%S).txt
```

## サポート情報の収集

問題が解決しない場合、以下の情報を収集してサポートに連絡：

### システム情報

```bash
# Docker バージョン
docker --version
docker compose --version

# OS 情報
uname -a

# 実行環境
make status
```

### エラー情報

```bash
# 最新のエラーログ
make logs --tail=100 | grep -A5 -B5 "ERROR"

# 設定情報（機密情報は除外）
env | grep -E '(FEED_URL|CHECK_INTERVAL|LOG_LEVEL|TIMEZONE)'
```

### 実行コマンド履歴

- 問題発生前に実行したコマンド
- エラーが発生した際の具体的な操作
- 設定変更の内容
