package pkg

import "time"

// TimeCondition 通过时间决定是否要 kill resource
type TimeCondition struct {
	needToKill bool
	oldDate    time.Time
	newDate    time.Time
}

func NewTimeCondition(oldDate, newDate time.Time) *TimeCondition {
	c := TimeCondition{
		oldDate: oldDate,
		newDate: newDate,
	}

	return &c
}

func (c *TimeCondition) NeedToKill() bool {
	return c.needToKill
}
