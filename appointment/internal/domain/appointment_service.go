package domain

type AppointmentServiceRef struct {
	Name string
}

type AppointmentService struct {
	ID    string
	Name  string
	Tags  []string
	Color *string
}
