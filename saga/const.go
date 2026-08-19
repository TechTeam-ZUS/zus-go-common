package saga

type SagaKey struct{ v string }

// TODO: All queue and workflow key should park here
var Queue = struct {
	OrderQueue SagaKey
}{
	OrderQueue: SagaKey{"order_queue"},
}

var Workflow = struct {
	OrderWorkflow SagaKey
}{
	OrderWorkflow: SagaKey{"order_workflow"},
}

func (s SagaKey) String() string {
	return s.v
}
