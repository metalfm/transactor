module github.com/metalfm/transactor/internal

go 1.24.1

require (
	github.com/Thiht/transactor v1.1.0
	github.com/aneshas/tx/v2 v2.3.0
	github.com/avito-tech/go-transaction-manager/drivers/sql/v2 v2.0.0
	github.com/avito-tech/go-transaction-manager/trm/v2 v2.0.0-rc9.2
	github.com/lib/pq v1.10.9
	github.com/metalfm/transactor/driver/sql/trm v1.0.0
	github.com/metalfm/transactor/tr v1.0.0
	github.com/metalfm/transactor/trtest v1.0.0
	github.com/stretchr/testify v1.10.0
	go.uber.org/mock v0.5.1
)

require (
	github.com/aclements/go-moremath v0.0.0-20210112150236-f10218a38794 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	go.uber.org/multierr v1.10.0 // indirect
	golang.org/x/perf v0.0.0-20250414141303-3fc2b901edf3 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

tool golang.org/x/perf/cmd/benchstat
