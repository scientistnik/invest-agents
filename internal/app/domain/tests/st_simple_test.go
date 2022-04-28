package test_domain

import (
	"github.com/scientistnik/invest-agents/internal/app/domain"
	mock_domain "github.com/scientistnik/invest-agents/internal/app/domain/tests/mocks"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestBalanceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mLogger := mock_domain.NewMockLogger(ctrl)
	mStorage := mock_domain.NewMockSimpleStorage(ctrl)
	mExchange := mock_domain.NewMockExchange(ctrl)

	mLogger.EXPECT().Info(gomock.Any())
	mLogger.EXPECT().Debug(gomock.Any()).Times(10)

	empytSlice := make([]domain.Balance, 0)
	mExchange.EXPECT().Balances(gomock.Any()).Return(empytSlice, nil)

	simple := domain.SimpleStrategy{}
	simple.Run(mStorage, []domain.Exchange{mExchange}, mLogger)
}
