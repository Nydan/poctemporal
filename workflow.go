package poctemporal

import (
	"log"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func DepositWorkflow(ctx workflow.Context, r DepositRequest) (Deposit, error) {
	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    time.Second,
		BackoffCoefficient: 2.0,
		MaximumInterval:    time.Minute,
		MaximumAttempts:    3,
	}

	options := workflow.ActivityOptions{
		// Timeout options specify when to automatically timeout Activity functions.
		StartToCloseTimeout: time.Minute,
		// Optionally provide a customized RetryPolicy.
		// Temporal retries failures by default, this is just an example.
		RetryPolicy: retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, options)

	var err error
	var trx Transaction
	txActivity := workflow.ExecuteActivity(ctx, CreateTransaction, r)
	if err = txActivity.Get(ctx, &trx); err != nil {
		return Deposit{}, err
	}

	var w Wallet
	walletActivity := workflow.ExecuteActivity(ctx, CreateWallet, trx)
	if err = walletActivity.Get(ctx, &w); err != nil {
		return Deposit{}, err
	}

	var d Deposit
	depositActivity := workflow.ExecuteActivity(ctx, CreateDeposit, w)
	if err = depositActivity.Get(ctx, &d); err != nil {
		return Deposit{}, err
	}

	return d, nil
}

type Transaction struct {
	TxID   string
	UserID string
	Amount int
}

// transaction.Create
func CreateTransaction(r DepositRequest) (Transaction, error) {
	log.Println("create transaction: ", time.Now())
	return Transaction{
		TxID:   uuid.NewString(),
		UserID: r.UserID,
		Amount: r.Amount,
	}, nil
}

type Wallet struct {
	WalletID string
	TxID     string
	Amount   int
	UserID   string
}

// wallet.Create
func CreateWallet(t Transaction) (Wallet, error) {
	log.Println("create wallet: ", time.Now())
	return Wallet{
		WalletID: uuid.NewString(),
		TxID:     t.TxID,
		Amount:   t.Amount,
		UserID:   t.UserID,
	}, nil
}

type Deposit struct {
	DepositID string
	WalletID  string
	TxID      string
	Amount    int
	UserID    string
}

// deposit.Create
func CreateDeposit(w Wallet) (Deposit, error) {
	log.Println("create deposit: ", time.Now())
	return Deposit{
		DepositID: uuid.NewString(),
		WalletID:  w.WalletID,
		TxID:      w.TxID,
		Amount:    w.Amount,
		UserID:    w.UserID,
	}, nil
}
