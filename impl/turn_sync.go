package impl

import (
	"fmt"
	"github.com/jumperzq86/turn/interf"
)

//note： 这里注意数据同步的问题
//	Signal位于客户端主逻辑协程
//	Run位于另一个单独协程
//	Run中调用 group.ready 以及 group.exec 时，会去访问客户端业务数据，需要进行同步处理
//	考虑两种方式
//	1.异步方式：
// 	  把协程同步的工作交给业务层代码，本库中不用考虑协程同步操作
// 	  而是在业务代码中使用锁来进行协程同步操作（包括这里interf中对所有接口实现）
//	2.同步方式：
//    本库中放弃超时逻辑和单独协程，仅仅提供一个轮次管理的逻辑
//    即不要timer和Run协程，直接在业务主协程中执行 AddAction/Run

type TurnSync struct {
	group   *Group
	finish  interf.Finish
	running interf.RunStatus
}

func NewTurnSync(groupType GroupType, finish interf.Finish) *TurnSync {
	return &TurnSync{
		group:   NewGroup(groupType),
		finish:  finish,
		running: interf.RunInit,
	}
}

func (this *TurnSync) AddAction(action *Action) error {
	if this.running != interf.RunInit {
		fmt.Println("非未运行状态，不能添加action")
		return interf.ErrNotAllowedAddActionAfterRun
	}
	this.group.addAction(action)
	return nil
}

//note: 每次有玩家做出决策就调用Run(false)，超时时调用Run(true)
//	可以通过返回值判断是否执行了turn
func (this *TurnSync) Run(force bool) (bool, error) {

	if this.running == interf.RunFinish {
		fmt.Println("运行完成状态，不能开启运行，退出")
		return false, interf.ErrTurnAlreadyFinishRun
	}

	this.running = interf.Running

	if this.group.empty() {
		fmt.Println("没有action，退出")
		return false, interf.ErrNoExecutableAction
	}

	ready, err := this.group.ready()
	if err != nil {
		fmt.Println("发生错误： ", err)
		this.clean()
		return false, interf.ErrGroupCheckReadyFailed
	}

	if !force && !ready {
		fmt.Printf("not ready, force: %v, ready: %v\n", force, ready)
		return false, nil
	}

	fmt.Println("ready ok or force.")

	this.group.exec()

	// 流程检测与推动
	this.finish.FinishTurn()

	this.clean()

	return true, nil
}

func (this *TurnSync) clean() {
	this.running = interf.RunFinish
	this.group = nil
	this.finish = nil
}
