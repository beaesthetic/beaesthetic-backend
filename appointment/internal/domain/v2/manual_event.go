package v2

type ManualEvent struct {
	Title       string
	Description string
	Location    *string
}

func NewManualEvent(title string, description string, location *string) (ManualEvent, error) {
	if title == "" {
		return ManualEvent{}, ErrMissingRequiredData
	}
	return ManualEvent{Title: title, Description: description, Location: location}, nil
}

func (ManualEvent) EventType() CalendarEventType {
	return CalendarEventTypeManual
}

func (event *ManualEvent) Rename(title string) error {
	if title == "" {
		return ErrMissingRequiredData
	}
	event.Title = title
	return nil
}

func (event *ManualEvent) ChangeDescription(description string) {
	event.Description = description
}

func (event *ManualEvent) ChangeLocation(location *string) {
	event.Location = location
}
