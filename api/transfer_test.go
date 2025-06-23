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
	fromAccount := randAccount()
	toAccount := randAccount()
	currency := fromAccount.Currency
	amount := util.RandomMoney()

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
				store.EXPECT().TransferTx(gomock.Any(), gomock.Any()).Times(1).Return(db.TransferTxResult{}, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
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
				require.Equal(t, http.StatusNotFound, recorder.Code)
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
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name: "CurrencyMismatch",
			body: gin.H{
				"from_account_id": fromAccount.ID,
				"to_account_id":   toAccount.ID,
				"amount":          amount,
				"currency":        "EUR",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetAccount(gomock.Any(), fromAccount.ID).Times(1).Return(fromAccount, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
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

// randAccount и requireBodyMatchTransfer можно доработать при необходимости
type ginH map[string]interface{}

func randTransfer() db.Transfer {
	return db.Transfer{
		ID:       int64(util.RandomInt(1, 100)),
		FromAccountID:    int64(util.RandomInt(1, 100)),
		ToAccountID:  int64(util.RandomInt(1, 100)),
		Amount: util.RandomMoney(),
	}
}

// Можно реализовать requireBodyMatchTransfer для глубокого сравнения результата, если нужно
func requireBodyMatchTransfer(t *testing.T, recorder *httptest.ResponseRecorder, result db.TransferTxResult) {
	data, err := io.ReadAll(recorder.Body)
	require.NoError(t, err)

	var gotResult db.TransferTxResult
	require.NoError(t, json.Unmarshal(data, &gotResult))
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, result, gotResult)
}
