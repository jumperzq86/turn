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
// 	  而是在业务代码中使用锁来进行协程同步操作（包括这里interf中对所有接口实现）
//	2.同步方式：
//    本库中放弃超时逻辑和单独协程，仅仅提供一个轮次管理的逻辑
//    即不要timer和Run协程，直接在业务主协程中执行 AddAction/Run

type TurnAsync struct {
	group   *Group
	timer   *time.Timer
	finish  interf.Finish
	running int32
	signal  chan struct{}
}

func NewTurnAsync(groupType GroupType, duration time.Duration, finish interf.Finish) *TurnAsync {
	return &TurnAsync{
		group:   NewGroup(groupType),
		timer:   time.NewTimer(duration),
		finish:  finish,
		running: int32(interf.RunInit),
		signal:  make(chan struct{}, 1),
	}
}

func (this *TurnAsync) AddAction(action *Action) error {

	//note: 防止运行之后再添加action，避免在group中对 actionlist 加锁
	running := atomic.LoadInt32(&this.running)
	if running != int32(interf.RunInit) {
		fmt.Println("运行之后，不能再添加action.")
		return interf.ErrNotAllowedAddActionAfterRun
	}

	this.group.addAction(action)
	return nil
}

func (this *TurnAsync) Signal() error {

	running := atomic.LoadInt32(&this.running)
	if running != int32(interf.Running) {
		fmt.Println("还未启动turn")
		return interf.ErrTurnNotRun
	}

	this.signal <- struct{}{}
	return nil
}

func (this *TurnAsync) Run() error {

	ok := atomic.CompareAndSwapInt32(&this.running, int32(interf.RunInit), int32(interf.Running))
	if !ok {
		fmt.Println("不能重复运行")
		return interf.ErrTurnAlreadyRun
	}

	if this.group.empty() {
		fmt.Println("没有action，无法运行")
		atomic.StoreInt32(&this.running, int32(interf.RunInit))
		return interf.ErrNoExecutableAction
	}

	// 退出协程时清理turn
	defer func() {
		this.clean()
	}()

end:
	for {
		select {
		case <-this.timer.C:
			fmt.Println("---超时")
			this.group.exec()
			break end

		case <-this.signal:
			ready, err := this.group.ready()
			if err != nil {
				fmt.Println("发生错误： ", err)
				return interf.ErrGroupCheckReadyFailed
			}
			fmt.Println("ready: ", ready)
			if ready {
				this.group.exec()
				break end
			}
		}
	}

	// 流程检测与推动
	this.finish.FinishTurn()
	return nil
}

func (this *TurnAsync) clean() {
	atomic.StoreInt32(&this.running, int32(interf.RunFinish))
	close(this.signal)
	this.timer.Stop()
	this.group = nil
	this.finish = nil
}
