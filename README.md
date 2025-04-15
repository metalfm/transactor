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
        repoUser:    repoUser,
        repoOrder: repoOrder,
    }
}

// WithTx example of a factory method for combining logic from multiple repositories
func (slf *Adapter) WithTx(tx trm.Transaction) *Adapter {
    return &Adapter{
        repoUser: slf.repoUser.WithTx(tx),
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

To run benchmarks on your machine, execute the following commands — `make up && make bench`

### Libraries Used in Comparison

The following libraries and approaches were used for benchmarking:

1. **Native** — a basic approach using the standard `sql.DB` driver from Go's standard library without additional
     abstractions.

2. **⚡ Transactor** — the tested `Transactor` library, which provides the `Transactor[T any]` interface for transaction
     management.

3. **[Avito](https://github.com/avito-tech/go-transaction-manager)** — an approach based on the transaction manager
     implementation used in Avito projects. This approach includes passing transactions through context and using custom
     interfaces.

4. **[Aneshas](https://github.com/aneshas/tx)** — an alternative library for transaction management, which also uses
     context but with a different architecture and dependency management approach.

Each approach was tested on identical scenarios to ensure an objective comparison of performance, memory consumption,
and allocation count.

### Benchmark Results

#### Execution Time (sec/op)

| Metric     | Native      | ⚡ Transactor                    | Avito                      | Aneshas                    |
|------------|-------------|---------------------------------|----------------------------|----------------------------|
| **sec/op** | 263.2µ ± 2% | 261.4µ ± 1%    -0.68% (p=0.035) | 266.3µ ± 3%    ~ (p=0.265) | 261.5µ ± 1%    ~ (p=0.056) |

#### Memory Consumption (B/op)

| Metric   | Native     | ⚡ Transactor                   | Avito                            | Aneshas                         |
|----------|------------|--------------------------------|----------------------------------|---------------------------------|
| **B/op** | 822.0 ± 3% | 895.5 ± 1%    +8.94% (p=0.000) | 1456.5 ± 1%    +77.19% (p=0.000) | 913.0 ± 1%    +11.07% (p=0.000) |

#### Allocation Count (allocs/op)

| Metric        | Native     | ⚡ Transactor                    | Avito                           | Aneshas                         |
|---------------|------------|---------------------------------|---------------------------------|---------------------------------|
| **allocs/op** | 18.00 ± 0% | 21.00 ± 5%    +16.67% (p=0.000) | 33.00 ± 0%    +83.33% (p=0.000) | 21.00 ± 5%    +16.67% (p=0.000) |

### Benchmark Analysis

#### Execution Time (sec/op):

- **native**: 263.2µs ± 2% — baseline performance.
- **⚡ transactor**: 261.4µs ± 1% — a **0.68%** reduction in time, statistically significant (p=0.035).
- **avito**: 266.3µs ± 3% — a **~1.18%** increase in time, statistically insignificant (p=0.265).
- **aneshas**: 261.5µs ± 1% — a **~0.65%** reduction in time, statistically insignificant (p=0.056).

**Conclusion**: All approaches demonstrate comparable execution times. `transactor` and `aneshas` show slight
improvements, but only for `transactor` is it statistically significant.

---

#### Memory Consumption (B/op):

- **native**: 822.0 B ± 3% — baseline memory usage.
- **⚡ transactor**: 895.5 B ± 1% — an **8.94%** increase, statistically significant (p=0.000).
- **avito**: 1456.5 B ± 1% — a **77.19%** increase, statistically significant (p=0.000).
- **aneshas**: 913.0 B ± 1% — an **11.07%** increase, statistically significant (p=0.000).

**Conclusion**: `transactor` and `aneshas` moderately increase memory usage, while `avito` significantly increases
memory consumption.

---

#### Allocation Count (allocs/op):

- **native**: 18.00 ± 0% — baseline allocation count.
- **⚡ transactor**: 21.00 ± 5% — a **16.67%** increase, statistically significant (p=0.000).
- **avito**: 33.00 ± 0% — an **83.33%** increase, statistically significant (p=0.000).
- **aneshas**: 21.00 ± 5% — a **16.67%** increase, statistically significant (p=0.000).

**Conclusion**: `transactor` and `aneshas` increase allocation counts equally but moderately. `avito` requires
significantly more allocations.

---

### Overall Conclusion

- **native** demonstrates the best performance across all metrics.
- **transactor** and **aneshas** add slight overhead in memory and allocations while maintaining comparable execution
    times.
- **avito** significantly increases memory consumption and allocation count, which may be critical for high-load
    systems.

## License

Transactor is licensed under the MIT License. See [LICENSE](https://github.com/metalfm/transactor/blob/master/LICENSE) for more
information.
