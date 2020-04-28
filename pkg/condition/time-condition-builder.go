package pkg

type TimeConditionBuilder struct {
}

func (b *TimeConditionBuilder) Build() *TimeCondition {
	return &TimeCondition{}
}

func NewTimeConditionBuilder() *TimeConditionBuilder {
	builder := TimeConditionBuilder{}
	return &builder
}
