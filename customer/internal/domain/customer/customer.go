package customer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var phonePrefixRegex = regexp.MustCompile(`^\+?[0-9]{2}`)

type Customer struct {
	ID      string
	Name    string
	Surname string
	Email   *string
	Phone   *Phone
	Note    string
}

type Phone struct {
	Prefix string
	Number string
}

func New(name, surname string, email *string, phone *Phone, note string) (Customer, error) {
	if strings.TrimSpace(name) == "" {
		return Customer{}, fmt.Errorf("customer name is required")
	}
	return Customer{ID: uuid.NewString(), Name: name, Surname: surname, Email: cleanStringPtr(email), Phone: phone, Note: note}, nil
}

func ParsePhone(value string) (*Phone, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, nil
	}
	prefix := phonePrefixRegex.FindString(value)
	if prefix == "" || len(value) == len(prefix) {
		return nil, fmt.Errorf("invalid phone number format: %s", value)
	}
	return &Phone{Prefix: prefix, Number: value[len(prefix):]}, nil
}

func (phone Phone) FullNumber() string { return phone.Prefix + phone.Number }

func (customer Customer) Update(name, surname, email, phone, note *string) (Customer, error) {
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return Customer{}, fmt.Errorf("customer name is required")
		}
		customer.Name = *name
	}
	if surname != nil {
		customer.Surname = *surname
	}
	if email != nil {
		customer.Email = cleanStringPtr(email)
	}
	if phone != nil {
		parsed, err := ParsePhone(*phone)
		if err != nil {
			return Customer{}, err
		}
		customer.Phone = parsed
	}
	if note != nil {
		customer.Note = *note
	}
	return customer, nil
}

func cleanStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
