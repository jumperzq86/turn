package main

import (
	"fmt"
	"github.com/jumperzq86/turn/impl"
	"github.com/jumperzq86/turn/interf"
	"time"
)

var checktag1 = true
var checktag2 = true
var checktag3 = true

func main() {
	var c1 ConditionS1
	var o1 OperationS
	var cl1 CleanerS
	a1, err := impl.NewAction(3, &c1, &o1, &cl1)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	var c2 ConditionS2
	var o2 OperationS
	var cl2 CleanerS
	a2, err := impl.NewAction(3, &c2, &o2, &cl2)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}
	//
	var c3 ConditionS3
	var o3 OperationS
	var cl3 CleanerS
	a3, err := impl.NewAction(2, &c3, &o3, &cl3)
	if err != nil {
		fmt.Println("err: ", err)
		return
	}

	//var f FinishS
	//turnSync := impl.NewTurnSync(impl.PriorActiveAction, &f)
	//turnSync.AddAction(a1)
	//turnSync.AddAction(a2)
	//turnSync.AddAction(a3)
	//
	//turnSync.Run(false)
	//
	//fmt.Println("--------------------")
	//checktag1 = false
	//checktag2 = false
	//checktag3 = true
	//
	//turnSync.Run(false)

	var f FinishS
	//checktag1 = false
	checktag2 = false
	checktag3 = false
	turnAsync := impl.NewTurnAsync(impl.PriorActiveAction, 3*time.Second, &f)
	turnAsync.AddAction(a1)
	turnAsync.AddAction(a2)
	turnAsync.AddAction(a3)

	go turnAsync.Run()
	time.Sleep(1 * time.Second)
	turnAsync.Signal()

	time.Sleep(5 * time.Second)
}

type ConditionS1 struct {
}

func (this *ConditionS1) Check(values ...interf.TernaryValue) (interf.TernaryValue, error) {
	if checktag1 {
		return interf.TernaryInit, nil
	} else {
		return interf.TernaryDeactive, nil
	}
}

type ConditionS2 struct {
}

func (this *ConditionS2) Check(values ...interf.TernaryValue) (interf.TernaryValue, error) {
	if checktag2 {
		return interf.TernaryInit, nil
	} else {
		return interf.TernaryActive, nil
	}
}

type ConditionS3 struct {
}

func (this *ConditionS3) Check(values ...interf.TernaryValue) (interf.TernaryValue, error) {
	if checktag3 {
		return interf.TernaryInit, nil
	} else {
		return interf.TernaryActive, nil
	}
}

type CleanerS struct {
}

func (this *CleanerS) Clean() error {
	return nil
}

type OperationS struct {
}

func (this *OperationS) OperateInit() error {
	fmt.Println("operate init.")
	return nil
}

func (this *OperationS) OperateActive() error {
	fmt.Println("operate active.")
	return nil
}

func (this *OperationS) OperateDeactive() error {
	fmt.Println("operate deactive.")
	return nil
}

type FinishS struct {
}

func (this *FinishS) FinishTurn() {
	fmt.Println("finish turn")
	return
}
