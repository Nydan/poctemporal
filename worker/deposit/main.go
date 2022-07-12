package main

import (
	"log"

	"github.com/Nydan/poctemporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	w := worker.New(c, poctemporal.TaskQueueDeposit, worker.Options{})
	w.RegisterWorkflow(poctemporal.DepositWorkflow)
	w.RegisterActivity(poctemporal.CreateDeposit)

	if err = w.Run(worker.InterruptCh()); err != nil {
		log.Fatal(err)
	}
}
