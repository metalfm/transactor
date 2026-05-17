include .env
-include .envrc
export

WORK_MODULES = ./... ./internal/benchmark/... ./internal/example/...
COVER_PACKAGES = ./tr/... ./driver/...

up:
	@docker compose up -d --remove-orphans

down:
	@docker compose rm -fsv

bench:
	@cd internal/benchmark &&\
	go test -bench=BenchmarkSQLPostgres -benchmem -count=20 > sql &&\
	go tool benchstat -col ".name /tx" -row ".name" sql

lint:
	@go tool golangci-lint run $(WORK_MODULES)

gen:
	@go generate $(WORK_MODULES)

test:
	@go test -race $(WORK_MODULES)

coverage:
	@go test -race -covermode=atomic -coverprofile=coverage.out $(COVER_PACKAGES)
