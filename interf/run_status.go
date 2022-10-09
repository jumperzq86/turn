package interf

type RunStatus int32

const (
	RunInit   RunStatus = 0 // 未运行
	Running   RunStatus = 1 // 运行中
	RunFinish RunStatus = 2 // 运行完
)
