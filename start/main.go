package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/Nydan/poctemporal"
	"github.com/google/uuid"
	"go.temporal.io/sdk/client"
)

func main() {
	a, err := newApp()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()
	router.Handle("/", a.deposit())

	srv := http.Server{
		Addr:    ":8084",
		Handler: router,
	}

	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

type App struct {
	Client client.Client
}

func newApp() (*App, error) {
	c, err := client.Dial(client.Options{})
	if err != nil {
		return &App{}, err
	}

	return &App{
		Client: c,
	}, nil
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

		wRun, err := a.Client.ExecuteWorkflow(r.Context(), workflowOpt, poctemporal.DepositWorkflow, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var deposit poctemporal.Deposit
		if err = wRun.Get(r.Context(), &deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
