package turn

type OperationType int

const (
	OperationInit     OperationType = 0
	OperationActive   OperationType = 1
	OperationDeactive OperationType = 2
)

type Operation interface {
	OperateInit() error
	OperateActive() error
	OperateDeactive() error
}
