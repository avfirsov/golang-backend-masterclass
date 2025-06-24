package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (server *Server) createTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	if err := server.validAccountFrom(req.FromAccountID, req.Currency, req.Amount); err != nil {
		fmt.Println("validAccountFrom error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		}
		return
	}

	if err := server.validAccountTo(req.ToAccountID, req.Currency); err != nil {
		fmt.Println("validAccountTo error", err)
		if errors.Is(err, pgx.ErrNoRows) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
		} else {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
		}
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		fmt.Println("TransferTx error", err)
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}

func (server *Server) validAccountFrom(accountId int64, currency string, amount int64) error {
	account, err := server.getValidAccount(accountId, currency)
	if err != nil {
		return err
	}

	if account.Balance < amount {
		return fmt.Errorf("account [%d] balance %d is less than request amount %d", accountId, account.Balance, amount)
	}

	return nil
}

func (server *Server) validAccountTo(accountId int64, currency string) error {
	_, err := server.getValidAccount(accountId, currency)
	return err
}

func (server *Server) getValidAccount(accountId int64, currency string) (db.Account, error) {
	account, err := server.store.GetAccount(context.Background(), accountId)
	if err != nil {
		return db.Account{}, err
	}

	if account.Currency != currency {
		return db.Account{}, fmt.Errorf("account [%d] currency %s does not match request currency %s", accountId, account.Currency, currency)
	}

	return account, nil
}
