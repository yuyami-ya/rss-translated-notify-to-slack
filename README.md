# RSS 英日翻訳通知システム

## 概要

ByteByteGo の RSS フィードを監視し、新記事を自動的に日本語翻訳して Slack で通知するシステム。GitHub Actions による完全無料での自動実行にも対応。

## 必要要件

- Docker
- Docker Compose
- Make

## 環境変数の設定

本アプリケーションでは以下の環境変数を使用します。セキュリティ上の理由から、`.env`ファイルは Git リポジトリで管理していません。
開発・本番環境では、以下の環境変数を`.env`ファイルに設定してください。

| カテゴリ                 | 環境変数                 | 説明                        | デフォルト値                              | 必須 |
| ------------------------ | ------------------------ | --------------------------- | ----------------------------------------- | ---- |
| **RSS フィード設定**     | `FEED_URLS`              | 監視する RSS フィードの URL（複数可、カンマ区切り） | `https://blog.bytebytego.com/feed` | ❌   |
|                          | `MAX_ARTICLES_PER_FEED`  | フィードあたりの最大記事数  | `10`                                      | ❌   |
| **DeepL API 設定**       | `DEEPL_API_KEY`          | DeepL API キー              | -                                         | ✅   |
|                          | `DEEPL_API_URL`          | DeepL API URL               | `https://api-free.deepl.com/v2/translate` | ❌   |
| **OpenAI API 設定**      | `OPENAI_API_KEY`         | OpenAI API キー             | -                                         | ✅   |
|                          | `OPENAI_MODEL`           | OpenAI モデル               | `gpt-3.5-turbo`                           | ❌   |
| **Slack 通知設定**       | `SLACK_WEBHOOK_URL`      | Slack Webhook URL           | -                                         | ✅   |
|                          | `SLACK_CHANNEL`          | Slack チャンネル            | `#general`                                | ❌   |
|                          | `SLACK_USE_THREADS`      | スレッド形式通知の有効化    | `true`                                    | ❌   |
| **アプリケーション設定** | `LOG_LEVEL`              | ログレベル                  | `info`                                    | ❌   |
|                          | `TIMEZONE`               | タイムゾーン                | `Asia/Tokyo`                              | ❌   |

### 環境変数ファイルの作成

プロジェクトルートに`env.example`ファイルが用意されています。これをコピーして`.env`ファイルを作成し、適切な値を設定してください。

```bash
cp env.example .env
```

> **注意**: `.env`ファイルには機密情報が含まれるため、Git リポジトリにコミットしないでください。

## アプリケーションの起動方法

### 1. 環境構築

```bash
# プロジェクトの初期化（ビルド、依存関係のダウンロード）
make init
```

### 2. 基本的なコマンド

```bash
# Dockerコンテナの起動
make up

# Dockerコンテナの停止
make down

# Dockerコンテナの再起動
make restart

# コンテナのログを表示
make logs

# コンテナの状態を確認
make status
```

### 3. アプリケーション実行

```bash
# アプリケーションを実行
make run

# コンテナ内でシェル起動
make shell

# カスタムコマンド実行
make exec cmd="go version"
```

### 4. 開発・メンテナンス

```bash
# Dockerイメージを再ビルド
make build

# コードフォーマット実行
make fmt

# テスト実行
make test

# 静的解析実行
make vet

# 依存関係整理
make mod-tidy

# コンテナ、イメージ、ボリュームを削除
make clean
```

## ドキュメント

詳細なドキュメントは[docs](./docs/)フォルダを参照してください：

- [システム機能詳細](./docs/features.md)
- [API キー取得手順](./docs/api-setup.md)
- [GitHub Actions セットアップガイド](./docs/github-actions-setup.md)
- [通知形式の設定](./docs/notification-formats.md)
- [トラブルシューティング](./docs/troubleshooting.md)
- [運用ベストプラクティス](./docs/best-practices.md)

## ライセンス

このプロジェクトは MIT ライセンスの下で公開されています。

## サポート

問題が発生した場合は、以下の情報とともに Issue を作成してください：

- 実行環境（OS、Docker バージョンなど）
- エラーメッセージ
- 実行したコマンド
- 設定内容（API キーは除く）
