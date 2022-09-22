package turn

type Action struct {
	priority int
	Condition
	Operation
	Cleaner
}

func NewAction(priority int, condition Condition, operation Operation, cleaner Cleaner) (*Action, error) {
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

func (this *Action) Exec() error {
	c, err := this.Check()
	if err != nil {
		return err
	}

	switch c {
	case TernaryInit:
		err = this.OperateInit()
		if err != nil {
			return err
		}
	case TernaryActive:
		err = this.OperateActive()
		if err != nil {
			return err
		}

	case TernaryDeactive:
		err = this.OperateDeactive()
		if err != nil {
			return err
		}

	}

	return nil
}
