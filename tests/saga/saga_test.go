package saga_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	"github.com/TechTeam-ZUS/zus-go-common/saga"
)

func TestSagaKey_String(t *testing.T) {
	tests := []struct {
		name     string
		key      saga.SagaKey
		expected string
	}{
		{name: "order queue", key: saga.Queue.OrderQueue, expected: "order_queue"},
		{name: "order workflow", key: saga.Workflow.OrderWorkflow, expected: "order_workflow"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.key.String())
		})
	}
}

var (
	testWorkflowRun     func() error
	testWorkflowMessage string
)

func orderSagaTestWorkflow(ctx workflow.Context) error {
	if testWorkflowRun == nil {
		return fmt.Errorf("No Workflow")
	}

	return testWorkflowRun()
}

// waitForWorker blocks until this subtest's own worker has registered.
func waitForWorker(t *testing.T, ready <-chan struct{}, initErr <-chan error) {
	t.Helper()
	select {
	case <-ready:
	case err := <-initErr:
		t.Fatalf("saga.Init failed: %v", err)
	}
}

// TestInit_ExecutesRegisteredWorkflow is an integration test: it requires a
// real Temporal server running locally (e.g. `temporal server start-dev`).
// Each case calls the real saga.Init on its own task queue and confirms the
// workflow function registered through it actually executes.
func TestInit_ExecutesRegisteredWorkflow(t *testing.T) {
	t.Setenv("SAGA_HOST", "localhost:7233")
	t.Setenv("SAGA_NAMESPACE", "default")

	tests := []struct {
		name                    string
		queue                   saga.SagaKey
		runFunc                 func() error
		expectedWorkflowMessage string
	}{
		{
			name:  "order queue",
			queue: saga.Queue.OrderQueue,
			runFunc: func() error {
				testWorkflowMessage = "queue ran"
				return nil
			},
			expectedWorkflowMessage: "queue ran",
		},
		{
			name:  "order workflow key",
			queue: saga.Workflow.OrderWorkflow,
			runFunc: func() error {
				testWorkflowMessage = "workflow ran"
				return nil
			},
			expectedWorkflowMessage: "workflow ran",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWorkflowRun = tt.runFunc
			testWorkflowMessage = ""

			ready := make(chan struct{})
			initErr := make(chan error, 1)
			go func() {
				initErr <- saga.Init(tt.queue, func(w worker.Worker) {
					w.RegisterWorkflow(orderSagaTestWorkflow)
					close(ready)
				})
			}()

			waitForWorker(t, ready, initErr)
			c := saga.GetClient()

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			run, err := c.ExecuteWorkflow(ctx, client.StartWorkflowOptions{TaskQueue: tt.queue.String()}, orderSagaTestWorkflow)
			require.NoError(t, err)
			require.NoError(t, run.Get(ctx, nil))

			assert.Equal(t, tt.expectedWorkflowMessage, testWorkflowMessage)
		})
	}
}
