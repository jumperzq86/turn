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
// 	  而是在业务代码中使用锁来进行协程同步操作（包括这里的condition接口实现和operation接口实现）
//	2.同步方式：
//    本库中放弃超时逻辑和单独协程，仅仅提供一个轮次管理的逻辑
//    即不要timer和Run协程，仅仅将 group.ready 和 group.exec 封装起来，供业务层调用

type TurnSync struct {
	group  *Group
	finish interf.Finish
}

func NewTurnSync(groupType GroupType, finish interf.Finish) *TurnSync {
	return &TurnSync{
		group:  NewGroup(groupType),
		finish: finish,
	}
}

func (this *TurnSync) AddAction(action Action) {
	this.group.addAction(action)
}

func (this *TurnSync) Run(force bool) {

	if this.group.empty() {
		fmt.Println("没有action，退出")
		return
	}

	ready, err := this.group.ready()
	if err != nil {
		fmt.Println("发生错误： ", err)
		this.clean()
		return
	}

	if !force && !ready {
		return
	}

	this.group.exec()

	// 流程检测与推动
	this.finish.FinishTurn()

	this.clean()
	return
}

func (this *TurnSync) clean() {
	this.group = nil
	this.finish = nil
}
