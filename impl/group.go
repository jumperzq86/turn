package impl

import (
	"fmt"
	"github.com/jumperzq86/turn/interf"
	"math"
	"sort"
)

type ActionList []*Action

func (this ActionList) Len() int {
	return len(this)
}

func (this ActionList) Less(i, j int) bool {
	if this[i].priority > this[j].priority {
		return true
	}
	return false
}

func (this ActionList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

//note: group 三种执行方式
//  1. 可以具有或者不具有相同优先级的action，按照优先级降序执行完所有action，执行内容按照action 状态来选择 operation_list_active / operation_list_deactive / operation_list_init
//  这种方式会执行所有action
//  执行时机为所有action做出决策
//  超时时按照各种状态对应 operation执行
//  2. 可以具有或者不具有相同优先级的action，按照优先级降序执行所有 active action
//  这种方式只需要执行 active action
//  执行时机为所有action做出决策
//  超时时只执行active action
//  3. 可以具有或者不具有相同优先级的action，只执行当前优先级最高的active actions
//  这种方式执行最高优先级的active action，但是对于相同优先级的 action 执行顺序无法保证
//  所执行最高等级的action中，只执行其中的active action
//  执行时机为 判断 没有比x优先级更高的init action / active action， 并且没有与x优先级相等的init action
//  超时时只执行最高active等级的active action

type GroupType int

const (
	AllAction         GroupType = 1
	ActiveAction      GroupType = 2
	PriorActiveAction GroupType = 3
)

type Group struct {
	groupType  GroupType
	actionList ActionList
}

func NewGroup(groupType GroupType) *Group {
	return &Group{
		groupType:  groupType,
		actionList: make([]*Action, 0),
	}
}

func (this *Group) addAction(action *Action) {
	fmt.Printf("添加action %d，优先级: %d\n", action.id, action.priority)
	this.actionList = append(this.actionList, action)
}

func (this *Group) empty() bool {
	return len(this.actionList) == 0
}

//判断当前（超时前）是否能够执行action，这里需要根据grouptype来判断
func (this *Group) ready() (bool, error) {
	//前两种方式中，没有人未决即可开始执行group
	if this.groupType == AllAction || this.groupType == ActiveAction {
		for _, action := range this.actionList {
			fmt.Printf("action: %d, check\n", action.id)
			ternary, err := action.Check()
			if err != nil {
				return false, err
			}
			if ternary == interf.TernaryInit {
				return false, nil
			}
		}
		return true, nil
	}

	//第三种方式中，
	//1. 查找当前优先级最高的active action，若是比该优先级更高或者相等的action中没有init，则应该执行该group(active action)
	//2. 在所有action都是deactive action时，则应该执行该group
	highestInitLevel, highestActiveLevel, err := this.findHighestLevelAction()
	if err != nil {
		return false, err
	}

	//此处注意 highestActiveLevel == highestInitLevel 需要等待init 玩家决策
	if highestActiveLevel > highestInitLevel {
		return true, nil
	}
	if highestActiveLevel == math.MinInt && highestInitLevel == math.MinInt {
		return true, nil
	}

	return false, nil
}

func (this *Group) exec() error {
	//排序
	sort.Sort(this.actionList)
	switch this.groupType {

	case AllAction:
		err := this.execAllAction()
		if err != nil {
			return err
		}

	case ActiveAction:
		err := this.execActiveAction()
		if err != nil {
			return err
		}

	case PriorActiveAction:
		err := this.execPriorActiveAction()
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *Group) execAllAction() error {
	fmt.Printf("execAllAction\n")

	for _, action := range this.actionList {
		ternary, err := action.Check()
		if err != nil {
			return err
		}
		fmt.Printf("action: %d, exec: %s\n", action.id, ternary)
		if err = action.Exec(ternary); err != nil {
			return err
		}
		fmt.Printf("action: %d, clean\n", action.id)
		action.Clean()
	}
	return nil
}

func (this *Group) execActiveAction() error {
	fmt.Printf("execActiveAction\n")

	for _, action := range this.actionList {
		ternary, err := action.Check()
		if err != nil {
			return err
		}
		if ternary == interf.TernaryActive {
			fmt.Printf("action: %d, exec: %s\n", action.id, ternary)
			if err = action.Exec(ternary); err != nil {
				return err
			}
		}
		fmt.Printf("action: %d, clean\n", action.id)
		action.Clean()
	}
	return nil
}

func (this *Group) execPriorActiveAction() error {
	fmt.Printf("execPriorActiveAction\n")

	_, highestActiveLevel, err := this.findHighestLevelAction()
	if err != nil {
		return err
	}

	//没有active action，由turn层面推动流程
	if highestActiveLevel == math.MinInt {
		for _, action := range this.actionList {
			action.Clean()
		}
		return nil
	}

	fmt.Printf("highestActiveLevel: %d\n", highestActiveLevel)

	//执行最高优先级active action，可能不止一个
	for _, action := range this.actionList {
		if action.priority == highestActiveLevel {
			ternary, err := action.Check()
			if err != nil {
				return err
			}
			if ternary == interf.TernaryActive {
				fmt.Printf("action: %d, exec: %s\n", action.id, ternary)
				action.Exec(ternary)
			}
		}
		fmt.Printf("action: %d, clean\n", action.id)
		action.Clean()
	}
	return nil
}

func (this *Group) findHighestLevelAction() (int, int, error) {
	//排序
	sort.Sort(this.actionList)

	highestInitLevel := math.MinInt
	highestActiveLevel := math.MinInt

	for _, action := range this.actionList {
		ternary, err := action.Check()
		if err != nil {
			return highestInitLevel, highestActiveLevel, err
		}

		if highestInitLevel == math.MinInt && ternary == interf.TernaryInit {
			highestInitLevel = action.priority
		}

		if highestActiveLevel == math.MinInt && ternary == interf.TernaryActive {
			highestActiveLevel = action.priority
		}

		if highestInitLevel != math.MinInt && highestActiveLevel != math.MinInt {
			break
		}
	}

	return highestInitLevel, highestActiveLevel, nil
}
