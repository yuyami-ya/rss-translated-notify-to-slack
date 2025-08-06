# API キー取得手順

## DeepL API

### 1. アカウント作成

1. [DeepL Pro API](https://www.deepl.com/pro-api) にアクセス
2. 「無料で始める」をクリック
3. アカウント情報を入力して登録

### 2. API キー取得

1. DeepL アカウントにログイン
2. 「アカウント」→「API キー」に移動
3. API キーをコピー

### 3. 利用制限

- **無料プラン**: 月間 500,000 文字まで
- **有料プラン**: 従量課金制（文字数による）
- **API URL**:
  - 無料プラン: `https://api-free.deepl.com/v2/translate`
  - 有料プラン: `https://api.deepl.com/v2/translate`

### 4. 設定例

```bash
DEEPL_API_KEY=your_actual_deepl_api_key_here
DEEPL_API_URL=https://api-free.deepl.com/v2/translate
```

## OpenAI API

### 1. アカウント作成

1. [OpenAI Platform](https://platform.openai.com/) にアクセス
2. 「Sign up」でアカウント作成
3. 電話番号認証を完了

### 2. API キー作成

1. OpenAI Platform にログイン
2. 左メニューから「API keys」を選択
3. 「Create new secret key」をクリック
4. キー名を入力し、「Create secret key」
5. 表示された API キーをコピー（一度しか表示されません）

### 3. 支払い設定

1. 「Settings」→「Billing」に移動
2. 支払い方法を設定
3. 使用量制限を設定（推奨: 月額 10-50 ドル）

### 4. 利用料金（2024 年 1 月現在）

- **GPT-3.5-turbo**: 入力 $0.0015/1K tokens, 出力 $0.002/1K tokens
- **GPT-4**: 入力 $0.03/1K tokens, 出力 $0.06/1K tokens
- **GPT-4-turbo**: 入力 $0.01/1K tokens, 出力 $0.03/1K tokens

### 5. 設定例

```bash
OPENAI_API_KEY=sk-your_actual_openai_api_key_here
OPENAI_MODEL=gpt-3.5-turbo
```

### 6. モデル選択の指針

- **gpt-3.5-turbo**: コスト重視、十分な要約品質
- **gpt-4**: 品質重視、より詳細で正確な要約
- **gpt-4-turbo**: バランス重視、コストと品質の中間

## Slack Webhook

### 1. Slack アプリの作成

1. [Slack API](https://api.slack.com/apps) にアクセス
2. 「Create New App」をクリック
3. 「From scratch」を選択
4. アプリ名とワークスペースを指定

### 2. Incoming Webhooks の有効化

1. 作成したアプリの設定画面で「Incoming Webhooks」を選択
2. 「Activate Incoming Webhooks」をオンにする
3. 「Add New Webhook to Workspace」をクリック
4. 通知先チャンネルを選択
5. 「Allow」をクリック

### 3. Webhook URL の取得

1. 作成された Webhook URL をコピー
2. URL 形式: `https://hooks.slack.com/services/T.../B.../...`

### 4. 設定例

```bash
SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
SLACK_CHANNEL=#general
SLACK_USE_THREADS=true
```

### 5. チャンネル設定

- **#general**: 全体向け通知
- **#tech-news**: 技術記事専用チャンネル
- **#rss-feeds**: RSS フィード専用チャンネル

### 6. 権限設定

Slack アプリに以下の権限が必要です：

- `incoming-webhook`: メッセージ送信
- `chat:write`: メッセージ投稿
- `chat:write.public`: パブリックチャンネルへの投稿

## セキュリティベストプラクティス

### 1. API キーの管理

- **環境変数**: `.env` ファイルで管理
- **Git 除外**: `.gitignore` に `.env` を追加
- **定期ローテーション**: 3-6 ヶ月ごとにキーを更新
- **最小権限**: 必要最小限の権限のみ付与

### 2. 使用量の監視

- **DeepL**: 月間文字数の監視
- **OpenAI**: 月間使用量とコストの監視
- **アラート設定**: 使用量上限に近づいた際の通知

### 3. エラー処理

- **API 制限**: レート制限への適切な対応
- **フォールバック**: API 障害時の代替処理
- **ログ記録**: API エラーの詳細なログ保存

## コスト見積もり

### 月間 30 記事（1 日 1 記事）の場合

#### DeepL 翻訳

- 1 記事あたり約 1,500 文字
- 月間: 30 × 1,500 = 45,000 文字
- コスト: 無料プラン内（500,000 文字まで）

#### OpenAI 要約

- 1 記事あたり約 400 トークン
- 月間: 30 × 400 = 12,000 トークン
- GPT-3.5-turbo の場合: 約 $0.02/月
- GPT-4 の場合: 約 $0.40/月

#### Slack

- 無料（Webhook 使用）

#### 合計月額コスト

- GPT-3.5-turbo 使用: 約 $0.02/月
- GPT-4 使用: 約 $0.40/月
