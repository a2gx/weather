# Переменные среды по умолчанию
ENV_MODE ?= production

run:
	go run cmd/server/main.go

.PHONY: run
