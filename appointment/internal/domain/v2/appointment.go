package v2

import "time"

type CustomerRef struct {
	ID          string
	DisplayName string
}

func NewCustomerRef(id string, displayName string) (CustomerRef, error) {
	if id == "" {
		return CustomerRef{}, ErrMissingRequiredData
	}
	return CustomerRef{ID: id, DisplayName: displayName}, nil
}

type ServiceItem struct {
	ServiceID   *string
	ServiceName string
	Price       *float64
	Position    int
}

func NewServiceItem(serviceID *string, serviceName string, price *float64, position int) (ServiceItem, error) {
	if serviceName == "" || position < 0 {
		return ServiceItem{}, ErrMissingRequiredData
	}
	return ServiceItem{ServiceID: serviceID, ServiceName: serviceName, Price: price, Position: position}, nil
}

type Appointment struct {
	Customer  CustomerRef
	Services  []ServiceItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewAppointment(customer CustomerRef, services []ServiceItem, now time.Time) (Appointment, error) {
	if customer.ID == "" {
		return Appointment{}, ErrMissingRequiredData
	}
	return Appointment{
		Customer:  customer,
		Services:  normalizeServicePositions(services),
		CreatedAt: now.UTC(),
		UpdatedAt: now.UTC(),
	}, nil
}

func ReconstituteAppointment(customer CustomerRef, services []ServiceItem, createdAt time.Time, updatedAt time.Time) (Appointment, error) {
	if customer.ID == "" {
		return Appointment{}, ErrMissingRequiredData
	}
	return Appointment{
		Customer:  customer,
		Services:  normalizeServicePositions(services),
		CreatedAt: createdAt.UTC(),
		UpdatedAt: updatedAt.UTC(),
	}, nil
}

func (Appointment) EventType() CalendarEventType {
	return CalendarEventTypeAppointment
}

func (appointment *Appointment) ReplaceServices(services []ServiceItem, now time.Time) {
	appointment.Services = normalizeServicePositions(services)
	appointment.UpdatedAt = now.UTC()
}

func normalizeServicePositions(services []ServiceItem) []ServiceItem {
	out := make([]ServiceItem, len(services))
	copy(out, services)
	for i := range out {
		out[i].Position = i
	}
	return out
}
