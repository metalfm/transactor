include .env
-include .envrc
export

WORK_MODULES = $(shell go work edit -json | sed -n 's/.*"DiskPath": "\(.*\)".*/\1/p' | grep -v '^./tool$$' | awk '{print $$0 "/..."}' | xargs echo)

work:
	@go work init ./driver/sql/trm ./internal ./tool ./tr ./trtest

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

test:
	@go test -race $(WORK_MODULES)
