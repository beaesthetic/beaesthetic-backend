package postgres

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	customerdomain "github.com/petretiandrea/beaesthetic-backend/customer/internal/domain/customer"
)

type CustomerRepository struct{ db *sql.DB }

func NewCustomerRepository(db *sql.DB) *CustomerRepository { return &CustomerRepository{db: db} }

func (r *CustomerRepository) Save(ctx context.Context, c customerdomain.Customer) (customerdomain.Customer, error) {
	phone := (*string)(nil)
	if c.Phone != nil {
		value := c.Phone.FullNumber()
		phone = &value
	}
	_, err := r.db.ExecContext(ctx, `INSERT INTO customers (id,name,surname,email,phone,note,updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (id) DO UPDATE SET name=$2,surname=$3,email=$4,phone=$5,note=$6,updated_at=$7`, c.ID, c.Name, c.Surname, c.Email, phone, c.Note, time.Now().UTC())
	return c, err
}

func (r *CustomerRepository) FindByID(ctx context.Context, id string) (*customerdomain.Customer, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,surname,email,phone,note FROM customers WHERE id=$1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	customers, err := scanCustomers(rows)
	if err != nil || len(customers) == 0 {
		return nil, err
	}
	return &customers[0], nil
}

func (r *CustomerRepository) FindAll(ctx context.Context, filter string, limit int) ([]customerdomain.Customer, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if strings.TrimSpace(filter) == "" {
		rows, err := r.db.QueryContext(ctx, `SELECT id,name,surname,email,phone,note FROM customers ORDER BY name,surname LIMIT $1`, limit)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		return scanCustomers(rows)
	}
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,surname,email,phone,note FROM customers WHERE search_text ILIKE '%' || $1 || '%' OR search_text % $1 ORDER BY similarity(search_text,$1) DESC,name,surname LIMIT $2`, strings.ToLower(filter), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanCustomers(rows)
}

func (r *CustomerRepository) FindByPhone(ctx context.Context, phone string) (*customerdomain.Customer, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,name,surname,email,phone,note FROM customers WHERE phone=$1 LIMIT 1`, phone)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	customers, err := scanCustomers(rows)
	if err != nil || len(customers) == 0 {
		return nil, err
	}
	return &customers[0], nil
}

func (r *CustomerRepository) FindPage(ctx context.Context, pageToken string, limit int, sortBy string, direction string) ([]customerdomain.Customer, string, bool, bool, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := decodePageToken(pageToken)
	orderBy := "name"
	if sortBy == "surname" || sortBy == "updated_at" {
		orderBy = sortBy
	}
	dir := "ASC"
	if direction == "prev" {
		dir = "DESC"
	}
	query := fmt.Sprintf(`SELECT id,name,surname,email,phone,note FROM customers ORDER BY %s %s,id LIMIT $1 OFFSET $2`, orderBy, dir)
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, "", false, false, err
	}
	defer rows.Close()
	items, err := scanCustomers(rows)
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
	cmd, err := tx.ExecContext(ctx, `INSERT INTO deleted_customers (id,name,surname,email,phone,note)
SELECT id,name,surname,email,phone,note FROM customers WHERE id=$1 ON CONFLICT (id) DO NOTHING`, id)
	if err != nil {
		return false, err
	}
	rowsAffected, err := cmd.RowsAffected()
	if err != nil {
		return false, err
	}
	if rowsAffected == 0 {
		return false, nil
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM customers WHERE id=$1`, id); err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func scanCustomers(rows *sql.Rows) ([]customerdomain.Customer, error) {
	out := []customerdomain.Customer{}
	for rows.Next() {
		var c customerdomain.Customer
		var email, phone *string
		if err := rows.Scan(&c.ID, &c.Name, &c.Surname, &email, &phone, &c.Note); err != nil {
			return nil, err
		}
		c.Email = email
		if phone != nil {
			parsed, err := customerdomain.ParsePhone(*phone)
			if err == nil {
				c.Phone = parsed
			}
		}
		out = append(out, c)
	}
	return out, rows.Err()
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
