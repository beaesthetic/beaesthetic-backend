package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
	"github.com/petretiandrea/beaesthetic-backend/customer/internal/infra/postgres/queries"
)

type CustomerRepository struct {
	db      *sql.DB
	queries *queries.Queries
}

func NewCustomerRepository(db *sql.DB) *CustomerRepository {
	return &CustomerRepository{db: db, queries: queries.New(db)}
}

func (r *CustomerRepository) Save(ctx context.Context, c customerdomain.Customer) (customerdomain.Customer, error) {
	phone := (*string)(nil)
	if c.Phone != nil {
		value := c.Phone.FullNumber()
		phone = &value
	}
	return c, r.queries.SaveCustomer(ctx, queries.SaveCustomerParams{
		ID:        c.ID,
		Name:      c.Name,
		Surname:   c.Surname,
		Email:     nullableString(c.Email),
		Phone:     nullableString(phone),
		Note:      c.Note,
		UpdatedAt: time.Now().UTC(),
	})
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string) (*customerdomain.Customer, error) {
	row, err := r.queries.FindCustomerByID(ctx, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	customer := mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note)
	return &customer, nil
}

func (r *CustomerRepository) FindAll(ctx context.Context, filter string, limit int) ([]customerdomain.Customer, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if strings.TrimSpace(filter) == "" {
		rows, err := r.queries.FindCustomers(ctx, int32(limit))
		if err != nil {
			return nil, err
		}
		out := make([]customerdomain.Customer, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
		}
		return out, nil
	}
	rows, err := r.queries.SearchCustomers(ctx, queries.SearchCustomersParams{
		Column1: sql.NullString{String: strings.ToLower(filter), Valid: true},
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out, nil
}

func (r *CustomerRepository) FindByPhone(ctx context.Context, phone string) (*customerdomain.Customer, error) {
	row, err := r.queries.FindCustomerByPhone(ctx, sql.NullString{String: phone, Valid: true})
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	customer := mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note)
	return &customer, nil
}

func (r *CustomerRepository) FindPage(ctx context.Context, pageToken string, limit int, sortBy string, direction string) ([]customerdomain.Customer, string, bool, bool, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := decodePageToken(pageToken)
	items, err := r.findPage(ctx, limit, offset, sortBy, direction)
	if err != nil {
		return nil, "", false, false, err
	}
	next := ""
	if len(items) == limit {
		next = encodePageToken(offset + len(items))
	}
	return items, next, next != "", offset > 0, nil
}

func (r *CustomerRepository) Delete(ctx context.Context, id string) (bool, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)
	rowsAffected, err := qtx.ArchiveDeletedCustomer(ctx, id)
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}
	if err := qtx.DeleteCustomer(ctx, id); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *CustomerRepository) findPage(ctx context.Context, limit int, offset int, sortBy string, direction string) ([]customerdomain.Customer, error) {
	limit32 := int32(limit)
	offset32 := int32(offset)
	mapRows := func(rows []queries.FindCustomersPageByNameAscRow) []customerdomain.Customer {
		out := make([]customerdomain.Customer, 0, len(rows))
		for _, row := range rows {
			out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
		}
		return out
	}
	if direction == "prev" {
		switch sortBy {
		case "surname":
			rows, err := r.queries.FindCustomersPageBySurnameDesc(ctx, queries.FindCustomersPageBySurnameDescParams{Limit: limit32, Offset: offset32})
			return mapCustomerPageBySurnameDesc(rows), err
		case "updated_at":
			rows, err := r.queries.FindCustomersPageByUpdatedAtDesc(ctx, queries.FindCustomersPageByUpdatedAtDescParams{Limit: limit32, Offset: offset32})
			return mapCustomerPageByUpdatedAtDesc(rows), err
		default:
			rows, err := r.queries.FindCustomersPageByNameDesc(ctx, queries.FindCustomersPageByNameDescParams{Limit: limit32, Offset: offset32})
			return mapCustomerPageByNameDesc(rows), err
		}
	}
	switch sortBy {
	case "surname":
		rows, err := r.queries.FindCustomersPageBySurnameAsc(ctx, queries.FindCustomersPageBySurnameAscParams{Limit: limit32, Offset: offset32})
		return mapCustomerPageBySurnameAsc(rows), err
	case "updated_at":
		rows, err := r.queries.FindCustomersPageByUpdatedAtAsc(ctx, queries.FindCustomersPageByUpdatedAtAscParams{Limit: limit32, Offset: offset32})
		return mapCustomerPageByUpdatedAtAsc(rows), err
	default:
		rows, err := r.queries.FindCustomersPageByNameAsc(ctx, queries.FindCustomersPageByNameAscParams{Limit: limit32, Offset: offset32})
		return mapRows(rows), err
	}
}

func mapCustomer(id string, name string, surname string, email sql.NullString, phone sql.NullString, note string) customerdomain.Customer {
	c := customerdomain.Customer{ID: id, Name: name, Surname: surname, Note: note}
	if email.Valid {
		c.Email = &email.String
	}
	if phone.Valid {
		parsed, err := customerdomain.ParsePhone(phone.String)
		if err == nil {
			c.Phone = parsed
		}
	}
	return c
}

func mapCustomerPageByNameDesc(rows []queries.FindCustomersPageByNameDescRow) []customerdomain.Customer {
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out
}

func mapCustomerPageBySurnameAsc(rows []queries.FindCustomersPageBySurnameAscRow) []customerdomain.Customer {
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out
}

func mapCustomerPageBySurnameDesc(rows []queries.FindCustomersPageBySurnameDescRow) []customerdomain.Customer {
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out
}

func mapCustomerPageByUpdatedAtAsc(rows []queries.FindCustomersPageByUpdatedAtAscRow) []customerdomain.Customer {
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out
}

func mapCustomerPageByUpdatedAtDesc(rows []queries.FindCustomersPageByUpdatedAtDescRow) []customerdomain.Customer {
	out := make([]customerdomain.Customer, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapCustomer(row.ID, row.Name, row.Surname, row.Email, row.Phone, row.Note))
	}
	return out
}

func nullableString(value *string) sql.NullString {
	if value == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *value, Valid: true}
}

func encodePageToken(offset int) string {
	return base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(offset)))
}
func decodePageToken(token string) int {
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return 0
	}
	v, err := strconv.Atoi(string(b))
	if err != nil {
		return 0
	}
	return v
}
