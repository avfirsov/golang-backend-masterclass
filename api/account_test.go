package api

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/avfirsov/golang-backend-masterclass/db/mock"
	db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func randAccount(currency string) db.Account {
	return db.Account{
		ID:       int64(util.RandomInt(1, 100)),
		Owner:    util.RandomOwner(),
		Balance:  rand.Int63n(100) + 10,
		Currency: currency,
	}
}
func TestGetAccountAPI(t *testing.T) {
	currency := util.RandomCurrency()
	account := randAccount(currency)

	testCases := []struct {
		name          string
		accountId     int64
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(account, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchAccount(t, recorder, account)
			},
		},
		{
			name:      "NotFound",
			accountId: 1000,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(1000))).Times(1).Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			accountId: account.ID,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(account.ID)).Times(1).Return(db.Account{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name:      "InvalidId",
			accountId: -1,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), gomock.Eq(int64(-1))).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("/accounts/%d", tc.accountId), nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func requireBodyMatchAccount(t *testing.T, recorder *httptest.ResponseRecorder, account db.Account) {
	data, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)

	var gotAccount db.Account
	require.NoError(t, json.Unmarshal(data, &gotAccount))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, account, gotAccount)
}

func requireBodyMatchAccounts(t *testing.T, recorder *httptest.ResponseRecorder, accounts []db.Account) {
	data, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)

	var gotAccounts []db.Account
	require.NoError(t, json.Unmarshal(data, &gotAccounts))

	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, accounts, gotAccounts)
}

func TestListAccountsAPI(t *testing.T) {
	currency := util.RandomCurrency()
	accounts := make([]db.Account, 5)
	for i := 0; i < 5; i++ {
		accounts[i] = randAccount(currency)
	}

	testCases := []struct {
		name          string
		query         string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:  "OK",
			query: "limit=5&page=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}).Times(1).Return(accounts, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchAccounts(t, recorder, accounts)
			},
		},
		{
			name:  "InternalError",
			query: "limit=5&page=1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().ListAccounts(gomock.Any(), db.ListAccountsParams{
					Limit:  5,
					Offset: 0,
				}).Times(1).Return(nil, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:       "InvalidLimit",
			query:      "limit=2&page=1",
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:       "InvalidOffset",
			query:      "limit=5&page=-1",
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/accounts?" + tc.query
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}
