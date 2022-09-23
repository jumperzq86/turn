package impl

import (
	"github.com/jumperzq86/turn/interf"
)

type Action struct {
	priority int
	interf.Condition
	interf.Operation
	interf.Cleaner
}

func NewAction(priority int, condition interf.Condition, operation interf.Operation, cleaner interf.Cleaner) (*Action, error) {
	if condition == nil || operation == nil || cleaner == nil {
		return nil, nil
	}

	return &Action{
		priority:  priority,
		Condition: condition,
		Operation: operation,
		Cleaner:   cleaner,
	}, nil

}

func (this *Action) Exec(c interf.TernaryValue) error {
	var err error

	switch c {
	case interf.TernaryInit:
		err = this.OperateInit()
		if err != nil {
			return err
		}
	case interf.TernaryActive:
		err = this.OperateActive()
		if err != nil {
			return err
		}

	case interf.TernaryDeactive:
		err = this.OperateDeactive()
		if err != nil {
			return err
		}

	}

	return nil
}
