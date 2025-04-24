# Transactor

`Transactor` is a library for simplifying transaction management in Go.

It provides the `Transactor[T any]` interface,
which allows performing operations within a transaction while abstracting the transaction management logic.

## Installation

Add the library to your project along with the implementation for a specific database driver.

```bash
go get github.com/metalfm/transactor
go get github.com/metalfm/transactor/driver/sql/trm
```

Currently, the `transactor` library supports working with the `sql.DB` driver from Go's standard library. However,
nothing prevents adding more drivers in the future, such as `sqlx`, `pgx`, etc.

## Key Concepts

### 1. `Transactor[T any]` Interface

The interface is simple, and `[T any]` means it can accept any type, allowing it to work with various repository
implementations while maintaining type safety at compile time.

```go
type Transactor[T any] interface {
InTx(ctx context.Context, fn func (T) error) error
}
```

The `InTx` method takes a context and a function. This function contains the logic that should be executed within the
transaction.
`T` is the type of repository that will be used in the business logic.

- If the function returns an error, the transaction is rolled back.
- If the function completes successfully, the transaction is committed.

Example usage:

```go
package example

type repoTx interface {
	CreateUser(ctx context.Context, name string) error
	CreateOrder(ctx context.Context, items []string) error
}
err := transactor.InTx(ctx, func (repo repoTx) error {
	err := repo.CreateUser(ctx, "John Doe")
	if err != nil {
		return err
	}

	err = repo.CreateOrder(ctx, []string{"item1", "item2"})
	if err != nil {
		return err
	}

	return nil
})
```

Note that all dependencies are based on interfaces, making it easy to mock them in tests as well as specific
implementations.

### 2. Repositories and Factory Method

Repositories depend on the `trm.Query` interface, which provides methods for executing SQL queries. This interface is
part of the specific database driver implementation.
The `trm.Transaction` interface, which extends `trm.Query`, is used for transaction management and adds `Commit` and
`Rollback` methods.

#### Definition of `trm.Query` and `trm.Transaction` Interfaces

```go
package trm

import (
	"context"
	"database/sql"
)

// Query — interface for executing SQL queries.
type Query interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Transaction — interface for transaction management.
// Extends Query and adds Commit and Rollback methods.
type Transaction interface {
	Query
	Commit() error
	Rollback() error
}
```

#### Factory Method `WithTx`

The factory method `WithTx` is used for transaction management and returns a new instance of the repository associated
with the `trm.Transaction`. This isolates transaction logic within repositories.

Example implementation of the factory method:

```go
package example

import (
	"github.com/metalfm/transactor/driver/sql/trm"
)

type RepoUser struct {
	q trm.Query
}

func NewRepoUser(q trm.Query) *RepoUser {
	return &RepoUser{q}
}

// WithTx example of a factory method
// all methods of *RepoUser will be called within the transaction
func (slf *RepoUser) WithTx(tx trm.Transaction) *RepoUser {
	return &RepoUser{q: tx}
}
```

Using the factory method allows explicit transaction passing, making the code more readable and safer. Note that the
factory method `WithTx` returns a new instance of `*RepoUser`, and duck typing avoids importing interfaces into business
logic.

### 3. Adapter for Repositories

The adapter is not part of the `transactor` library but provides the ability to combine code from various repositories
using an adapter. The adapter encapsulates the logic of working with multiple repositories, providing a unified
interface for working with them, including performing operations within a single transaction.

```go
package example

import (
	"github.com/metalfm/transactor/driver/sql/trm"
)

type Adapter struct {
	repoUser  *RepoUser
	repoOrder *RepoOrder
}

func NewAdapter(repoUser *svc.RepoUser, repoOrder *svc.RepoOrder) *Adapter {
	return &Adapter{
		repoUser:  repoUser,
		repoOrder: repoOrder,
	}
}

// WithTx example of a factory method for combining logic from multiple repositories
func (slf *Adapter) WithTx(tx trm.Transaction) *Adapter {
	return &Adapter{
		repoUser:  slf.repoUser.WithTx(tx),
		repoOrder: slf.repoOrder.WithTx(tx),
	}
}
```

### 4. Why is the Factory Method Better Than Passing Transactions Through Context?

- **Explicitness**: Transactions are passed explicitly through the factory method, not hidden in the context, making the
  code more readable and understandable.
- **Safety**: Context is intended for passing request-related data (e.g., timeouts or metadata), not for managing
  transaction state.
- **Encapsulation**: The factory method isolates transaction logic within repositories, preventing it from spreading to
  other parts of the code.
- **Testability**: The factory method simplifies creating mocks for testing since the transaction remains part of the
  repository interface.
- **Performance**: Passing transactions through the factory method does not require additional operations, such as
  extracting data from the context or type casting. This makes transaction management faster and more efficient compared
  to using context.

### 5. Example Service

The `Service` contains business logic and depends only on the `Transactor` and `repoTx` interfaces. It knows nothing
about the internal structure of transactions, simplifying testing and isolating logic.

Example:

