-include .env
export

PROJECT_ROOT := $(CURDIR)
export PROJECT_ROOT

GO ?= go
GO_TEST_PACKAGES = $(GO) list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...

.PHONY: db-up db-down db-cleanup port-up port-down migrate-create \
	migrate-action migrate-up migrate-down logs-cleanup swagger-gen \
	app-run build test test-race test-cover vet format format-check ci ps

db-up:
	@docker compose up -d --wait checklist-postgres

db-down:
	@docker compose down checklist-postgres

db-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных! [y/N]: " ans; \
		if [ "$$ans" = "y" ]; then \
			target="$(PROJECT_ROOT)/out/pgdata"; \
			if [ -z "$(PROJECT_ROOT)" ] || [ "$(PROJECT_ROOT)" = "/" ]; then \
				echo "Некорректный PROJECT_ROOT, очистка отменена"; \
				exit 1; \
			fi; \
			docker compose down checklist-postgres port-forwarder && \
			rm -rf -- "$$target" && \
			echo "Файлы окружения очищены"; \
		else \
			echo "Очистка окружения отменена"; \
		fi

port-up:
	@docker compose up -d port-forwarder

port-down:
	@docker compose down port-forwarder

migrate-create:
	@if [ -z "$(seq)" ]; then \
		echo "Отсутствует параметр seq. Пример: make migrate-create seq=init"; \
		exit 1; \
	fi

	@docker compose run --rm checklist-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "$(seq)"

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi

	@docker compose run --rm checklist-postgres-migrate \
		-path /migrations \
		-database postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@checklist-postgres:5432/$(POSTGRES_DB)?sslmode=disable \
		$(action)

migrate-up: db-up
	@$(MAKE) migrate-action action=up

migrate-down:
	@$(MAKE) migrate-action action=down

logs-cleanup:
	@read -p "Очистить все log файлы? Опасность утери логов! [y/N]: " ans; \
		if [ "$$ans" = "y" ]; then \
			target="$(PROJECT_ROOT)/out/logs"; \
			if [ -z "$(PROJECT_ROOT)" ] || [ "$(PROJECT_ROOT)" = "/" ]; then \
				echo "Некорректный PROJECT_ROOT, очистка отменена"; \
				exit 1; \
			fi; \
			rm -rf -- "$$target" && \
			echo "Файлы логов очищены"; \
		else \
			echo "Очистка логов отменена"; \
		fi

swagger-gen:
	@docker compose run --rm swagger \
		init \
		-g cmd/app/main.go \
		-o docs \
		--parseInternal \
		--parseDependency

app-run:
	@LOGGER_FOLDER="$(PROJECT_ROOT)/out/logs" POSTGRES_HOST=localhost \
		$(GO) run ./cmd/app

build:
	@mkdir -p "$(PROJECT_ROOT)/out/bin"
	@$(GO) build -o "$(PROJECT_ROOT)/out/bin/checklist-service" ./cmd/app

test:
	@$(GO) test ./...

test-race:
	@$(GO) test -race ./...

test-cover:
	@$(GO) test ./... -coverprofile=out/coverage.out
	@$(GO) tool cover -func=out/coverage.out

vet:
	@$(GO) vet ./...

format:
	@gofmt -w cmd internal

format-check:
	@unformatted="$$(gofmt -l cmd internal)"; \
		if [ -n "$$unformatted" ]; then \
			echo "Найдены неотформатированные Go-файлы:"; \
			echo "$$unformatted"; \
			exit 1; \
		fi

ci: format-check test test-race vet build

ps:
	@docker compose ps
