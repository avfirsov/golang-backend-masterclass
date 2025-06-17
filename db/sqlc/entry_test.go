package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/stretchr/testify/require"
)

func createRandomEntry(t *testing.T, accountID int64) Entry {
	arg := CreateEntryParams{
		AccountID: accountID,
		Amount:    util.RandomMoney(),
	}
	entry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry)

	require.Equal(t, arg.AccountID, entry.AccountID)
	require.Equal(t, arg.Amount, entry.Amount)
	require.NotZero(t, entry.ID)
	require.NotZero(t, entry.CreatedAt)

	return entry
}

func TestCreateEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	createRandomEntry(t, account.ID)
}

func TestGetEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	entry1 := createRandomEntry(t, account.ID)
	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, entry1.Amount, entry2.Amount)
	require.WithinDuration(t, entry1.CreatedAt.Time, entry2.CreatedAt.Time, 0)
}

func TestDeleteEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	entry1 := createRandomEntry(t, account.ID)
	err := testQueries.DeleteEntry(context.Background(), entry1.ID)
	require.NoError(t, err)

	entry2, err := testQueries.GetEntry(context.Background(), entry1.ID)
	require.Error(t, err)
	require.Empty(t, entry2)
	require.ErrorIs(t, err, sql.ErrNoRows)
}

func TestListEntries(t *testing.T) {
	account := CreateRandomAccount(t)
	for i := 0; i < 10; i++ {
		createRandomEntry(t, account.ID)
	}

	arg := ListEntriesParams{
		Limit:  5,
		Offset: 5,
	}

	entries, err := testQueries.ListEntries(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, entries, 5)

	for _, entry := range entries {
		require.NotEmpty(t, entry)
	}
}

func TestUpdateEntry(t *testing.T) {
	account := CreateRandomAccount(t)
	entry1 := createRandomEntry(t, account.ID)

	arg := UpdateEntryParams{
		ID:     entry1.ID,
		Amount: util.RandomMoney(),
	}

	entry2, err := testQueries.UpdateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, entry2)

	require.Equal(t, entry1.ID, entry2.ID)
	require.Equal(t, entry1.AccountID, entry2.AccountID)
	require.Equal(t, arg.Amount, entry2.Amount)
}

func TestListEntriesByAccount(t *testing.T) {
	account := CreateRandomAccount(t)
	for i := 0; i < 5; i++ {
		createRandomEntry(t, account.ID)
	}
	entries, err := testQueries.ListEntriesByAccount(context.Background(), account.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entries)
	for _, entry := range entries {
		require.Equal(t, account.ID, entry.AccountID)
	}
}
