// Filename: internal/data/schools.go

package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"appletree.joanneyong.net/internal/validator"
	"github.com/lib/pq"
)

type School struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"-"`
	Name      string    `json:"name"`
	Level     string    `json:"level"`
	Contact   string    `json:"contact"`
	Phone     string    `json:"phone"`
	Email     string    `json:"email,omitempty"`
	Website   string    `json:"website,omitempty"`
	Address   string    `json:"address"`
	Mode      []string  `json:"mode"`
	Version   int32     `json:"version"`
}

func ValidateSchool(v *validator.Validator, school *School) {
	// Use the Check() method to execute our validation checks
	v.Check(school.Name != "", "name", "must be provided")
	v.Check(len(school.Name) <= 200, "name", "must not be more than 200 bytes long")

	v.Check(school.Level != "", "level", "must be provided")
	v.Check(len(school.Level) <= 200, "level", "must not be more than 200 bytes long")

	v.Check(school.Contact != "", "contact", "must be provided")
	v.Check(len(school.Contact) <= 200, "contact", "must not be more than 200 bytes long")

	v.Check(school.Phone != "", "phone", "must be provided")
	v.Check(validator.Matches(school.Phone, validator.PhoneRX), "phone", "must be a valid phone number")

	v.Check(school.Email != "", "email", "must be provided")
	v.Check(validator.Matches(school.Email, validator.EmailRX), "email", "must be a valid email address")

	v.Check(school.Website != "", "website", "must be provided")
	v.Check(validator.ValidWebsite(school.Website), "website", "must be a valid URL")

	v.Check(school.Address != "", "address", "must be provided")
	v.Check(len(school.Address) <= 500, "address", "must not be more than 500 bytes long")

	v.Check(school.Mode != nil, "mode", "must be provided")
	v.Check(len(school.Mode) >= 1, "mode", "must contain at least 1 entry")
	v.Check(len(school.Mode) <= 5, "mode", "must contain at most 5 entries")
	v.Check(validator.Unique(school.Mode), "mode", "must not contain duplicate entries")
}

// Define a SchoolModel which wraps a sql.DB connection pool
type SchoolModel struct {
	DB *sql.DB
}

// Insert() allows us  to create a new School
func (m SchoolModel) Insert(school *School) error {
	query := `
		INSERT INTO schools (name, level, contact, phone, email, website, address, mode)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, version
	`
	// Collect the data fields into a slice
	args := []interface{}{
		school.Name, school.Level,
		school.Contact, school.Phone,
		school.Email, school.Website,
		school.Address, pq.Array(school.Mode),
	}
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&school.ID, &school.CreatedAt, &school.Version)
}

// Get() allows us to retrieve a specific School
func (m SchoolModel) Get(id int64) (*School, error) {
	// Ensure that there is a valid id
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Create the query
	query := `
		SELECT id, created_at, name, level, contact, phone, email, website, address, mode, version
		FROM schools
		WHERE id = $1
	`
	// Declare a School variable to hold the returned data
	var school School
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query using QueryRow()
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&school.ID,
		&school.CreatedAt,
		&school.Name,
		&school.Level,
		&school.Contact,
		&school.Phone,
		&school.Email,
		&school.Website,
		&school.Address,
		pq.Array(&school.Mode),
		&school.Version,
	)
	// Handle any errors
	if err != nil {
		// Check the type of error
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Success
	return &school, nil
}

// Update() allows us to edit/alter a specific School
// Optimistic locking (version number)
func (m SchoolModel) Update(school *School) error {
	// Create a query
	query := `
		UPDATE schools
		SET name = $1, level = $2, contact = $3,
		    phone = $4, email = $5, website = $6,
			address = $7, mode = $8, version = version + 1
		WHERE id = $9
		AND version = $10
		RETURNING version
	`
	args := []interface{}{
		school.Name,
		school.Level,
		school.Contact,
		school.Phone,
		school.Email,
		school.Website,
		school.Address,
		pq.Array(school.Mode),
		school.ID,
		school.Version,
	}
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Check for edit conflicts
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&school.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

// Delete() removes a specific School
func (m SchoolModel) Delete(id int64) error {
	// Ensure that there is a valid id
	if id < 1 {
		return ErrRecordNotFound
	}
	// Create the delete query
	query := `
		DELETE FROM schools
		WHERE id = $1
	`
	// Create a context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// Cleanup to prevent memory leaks
	defer cancel()
	// Execute the query
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	// Check how many rows were affected by the delete operation. We
	// call the RowsAffected() method on the result variable
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	// Check if no rows were affected
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// The GetAll() method retuns a list of all the schools sorted by id
func (m SchoolModel) GetAll(name string, level string, mode []string, filters Filters) ([]*School, error) {
	// Construct the query
	query := `
		SELECT id, created_at, name, level,
		       contact, phone, email, website,
			   address, mode, version
		FROM schools
		WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
		AND (to_tsvector('simple', level) @@ plainto_tsquery('simple', $2) OR $2 = '')
		AND (mode @> $3 OR $3 = '{}' )
		ORDER BY id
	`
	// Create a 3-second-timout context
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Execute the query
	rows, err := m.DB.QueryContext(ctx, query, name, level, pq.Array(mode))
	if err != nil {
		return nil, err
	}
	// Close the resultset
	defer rows.Close()
	// Initialize an empty slice to hold the School data
	schools := []*School{}
	// Iterate over the rows in the resultset
	for rows.Next() {
		var school School
		// Scan the values from the row into school
		err := rows.Scan(
			&school.ID,
			&school.CreatedAt,
			&school.Name,
			&school.Level,
			&school.Contact,
			&school.Phone,
			&school.Email,
			&school.Website,
			&school.Address,
			pq.Array(&school.Mode),
			&school.Version,
		)
		if err != nil {
			return nil, err
		}
		// Add the School to our slice
		schools = append(schools, &school)
	}
	// Check for errors after looping through the resultset
	if err = rows.Err(); err != nil {
		return nil, err
	}
	// Return the slice of Schools
	return schools, nil
}
