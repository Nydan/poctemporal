package main

import (
	"crypto/tls"
	"crypto/x509"
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
	w.RegisterWorkflow(poctemporal.DepositWorkflow)
	w.RegisterActivity(poctemporal.CreateTransaction)
	w.RegisterActivity(poctemporal.CreateWallet)
	w.RegisterActivity(poctemporal.CreateDeposit)

	// go func() {
	if err = w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
	// }()

	a.Client.Close()
}

type App struct {
	Client client.Client
}

func newApp(cfg poctemporal.Config) (*App, error) {
	identity := uuid.NewString()
	log.Printf("Identity: %s", identity)

	ca, err := ioutil.ReadFile(cfg.TLS.CertPoolPath)
	if err != nil {
		log.Fatal(err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(ca) {
		log.Fatal("Failed to append ca certs", err)
	}

	certTLS, err := tls.LoadX509KeyPair(cfg.TLS.CertPath, cfg.TLS.KeyPath)
	if err != nil {
		log.Fatal("openTLS", err)
	}

	c, err := client.Dial(client.Options{
		HostPort:  cfg.Temporal.HostPort,
		Namespace: cfg.Temporal.Namespace,
		Identity:  identity,
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{
				Certificates: []tls.Certificate{certTLS},
				RootCAs:      certPool,
				ServerName:   cfg.Temporal.ServerName,
			},
		},
	})
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

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if err = json.NewEncoder(w).Encode(deposit); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}
