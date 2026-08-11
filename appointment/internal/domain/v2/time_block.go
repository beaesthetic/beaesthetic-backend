package v2

type TimeBlock struct {
	Reason string
}

func NewTimeBlock(reason string) (TimeBlock, error) {
	if reason == "" {
		return TimeBlock{}, ErrMissingRequiredData
	}
	return TimeBlock{Reason: reason}, nil
}

func (TimeBlock) EventType() CalendarEventType {
	return CalendarEventTypeTimeBlock
}

func (block *TimeBlock) ChangeReason(reason string) error {
	if reason == "" {
		return ErrMissingRequiredData
	}
	block.Reason = reason
	return nil
}
