package pkg

import "time"

// TimeCondition 通过时间决定是否要 kill resource
type TimeCondition struct {
	needToKill bool
	oldDate    time.Time
	newDate    time.Time
}

func NewTimeCondition(newDate time.Time) *TimeCondition {
	c := TimeCondition{
		oldDate: time.Now(),
		newDate: newDate,
	}
	return &c
}

func (c *TimeCondition) After() {

}

func (c *TimeCondition) NeedToKill() bool {
	return c.needToKill
}
