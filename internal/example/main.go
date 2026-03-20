package main

import (
	"context"
	"database/sql"

	_ "github.com/lib/pq"

	"github.com/metalfm/transactor/driver/sql/trm"
	"github.com/metalfm/transactor/internal/example/app"
	"github.com/metalfm/transactor/internal/example/svc"
)

func main() {
	var db *sql.DB

	repoUser := svc.NewRepoUser(db)
	repoOrder := svc.NewRepoOrder(db)

	adapter := svc.NewAdapter(repoUser, repoOrder)
	tr := trm.New(db, adapter)

	ctx := context.Background()
	service := app.NewService(tr)

	_ = service.Create(ctx, "John Doe", []string{"item1", "item2"})
}
