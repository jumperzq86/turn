package interf

// 三元逻辑值
type TernaryValue int

const (
	TernaryNone     TernaryValue = -1
	TernaryInit     TernaryValue = 0
	TernaryActive   TernaryValue = 1
	TernaryDeactive TernaryValue = 2
)

func (this TernaryValue) String() string {
	switch this {
	case 0:
		return "TernaryInit"
	case 1:
		return "TernaryActive"
	case 2:
		return "TernaryDeactive"
	default:
		return "TernaryNone"
	}
}

// 三元与
func TernaryAnd(values ...TernaryValue) TernaryValue {
	existInit := false
	for _, value := range values {
		if value == TernaryDeactive {
			return TernaryDeactive
		} else if value == TernaryInit {
			existInit = true
		}
	}
	if !existInit {
		return TernaryActive
	}

	return TernaryInit
}

// 三元或
func TernaryOr(values ...TernaryValue) TernaryValue {
	existInit := false
	for _, value := range values {
		if value == TernaryActive {
			return TernaryActive
		} else if value == TernaryInit {
			existInit = true
		}
	}
	if !existInit {
		return TernaryDeactive
	}

	return TernaryInit
}

// 三元非
func TernaryNot(value TernaryValue) TernaryValue {
	switch value {
	case TernaryActive:
		return TernaryDeactive
	case TernaryDeactive:
		return TernaryActive
	default:
		return TernaryInit
	}
}
