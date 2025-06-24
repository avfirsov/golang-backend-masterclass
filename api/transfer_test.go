package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/avfirsov/golang-backend-masterclass/db/mock"
	db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/avfirsov/golang-backend-masterclass/util"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

type transferTestCase struct {
	name          string
	body          gin.H
	buildStubs    func(store *mockdb.MockStore)
	checkResponse func(recorder *httptest.ResponseRecorder)
}

func TestCreateTransferAPI(t *testing.T) {
	currency := util.RandomCurrency()
	otherCurrency := util.RandomCurrency()
	for otherCurrency == currency {
		if otherCurrency != currency {
			break
		}
		otherCurrency = util.RandomCurrency()
	}
	fromAccount := randAccount(currency)
	toAccount := randAccount(currency)
	toAccountWithOtherCurrency := randAccount(otherCurrency)
	for {
		if toAccount.ID != fromAccount.ID {
			break
		}
		toAccount = randAccount(currency)
	}
	for {
		if toAccountWithOtherCurrency.ID != fromAccount.ID {
			break
		}
		toAccountWithOtherCurrency = randAccount(otherCurrency)
	}
	amount := int64(util.RandomInt(1, 10))
	expectedResult := db.TransferTxResult{
		Transfer:    db.Transfer{ID: 1, FromAccountID: fromAccount.ID, ToAccountID: toAccount.ID, Amount: amount},
		FromAccount: fromAccount,
		ToAccount:   toAccount,
		FromEntry:   db.Entry{ID: 1, AccountID: fromAccount.ID, Amount: -amount},
		ToEntry:     db.Entry{ID: 2, AccountID: toAccount.ID, Amount: amount},
	}

	testCases := []transferTestCase{
		{
			name: "OK",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), toAccount.ID).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(expectedResult, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, expectedResult, http.StatusOK)
			},
		},
		{
			name: "FromAccountNotFound",
			body: gin.H{
				"from_account_id": 9999,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), int64(9999)).Times(1).Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusNotFound)
			},
		},
		{
			name: "ToAccountNotFound",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   9999,
				"amount":          amount,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), int64(9999)).Times(1).Return(db.Account{}, pgx.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusNotFound)
			},
		},
		{
			name: "CurrencyMismatchWithToAccount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccountWithOtherCurrency.ID,
				"amount":          amount,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), toAccountWithOtherCurrency.ID).Times(1).Return(toAccountWithOtherCurrency, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusBadRequest)
			},
		},
		{
			name: "CurrencyMismatchWithFromAccount",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        otherCurrency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusBadRequest)
			},
		},
		{
			name: "InsufficientFunds",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          fromAccount.Balance + 1000,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusBadRequest)
			},
		},
		{
			name: "InvalidRequest",
			body: gin.H{
				"from_account_id": 0,
				"to_account_id":   0,
				"amount":          -10,
				"currency":        "",
			},
			buildStubs: func(store *mockdb.MockStore) {},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusBadRequest)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        currency,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
				store.EXPECT().GetAccount(gomock.Any(), toAccount.ID).Times(1).Return(toAccount, nil)
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, pgx.ErrTxClosed)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				requireBodyMatchTransfer(t, recorder, db.TransferTxResult{}, http.StatusInternalServerError)
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/transfers", bytes.NewReader(data))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}


func requireBodyMatchTransfer(t *testing.T, recorder *httptest.ResponseRecorder, result db.TransferTxResult, expectedCode int) {
	data, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)

	var gotResult db.TransferTxResult
	require.NoError(t, json.Unmarshal(data, &gotResult))
	require.Equal(t, expectedCode, recorder.Code)
	require.Equal(t, result, gotResult)
}
