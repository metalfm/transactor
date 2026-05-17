// Package transactor documents the github.com/metalfm/transactor module.
//
// Transactor provides small, type-safe building blocks for running application
// code through a transactional boundary without coupling business logic to a
// concrete database transaction implementation. Transactions are passed through
// typed repository adapters instead of being stored in [context.Context].
//
// The core interface is defined in package tr, while database-specific
// implementations live under driver.
//
// The module includes adapters for database/sql, sqlx, and pgx:
//
//   - github.com/metalfm/transactor/tr
//   - github.com/metalfm/transactor/driver/sql/trm
//   - github.com/metalfm/transactor/driver/sqlx/trm
//   - github.com/metalfm/transactor/driver/pgx/trm
//
// Test helpers are available under github.com/metalfm/transactor/trtest/mock.
package transactor
