package user

import "context"

type User struct {
	ID              string
	Provider        string
	ProviderSubject string
	Email           string
}

// Repository owns user identity resolution independently from organization membership.
type Repository interface {
	FindOrCreate(ctx context.Context, provider, providerSubject, email string) (User, error)
}
