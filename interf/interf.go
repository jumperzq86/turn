package interf

//note：以下接口需要客户端代码根据业务逻辑进行实现
//  若是采用 TurnAsync 即异步方式，那么以下接口实现内部需要对数据进行同步操作

// action中条件判断接口
type Condition interface {
	Check(values ...TernaryValue) (TernaryValue, error)
}

// action中数据清理接口
type Cleaner interface {
	Clean() error
}

// action中操作接口
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

// turn中finish接口
type Finish interface {
	FinishTurn()
}
