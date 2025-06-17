package util

import (
	//db "github.com/avfirsov/golang-backend-masterclass/db/sqlc"
	"github.com/go-faker/faker/v4"
	"github.com/jackc/pgx/v5/pgtype"
	"math/rand"
	"time"
)

var idSeq int64 = 1

func nextID() int64 {
	idSeq++
	return idSeq
}

func RandomOwner() string {
	return faker.Name()
}

func RandomCurrency() string {
	currencies := []string{"USD", "EUR", "RUB", "CAD"}
	return currencies[rand.Intn(len(currencies))]
}

func RandomMoney() int64 {
	return rand.Int63n(1_000_000) + 1
}

func RandomAmount() int64 {
	return rand.Int63n(20_001) - 10_000 // [-10000, +10000]
}

func fakeTimestamp() pgtype.Timestamptz {
	return pgtype.Timestamptz{
		Time:  time.Now().Add(-time.Duration(rand.Int63n(int64(365*24*time.Hour.Seconds()))) * time.Second),
		Valid: true,
	}
}

//// --- Фейковые генераторы ---
//
//func FakeAccount() db.Account {
//	return db.Account{
//		ID:        nextID(),
//		Owner:     RandomOwner(),
//		Balance:   RandomMoney(),
//		Currency:  RandomCurrency(),
//		CreatedAt: fakeTimestamp(),
//	}
//}
//
//func FakeEntry(accounts ...db.Account) db.Entry {
//	var accountID int64
//	if len(accounts) > 0 {
//		accountID = accounts[rand.Intn(len(accounts))].ID
//	} else {
//		accountID = rand.Int63n(idSeq) + 1
//	}
//	return db.Entry{
//		ID:        nextID(),
//		AccountID: accountID,
//		Amount:    RandomAmount(),
//		CreatedAt: fakeTimestamp(),
//	}
//}
//
//func FakeTransfer(accounts ...db.Account) db.Transfer {
//	if len(accounts) < 2 {
//		return db.Transfer{
//			ID:            nextID(),
//			FromAccountID: rand.Int63n(idSeq) + 1,
//			ToAccountID:   rand.Int63n(idSeq) + 1,
//			Amount:        RandomMoney(),
//			CreatedAt:     fakeTimestamp(),
//		}
//	}
//	i := rand.Intn(len(accounts))
//	j := rand.Intn(len(accounts) - 1)
//	if j >= i {
//		j++
//	}
//	return db.Transfer{
//		ID:            nextID(),
//		FromAccountID: accounts[i].ID,
//		ToAccountID:   accounts[j].ID,
//		Amount:        RandomMoney(),
//		CreatedAt:     fakeTimestamp(),
//	}
//}
