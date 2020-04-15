package pkg

type TimeCondition struct {
}

type TimeConditionBuilder struct {
}

func NewTimeConditionBuilder() *TimeConditionBuilder {
	builder := TimeConditionBuilder{}
	return &TimeConditionBuilder
}
