include .env
export

export PROJECT_ROOT=${shell pwd}

db-up:
	@docker compose up -d checklist-postgres

db-down:
	@docker compose down checklist-postgres

db-cleanup:
	@read -p "Очистить все volume файлы окружения? Опасность утери данных! [y/N]: " ans; \
		if [ "$$ans" = "y" ]; then \
				docker compose down checklist-postgres port-forwarder && \
				rm -rf ${PROJECT_ROOT}/out/pgdata && \
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
	fi; \
	
	@docker compose run --rm checklist-postgres-migrate \
		create \
		-ext sql \
		-dir /migrations \
		-seq "${seq}"

migrate-action:
	@if [ -z "$(action)" ]; then \
		echo "Отсутствует параметр action. Пример: make migrate-action action=up"; \
		exit 1; \
	fi; \

	@docker compose run --rm checklist-postgres-migrate \
		-path /migrations \
		-database postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@checklist-postgres:5432/${POSTGRES_DB}?sslmode=disable \
		${action}

migrate-up:
	@make migrate-action action=up

migrate-down:
	@make migrate-action action=down

logs-cleanup:
	@read -p "Очистить все log файлы? Опасность утери логов! [y/N]: " ans; \
		if [ "$$ans" = "y" ]; then \
				rm -rf ${PROJECT_ROOT}/out/logs && \
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

up-run:
	@export LOGGER_FOLDER=${PROJECT_ROOT}/out/logs && \
	export POSTGRES_HOST=localhost && \
	go mod tidy && \
	go run ${PROJECT_ROOT}/cmd/app/main.go

ps:
	@docker compose ps