package interf

import "errors"

var (
	ErrTurnNotRun                  = errors.New("未运行 Turn")
	ErrTurnAlreadyRun              = errors.New("已运行 Turn")
	ErrTurnAlreadyFinishRun        = errors.New("已运行完成 Turn")
	ErrGroupCheckReadyFailed       = errors.New("内部错误，Group判断准备失败")
	ErrNoExecutableAction          = errors.New("没有可供执行的 Action")
	ErrNotAllowedAddActionAfterRun = errors.New("开始运行之后不能再添加 Action")
)
