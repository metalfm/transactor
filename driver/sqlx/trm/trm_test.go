package trm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"

	"github.com/metalfm/transactor/driver/sqlx/trm"
)

type InTx struct {
	suite.Suite

	ctx  context.Context
	db   *sqlx.DB
	mock sqlmock.Sqlmock
	impl *trm.Impl[*mockWithTx]
}

type mockWithTx struct{}

func (m *mockWithTx) WithTx(_ trm.Transaction) *mockWithTx {
	return m
}

func (slf *InTx) SetupTest() {
	db, mock, err := sqlmock.New()
	slf.Require().NoError(err)

	slf.ctx = context.Background()
	slf.mock = mock
	slf.db = sqlx.NewDb(db, "sqlmock")
	slf.impl = trm.New(slf.db, &mockWithTx{})
}

func (slf *InTx) TearDownTest() {}

func (slf *InTx) TestSuccess() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectCommit()

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})

	slf.NoError(err)
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestRollbackOnError() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectRollback()

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return errors.New("err")
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "trm callback: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestBeginTxError() {
	slf.mock.ExpectBegin().WillReturnError(errors.New("err"))

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "begin tx: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestCommitError() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectCommit().WillReturnError(errors.New("err"))

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "commit tx: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func TestInTx(t *testing.T) {
	suite.Run(t, new(InTx))
}