```go
package app

import (
	"context"
	"fmt"
	"github.com/metalfm/transactor/tr"
)

// repoTx declares dependencies for business logic
// all repository methods that will be used within the transaction
type repoTx interface {
	CreateUser(ctx context.Context, name string) error
	CreateOrder(ctx context.Context, items []string) error
}

type Service[T repoTx] struct {
	tr tr.Transactor[T]
}

func NewService[T repoTx](tr tr.Transactor[T]) *Service[T] {
	return &Service[T]{tr}
}

func (slf *Service[T]) Create(ctx context.Context, name string, items []string) error {
	err := slf.tr.InTx(ctx, func(r T) error {
		err := r.CreateUser(ctx, name)
		if err != nil {
			return fmt.Errorf("create user: %w", err)
		}

		err = r.CreateOrder(ctx, items)
		if err != nil {
			return fmt.Errorf("create order: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("create user & order: %w", err)
	}

	return nil
}
```

You can find the full example here — [example](https://github.com/metalfm/transactor/tree/master/internal/example).

### 6. Testing and `trtest` Package

To simplify testing, the library provides the `trtest` package, which allows creating mock implementations of the
`Transactor[T any]` interface. This is useful for isolating business logic from the real database.

Example usage of
`trtest.MockTransactor` — [example](https://github.com/metalfm/transactor/blob/master/internal/example/app/service_test.go)

## Benchmarks

All benchmarks were conducted using the following setup:

- **Machine**: Apple M1 Pro (Darwin, arm64)
- **Database**: PostgreSQL running in Docker

To reproduce the benchmarks, ensure you have Docker installed and run the following commands:
```bash
make up && make bench
```

### Libraries Used in Comparison

The following libraries and approaches were used for benchmarking:

1. **Native** — a basic approach using the standard `sql.DB` driver from Go's standard library without additional
   abstractions.
2. **⚡ Transactor** — the tested `Transactor` library, which provides the `Transactor[T any]` interface for transaction
   management.
3. **[Avito](https://github.com/avito-tech/go-transaction-manager)** — an approach based on the transaction manager
   implementation used in Avito projects.
4. **[Aneshas](https://github.com/aneshas/tx)** — an alternative library for transaction management.
5. **[Thiht](https://github.com/Thiht/transactor)** — another library for transaction management.

Each approach was tested on identical scenarios to ensure an objective comparison of performance, memory consumption,
and allocation count.

### Benchmark Results

#### Execution Time (sec/op)

| Metric     | Native      | ⚡ Transactor               | Avito                           | Aneshas                    | Thiht                      |
|------------|-------------|----------------------------|---------------------------------|----------------------------|----------------------------|
| **sec/op** | 261.9µ ± 1% | 264.1µ ± 5%    ~ (p=0.398) | 269.4µ ± 4%    +2.85% (p=0.002) | 263.3µ ± 1%    ~ (p=0.718) | 263.3µ ± 1%    ~ (p=0.201) |

#### Memory Consumption (B/op)

| Metric   | Native     | ⚡ Transactor         | Avito                  | Aneshas              | Thiht                 |
|----------|------------|----------------------|------------------------|----------------------|-----------------------|
| **B/op** | 833.0 ± 2% | 914.0 ± 2%    +9.72% | 1454.5 ± 2%    +74.61% | 912.0 ± 2%    +9.48% | 940.5 ± 1%    +12.91% |

#### Allocation Count (allocs/op)

| Metric        | Native     | ⚡ Transactor          | Avito                 | Aneshas               | Thiht                 |
|---------------|------------|-----------------------|-----------------------|-----------------------|-----------------------|
| **allocs/op** | 18.00 ± 0% | 21.00 ± 0%    +16.67% | 33.00 ± 0%    +83.33% | 21.00 ± 5%    +16.67% | 22.00 ± 0%    +22.22% |

### Benchmark Analysis

#### Execution Time (`sec/op`):

- **native**: 261.9µs ± 1% — baseline performance.
- **⚡ transactor**: 264.1µs ± 5% — a slight increase, statistically insignificant (p=0.398).
- **avito**: 269.4µs ± 4% — an increase of **2.85%**, statistically significant (p=0.002).
- **aneshas**: 263.3µs ± 1% — close to native, statistically insignificant (p=0.718).
- **Thiht**: 263.3µs ± 1% — close to native, statistically insignificant (p=0.201).

#### Memory Consumption (`B/op`):

- **native**: 833.0 B ± 2% — baseline memory usage.
- **⚡ transactor**: 914.0 B ± 2% — an **9.72%** increase.
- **avito**: 1454.5 B ± 2% — a **74.61%** increase.
- **aneshas**: 912.0 B ± 2% — an **9.48%** increase.
- **Thiht**: 940.5 B ± 1% — a **12.91%** increase.

#### Allocation Count (`allocs/op`):

- **native**: 18.00 ± 0% — baseline allocation count.
- **⚡ transactor**: 21.00 ± 0% — an **16.67%** increase.
- **avito**: 33.00 ± 0% — an **83.33%** increase.
- **aneshas**: 21.00 ± 5% — an **16.67%** increase.
- **Thiht**: 22.00 ± 0% — an **22.22%** increase.

### Overall Conclusion

- **native** remains the baseline for performance.
- **⚡ transactor** introduces moderate overhead in memory and allocations while maintaining comparable execution times.
- **avito** significantly increases memory consumption and allocation count, which may be critical for high-load systems.
- **aneshas** and **Thiht** show similar results, with `Thiht` consuming slightly more memory and allocations.

✅ **`transactor` remains an optimal choice** for projects requiring a balance between performance, memory consumption, and architectural clarity.

## License

Transactor is licensed under the MIT License. See [LICENSE](https://github.com/metalfm/transactor/blob/master/LICENSE)
for more
information.
