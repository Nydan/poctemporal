package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Nydan/poctemporal"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	cfg, err := poctemporal.Load("./development.yaml")
	if err != nil {
		log.Fatal(err)
	}

	a, err := newApp(cfg)
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	router.Handle("/", a.deposit())
	router.Handle("/async", a.depositAsync())
	router.Handle("/list", a.list())

	srv := http.Server{
		Addr:    ":8084",
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Println("Creating workers")
	w := worker.New(a.Client, poctemporal.TaskQueueDeposit, worker.Options{})
	w.RegisterWorkflow(a.Deposit.DepositWorkflow)
	w.RegisterActivity(a.Deposit.CreateTransaction)
	w.RegisterActivity(a.Deposit.CreateWallet)
	w.RegisterActivity(a.Deposit.CreateDeposit)

	if err = w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}

	a.Client.Close()
}

type App struct {
	Client  client.Client
	Deposit *poctemporal.DepositWorkflowApp
}

func newApp(cfg poctemporal.Config) (*App, error) {
	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
	})
	if err != nil {
		return &App{}, err
	}

	dw := poctemporal.NewDepositWorkflow()
	return &App{
		Client:  c,
		Deposit: dw,
	}, nil
}

func (a *App) list() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(a.Deposit.Deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (a *App) deposit() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req poctemporal.DepositRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		workflowOpt := client.StartWorkflowOptions{
			ID:        uuid.NewString(),
			TaskQueue: poctemporal.TaskQueueDeposit,
		}

		wRun, err := a.Client.ExecuteWorkflow(r.Context(), workflowOpt, "DepositWorkflow", req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var deposit poctemporal.Deposit
		if err = wRun.Get(r.Context(), &deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func (a *App) depositAsync() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req poctemporal.DepositRequest
		err = json.Unmarshal(body, &req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		workflowOpt := client.StartWorkflowOptions{
			ID:        uuid.NewString(),
			TaskQueue: poctemporal.TaskQueueDeposit,
		}

		_, err = a.Client.ExecuteWorkflow(r.Context(), workflowOpt, "DepositWorkflow", req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
	})
}
