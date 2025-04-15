include .env

work:
	@go work init ./driver/sql/trm ./internal ./tr ./trtest

up:
	@docker compose up -d --remove-orphans

down:
	@docker compose rm -fsv

bench:
	@cd internal/benchmark &&\
	go test -bench=BenchmarkSQLPostgres -benchmem -count=20 > sql &&\
	go tool benchstat -col ".name /tx" -row ".name" sql
