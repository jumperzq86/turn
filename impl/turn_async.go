package impl

import (
	"fmt"
	"github.com/jumperzq86/turn/interf"
	"sync/atomic"
	"time"
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

type TurnAsync struct {
	group   *Group
	timer   *time.Timer
	finish  interf.Finish
	running int32
	signal  chan struct{}
}

func NewTurnAsync(groupType GroupType, deadline time.Duration, finish interf.Finish) *TurnAsync {
	return &TurnAsync{
		group:   NewGroup(groupType),
		timer:   time.NewTimer(deadline),
		finish:  finish,
		running: 0,
		signal:  make(chan struct{}, 1),
	}
}

func (this *TurnAsync) AddAction(action Action) {
	this.group.addAction(action)
}

func (this *TurnAsync) Signal() {
	if atomic.LoadInt32(&this.running) == 0 {
		fmt.Println("还未启动turn")
		return
	}
	this.signal <- struct{}{}
	return
}

func (this *TurnAsync) Run() {
	ok := atomic.CompareAndSwapInt32(&this.running, 0, 1)
	if !ok {
		fmt.Println("不能重复运行")
		return
	}

	// 退出协程时清理turn
	defer func() {
		this.clean()
	}()

	if this.group.empty() {
		fmt.Println("没有action，退出")
		return
	}

end:
	for {
		select {
		case <-this.timer.C:
			this.group.exec()
			break end

		case <-this.signal:
			ready, err := this.group.ready()
			if err != nil {
				fmt.Println("发生错误： ", err)
				return
			}

			if ready {
				this.group.exec()
				break end
			}
		}
	}

	// 流程检测与推动
	this.finish.FinishTurn()
	return
}

func (this *TurnAsync) clean() {
	atomic.StoreInt32(&this.running, 0)
	close(this.signal)
	this.timer.Stop()
	this.group = nil
	this.finish = nil
}
