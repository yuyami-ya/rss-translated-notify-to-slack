.PHONY: help up down restart build logs clean status init run test fmt vet mod-tidy exec shell

# ---------------------------------------------
# ヘルプ
# ---------------------------------------------
help:
	@echo "使用可能なコマンド:"
	@echo ""
	@echo "Docker操作"
	@echo "  make up               Dockerコンテナを起動"
	@echo "  make down             Dockerコンテナを停止"
	@echo "  make restart          Dockerコンテナを再起動"
	@echo "  make build            Dockerイメージを再ビルド"
	@echo "  make logs             コンテナのログを表示"
	@echo "  make status           コンテナの状態を表示"
	@echo "  make shell            コンテナ内でbashを起動"
	@echo "  make exec             コンテナ内でコマンドを実行 (例: make exec cmd=\"go version\")"
	@echo ""
	@echo "Go開発"
	@echo "  make init             プロジェクトの初期化（ビルド、依存関係のダウンロード）"
	@echo "  make run              アプリケーションを実行"
	@echo "  make test             テストを実行"
	@echo "  make fmt              コードフォーマットを実行"
	@echo "  make vet              静的解析を実行"
	@echo "  make mod-tidy         go mod tidyを実行"
	@echo ""
	@echo "本番環境"
	@echo "  make prod-up          本番用コンテナを起動"
	@echo "  make prod-down        本番用コンテナを停止"
	@echo ""
	@echo "その他"
	@echo "  make clean            コンテナ、イメージ、ボリュームを削除"

# ---------------------------------------------
# Docker操作
# ---------------------------------------------
up:
	docker compose up -d app

down:
	docker compose down

restart: down up

build:
	docker compose build --no-cache app

logs:
	docker compose logs -f app

status:
	docker compose ps

shell:
	docker compose exec app sh

exec:
ifndef cmd
	@echo "使用法: make exec cmd=\"<command>\""
	@echo "例: make exec cmd=\"go version\""
	@exit 1
endif
	docker compose exec app $(cmd)

# ---------------------------------------------
# Go開発
# ---------------------------------------------
init: build up mod-tidy
	@echo "プロジェクトの初期化が完了しました"

run:
	docker compose exec app go run main.go

test:
	docker compose exec app go test ./...

fmt:
	docker compose exec app go fmt ./...

vet:
	docker compose exec app go vet ./...

mod-tidy:
	docker compose exec app go mod tidy

mod-download:
	docker compose exec app go mod download

# go.modファイルを初期化（初回のみ使用）
mod-init:
	docker compose exec app go mod init rss-en-to-jp-notification

# ---------------------------------------------
# 本番環境
# ---------------------------------------------
prod-up:
	docker compose --profile production up -d app-prod

prod-down:
	docker compose --profile production down

prod-logs:
	docker compose --profile production logs -f app-prod

# ---------------------------------------------
# その他
# ---------------------------------------------
clean:
	docker compose down -v --rmi all
	docker system prune -f