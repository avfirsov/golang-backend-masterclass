package db

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTransferTx(t *testing.T) {
	store := NewStore(testPool)

	fromAccount := CreateRandomAccount(t)
	toAccount := CreateRandomAccount(t)
	amount := int64(10)
	fmt.Println(">> before:", fromAccount.Balance, toAccount.Balance)

	n := 5

	errs := make(chan error)
	results := make(chan TransferTxResult)

	for i := 0; i < n; i++ {
		go func() {
			result, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
			})

			errs <- err
			results <- result
		}()
	}

	existed := make(map[int]bool)

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)

		result := <-results
		require.NotEmpty(t, result)
		require.Equal(t, fromAccount.ID, result.Transfer.FromAccountID)
		require.Equal(t, toAccount.ID, result.Transfer.ToAccountID)
		require.Equal(t, amount, result.Transfer.Amount)
		require.NotZero(t, result.Transfer.ID)
		require.NotZero(t, result.Transfer.CreatedAt)

		_, err = store.GetTransfer(context.Background(), result.Transfer.ID)
		require.NoError(t, err)

		//check from entry
		_, err = store.GetEntry(context.Background(), result.FromEntry.ID)
		require.NoError(t, err)
		require.Equal(t, fromAccount.ID, result.FromEntry.AccountID)
		require.Equal(t, -amount, result.FromEntry.Amount)
		require.NotZero(t, result.FromEntry.ID)
		require.NotZero(t, result.FromEntry.CreatedAt)

		//check to entry
		_, err = store.GetEntry(context.Background(), result.FromEntry.ID)
		require.NoError(t, err)

		fromEntry := result.FromEntry
		require.NotEmpty(t, fromEntry)
		require.Equal(t, fromAccount.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.NotZero(t, fromEntry.ID)
		require.NotZero(t, fromEntry.CreatedAt)

		//check from entry
		_, err = store.GetEntry(context.Background(), result.ToEntry.ID)
		require.NoError(t, err)

		toEntry := result.ToEntry
		require.NotEmpty(t, toEntry)
		require.Equal(t, toAccount.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, toEntry.ID)
		require.NotZero(t, toEntry.CreatedAt)

		//todo: check accounts balance
		resultFromAccount := result.FromAccount
		require.NotEmpty(t, resultFromAccount)
		require.Equal(t, fromAccount.ID, resultFromAccount.ID)

		resultToAccount := result.ToAccount
		require.NotEmpty(t, resultToAccount)
		require.Equal(t, toAccount.ID, resultToAccount.ID)

		fmt.Println(">> tx:", resultFromAccount.Balance, resultToAccount.Balance)

		//check accounts balance
		diff1 := fromAccount.Balance - resultFromAccount.Balance
		diff2 := resultToAccount.Balance - toAccount.Balance
		require.Equal(t, diff1, diff2)
		require.Positive(t, diff1)
		require.True(t, diff1%amount == 0)
		k := int(diff1 / amount)
		require.True(t, k >= 1)
		require.True(t, k <= n)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	//check updated account's balances
	updatedFromAccount, err := store.GetAccountForUpdate(context.Background(), fromAccount.ID)
	require.NoError(t, err)
	require.Equal(t, fromAccount.Balance-int64(n)*amount, updatedFromAccount.Balance)

	updatedToAccount, err := store.GetAccountForUpdate(context.Background(), toAccount.ID)
	require.NoError(t, err)

	fmt.Println(">> after:", updatedFromAccount.Balance, updatedToAccount.Balance)

	require.Equal(t, toAccount.Balance+int64(n)*amount, updatedToAccount.Balance)

}

func TestTransferTxDeadlock(t *testing.T) {
	store := NewStore(testPool)
	acc1 := CreateRandomAccount(t)
	acc2 := CreateRandomAccount(t)
	amount := int64(10)

	n := 10
	errs := make(chan error)

	for i := 0; i < n; i++ {
		fromAccount := acc1
		toAccount := acc2
		if i%2 == 1 {
			fromAccount = acc2
			toAccount = acc1
		}
		fmt.Println(">> loop", fromAccount.ID, toAccount.ID)

		go func() {
			_, err := store.TransferTx(context.Background(), TransferTxParams{
				FromAccountID: fromAccount.ID,
				ToAccountID:   toAccount.ID,
				Amount:        amount,
			})
			errs <- err
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		require.NoError(t, err)
	}

	updatedAcc1, err := store.GetAccountForUpdate(context.Background(), acc1.ID)
	require.NoError(t, err)
	require.Equal(t, acc1.Balance, updatedAcc1.Balance)

	updatedAcc2, err := store.GetAccountForUpdate(context.Background(), acc2.ID)
	require.NoError(t, err)
	require.Equal(t, acc2.Balance, updatedAcc2.Balance)
}
