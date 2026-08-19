package saga

import (
	"fmt"

	"github.com/TechTeam-ZUS/zus-go-common/config"
	"github.com/TechTeam-ZUS/zus-go-common/retry"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

type sagaClientStruct struct {
	client client.Client
}

var sagaClient *sagaClientStruct

func Init(queue SagaKey, workerRegistration func(w worker.Worker)) error {
	cfg := config.LoadSaga()

	var c client.Client
	err := retry.Do(cfg.RetryCount, retry.RetryDelay, func() error {
		dialed, err := client.Dial(client.Options{
			HostPort:  cfg.Host,
			Namespace: cfg.Namespace,
		})
		if err != nil {
			return err
		}

		c = dialed
		return nil
	}, "Temporal Connection")
	if err != nil {
		return fmt.Errorf("Failed to connect to temporal: %w", err)
	}
	defer c.Close()

	sagaClient = &sagaClientStruct{
		client: c,
	}

	w := worker.New(c, queue.String(), worker.Options{})
	// Register workerflow
	workerRegistration(w)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		return fmt.Errorf("Failed to run temporal worker, QueueKey: %s\nError: %w", queue, err)
	}

	return nil
}

// GetClient returns the Temporal client established by Init
func GetClient() client.Client {
	if sagaClient == nil {
		return nil
	}

	return sagaClient.client
}
