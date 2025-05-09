package app_test

import (
	"context"
	"errors"
	"github.com/metalfm/transactor/internal/example/app"
	"github.com/metalfm/transactor/internal/example/app/mock"
	"github.com/metalfm/transactor/trtest/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"testing"
)

type ServiceMock struct {
	suite.Suite
	ctx     context.Context
	ctrl    *gomock.Controller
	mockTx  *mock_app.MockrepoTx
	service *app.Service[*mock_app.MockrepoTx]
}

func (slf *ServiceMock) SetupTest() {
	slf.ctx = context.Background()
	slf.ctrl = gomock.NewController(slf.T())
	slf.mockTx = mock_app.NewMockrepoTx(slf.ctrl)

	mockTr := mock_tr.NewMockTransactor[*mock_app.MockrepoTx](slf.ctrl)
	mockTr.
		EXPECT().
		InTx(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, fn func(r *mock_app.MockrepoTx) error) error {
			return fn(slf.mockTx)
		}).
		AnyTimes()

	slf.service = app.NewService[*mock_app.MockrepoTx](mockTr)
}

func (slf *ServiceMock) TearDownTest() {
	slf.ctrl.Finish()
}

func (slf *ServiceMock) TestCreateErrUser() {
	expected := errors.New("user creation error")

	slf.mockTx.EXPECT().
		CreateUser(slf.ctx, "user-name").
		Return(expected)

	err := slf.service.Create(slf.ctx, "user-name", []string{"item1", "item2"})
	slf.ErrorIs(err, expected)
}

func (slf *ServiceMock) TestCreateErrOrder() {
	expected := errors.New("order creation error")

	slf.mockTx.EXPECT().
		CreateUser(slf.ctx, "user-name").
		Return(nil)

	slf.mockTx.EXPECT().
		CreateOrder(slf.ctx, []string{"item1", "item2"}).
		Return(expected)

	err := slf.service.Create(slf.ctx, "user-name", []string{"item1", "item2"})
	slf.ErrorIs(err, expected)
}

func (slf *ServiceMock) TestCreateSuccess() {
	slf.mockTx.EXPECT().
		CreateUser(slf.ctx, "user-name").
		Return(nil)

	slf.mockTx.EXPECT().
		CreateOrder(slf.ctx, []string{"item1", "item2"}).
		Return(nil)

	err := slf.service.Create(slf.ctx, "user-name", []string{"item1", "item2"})
	slf.NoError(err)
}

func TestServiceMock(t *testing.T) {
	suite.Run(t, new(ServiceMock))
}
