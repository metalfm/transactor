package trm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/stretchr/testify/suite"

	"github.com/metalfm/transactor/driver/pgx/trm"
)

type InTx struct {
	suite.Suite

	ctx  context.Context
	mock pgxmock.PgxPoolIface
	impl *trm.Impl[*mockWithTx]
}

type mockWithTx struct{}

func (m *mockWithTx) WithTx(_ trm.Transaction) *mockWithTx {
	return m
}

func (slf *InTx) SetupTest() {
	mock, err := pgxmock.NewPool()
	slf.Require().NoError(err)

	slf.ctx = context.Background()
	slf.mock = mock
	slf.impl = trm.New(mock, &mockWithTx{})
}

func (slf *InTx) TearDownTest() {
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestSuccess() {
	slf.mock.ExpectBeginTx(pgx.TxOptions{})
	slf.mock.ExpectCommit()
	slf.mock.ExpectRollback()

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})
	slf.Require().NoError(err)
}

func (slf *InTx) TestRollbackOnError() {
	slf.mock.ExpectBeginTx(pgx.TxOptions{})
	slf.mock.ExpectRollback()

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return errors.New("err")
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "trm callback: err")
}

func (slf *InTx) TestBeginTxError() {
	slf.mock.ExpectBeginTx(pgx.TxOptions{}).WillReturnError(errors.New("err"))

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "begin tx: err")
}

func (slf *InTx) TestCommitError() {
	slf.mock.ExpectBeginTx(pgx.TxOptions{})
	slf.mock.ExpectCommit().WillReturnError(errors.New("err"))
	slf.mock.ExpectRollback()

	err := slf.impl.InTx(slf.ctx, func(_ *mockWithTx) error {
		return nil
	})

	slf.Require().Error(err)
	slf.Require().EqualError(err, "commit tx: err")
}

func TestInTx(t *testing.T) {
	suite.Run(t, new(InTx))
}
