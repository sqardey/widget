# Автоматически подключаем .env, если он существует
ifneq (,$(wildcard .env))
    include .env
    export
endif

APP_NAME := widget-server
PORT ?= 8080

.PHONY: run build docker-build docker-run docker-stop docker-logs clean

## Локальный запуск на машине
run:
	go run ./cmd/server/main.go

## Компиляция локального бинарника
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/$(APP_NAME) ./cmd/server/main.go

## Сборка Docker-образа
docker-build:
	docker build -t $(APP_NAME) .

## Запуск контейнера в фоне с пробросом .env
docker-run:
	docker run -d \
		--name $(APP_NAME) \
		--restart unless-stopped \
		-p $(PORT):$(PORT) \
		--env-file .env \
		$(APP_NAME)

## Остановка и удаление контейнера
docker-stop:
	docker stop $(APP_NAME) || true
	docker rm $(APP_NAME) || true

## Просмотр логов контейнера (в реальном времени)
docker-logs:
	docker logs -f $(APP_NAME)

## Перезапуск контейнера со свежей сборкой
docker-restart: docker-stop docker-build docker-run

## Очистка локального бинарника
clean:
	rm -rf bin/