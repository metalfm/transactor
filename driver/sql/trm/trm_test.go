package trm

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/suite"
)

type InTx struct {
	suite.Suite
	ctx  context.Context
	db   *sql.DB
	mock sqlmock.Sqlmock
	impl *impl[*mockWithTx]
}

type mockWithTx struct{}

func (m *mockWithTx) WithTx(_ Transaction) *mockWithTx {
	return m
}

func (slf *InTx) SetupTest() {
	var err error
	slf.db, slf.mock, err = sqlmock.New()
	slf.Require().NoError(err)

	slf.ctx = context.Background()
	slf.impl = New(slf.db, &mockWithTx{})
}

func (slf *InTx) TearDownTest() {
	slf.db.Close()
}

func (slf *InTx) TestSuccess() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectCommit()

	err := slf.impl.InTx(slf.ctx, func(repo *mockWithTx) error {
		return nil
	})

	slf.NoError(err)
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestRollbackOnError() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectRollback()

	err := slf.impl.InTx(slf.ctx, func(repo *mockWithTx) error {
		return errors.New("err")
	})

	slf.Error(err)
	slf.EqualError(err, "trm callback: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestBeginTxError() {
	slf.mock.ExpectBegin().WillReturnError(errors.New("err"))

	err := slf.impl.InTx(slf.ctx, func(repo *mockWithTx) error {
		return nil
	})

	slf.Error(err)
	slf.EqualError(err, "begin tx: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func (slf *InTx) TestCommitError() {
	slf.mock.ExpectBegin()
	slf.mock.ExpectCommit().WillReturnError(errors.New("err"))

	err := slf.impl.InTx(slf.ctx, func(repo *mockWithTx) error {
		return nil
	})

	slf.Error(err)
	slf.EqualError(err, "commit tx: err")
	slf.NoError(slf.mock.ExpectationsWereMet())
}

func TestInTx(t *testing.T) {
	suite.Run(t, new(InTx))
}
