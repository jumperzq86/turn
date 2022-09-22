package turn

// 条件是在具体客户端逻辑中确定
type Condition interface {
	Check(values ...TernaryValue) (TernaryValue, error)
}
