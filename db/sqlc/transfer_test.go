package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/stretchr/testify/require"
)

func createRandomTransfer(t *testing.T, fromAccountID, toAccountID int64) Transfer {
	arg := CreateTransferParams{
		FromAccountID: fromAccountID,
		ToAccountID:   toAccountID,
		Amount:        util.RandomMoney(),
	}
	transfer, err := testQueries.CreateTransfer(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfer)

	require.Equal(t, arg.FromAccountID, transfer.FromAccountID)
	require.Equal(t, arg.ToAccountID, transfer.ToAccountID)
	require.Equal(t, arg.Amount, transfer.Amount)
	require.NotZero(t, transfer.ID)
	require.NotZero(t, transfer.CreatedAt)

	return transfer
}

func TestCreateTransfer(t *testing.T) {
	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	createRandomTransfer(t, fromAccount.ID, toAccount.ID)
}

func TestGetTransfer(t *testing.T) {
	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	transfer1 := createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfer2)

	require.Equal(t, transfer1.ID, transfer2.ID)
	require.Equal(t, transfer1.FromAccountID, transfer2.FromAccountID)
	require.Equal(t, transfer1.ToAccountID, transfer2.ToAccountID)
	require.Equal(t, transfer1.Amount, transfer2.Amount)
	require.WithinDuration(t, transfer1.CreatedAt.Time, transfer2.CreatedAt.Time, 0)
}

func TestDeleteTransfer(t *testing.T) {
	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	transfer1 := createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	err := testQueries.DeleteTransfer(context.Background(), transfer1.ID)
	require.NoError(t, err)

	transfer2, err := testQueries.GetTransfer(context.Background(), transfer1.ID)
	require.Error(t, err)
	require.Empty(t, transfer2)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestListTransfers(t *testing.T) {
	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	}

	arg := ListTransfersParams{
		Limit:  5,
		Offset: 5,
	}

	transfers, err := testQueries.ListTransfers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, transfers, 5)

	for _, transfer := range transfers {
		require.NotEmpty(t, transfer)
	}
}

func TestListTransfersBetweenAccounts(t *testing.T) {
	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	for i := 0; i < 5; i++ {
		createRandomTransfer(t, fromAccount.ID, toAccount.ID)
	}
	arg := ListTransfersBetweenAccountsParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
	}
	transfers, err := testQueries.ListTransfersBetweenAccounts(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, transfers)
	for _, transfer := range transfers {
		require.Equal(t, fromAccount.ID, transfer.FromAccountID)
		require.Equal(t, toAccount.ID, transfer.ToAccountID)
	}
}

func TestListTransfersByAccount(t *testing.T) {
	account := CreateRandomAccount(t)
	otherAccount := CreateRandomAccount(t)
	for i := 0; i < 5; i++ {
		createRandomTransfer(t, account.ID, otherAccount.ID)
		createRandomTransfer(t, otherAccount.ID, account.ID)
	}
	transfers, err := testQueries.ListTransfersByAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, transfers)
	for _, transfer := range transfers {
		require.True(t, transfer.FromAccountID == account.ID || transfer.ToAccountID == account.ID)
	}
}
