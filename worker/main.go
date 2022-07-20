package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"log"

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
		log.Fatal(err)
	}
	defer c.Close()

	w := worker.New(c, poctemporal.TaskQueueDeposit, worker.Options{})
	w.RegisterWorkflow(poctemporal.DepositWorkflow)
	w.RegisterActivity(poctemporal.CreateTransaction)
	w.RegisterActivity(poctemporal.CreateWallet)
	w.RegisterActivity(poctemporal.CreateDeposit)

	if err = w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
