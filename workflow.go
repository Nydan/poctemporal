package poctemporal

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type DepositWorkflowApp struct {
	Deposit []Deposit
}

func NewDepositWorkflow() *DepositWorkflowApp {
	return &DepositWorkflowApp{
		Deposit: []Deposit{},
	}
}

func (dw *DepositWorkflowApp) DepositWorkflow(ctx workflow.Context, r DepositRequest) (Deposit, error) {
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
	txActivity := workflow.ExecuteActivity(ctx, dw.CreateTransaction, r)
	if err = txActivity.Get(ctx, &trx); err != nil {
		return Deposit{}, err
	}

	var w Wallet
	walletActivity := workflow.ExecuteActivity(ctx, dw.CreateWallet, trx)
	if err = walletActivity.Get(ctx, &w); err != nil {
		return Deposit{}, err
	}

	var d Deposit
	depositActivity := workflow.ExecuteActivity(ctx, dw.CreateDeposit, w)
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
func (dw *DepositWorkflowApp) CreateTransaction(ctx context.Context, r DepositRequest) (Transaction, error) {
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
func (dw *DepositWorkflowApp) CreateWallet(ctx context.Context, t Transaction) (Wallet, error) {
	return Wallet{
		WalletID: uuid.NewString(),
		TxID:     t.TxID,
		Amount:   t.Amount,
		UserID:   t.UserID,
	}, nil
}

type Deposit struct {
	WorkflowID string
	DepositID  string
	WalletID   string
	TxID       string
	Amount     int
	UserID     string
}

// deposit.Create
func (dw *DepositWorkflowApp) CreateDeposit(ctx context.Context, w Wallet) (Deposit, error) {
	info := activity.GetInfo(ctx)
	deposit := Deposit{
		WorkflowID: info.WorkflowExecution.ID,
		DepositID:  uuid.NewString(),
		WalletID:   w.WalletID,
		TxID:       w.TxID,
		Amount:     w.Amount,
		UserID:     w.UserID,
	}
	dw.Deposit = append(dw.Deposit, deposit)
	return deposit, nil
}
