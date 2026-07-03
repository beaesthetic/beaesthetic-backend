package domain

type AppointmentServiceRef struct {
	Name string
}

type AppointmentService struct {
	ID    string
	Name  string
	Price float64
	Tags  []string
	Color *string
}
