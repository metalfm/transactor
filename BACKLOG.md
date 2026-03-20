# Backlog

## Planned ideas

### 1. Generic binding helpers for adapters in `driver/sql/trm`

Goal: reduce the risk of silently forgetting `WithTx(...)` on one of adapter fields.

Current idea:
- add reusable generic helpers in `trm`, not in `tr`;
- likely shape:
  - `Bind(tx, binders...)`
  - `Into(&dstField, srcWithTx)`
  - optional `Keep(&dstField, srcValue)` for non-transactional fields.

Example target usage:

```go
func (a *Adapter) WithTx(tx trm.Transaction) *Adapter {
	out := &Adapter{}

	trm.Bind(tx,
		trm.Into(&out.repoUser, a.repoUser),
		trm.Into(&out.repoOrder, a.repoOrder),
	)

	return out
}
```

Why:
- avoids a class of bugs where one repository is accidentally left out of transactional rebinding;
- stays type-safe;
- avoids reflection, code generation, and `Bind2/Bind3/...`.

Decision:
- helpers belong to `driver/sql/trm`, because they depend on `WithTx(tx trm.Transaction)` and are not part of the generic `tr.Transactor[T]` abstraction.
