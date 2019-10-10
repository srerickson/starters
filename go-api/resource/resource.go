package resource

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/civil"
)

// Resource is the main database object
type Resource struct {
	ID        string     `json:"id,omitempty"`
	Label     string     `json:"label,omitempty"`
	Fields    *FieldSet  `json:"fields,omitempty"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// FieldSet holds Resource data
type FieldSet struct {
	Title         string      `json:"title,omitempty"`
	Creator       []Creator   `json:"creator,omitempty"`
	Description   string      `json:"description,omitempty"`
	DatePublished *civil.Date `json:"datePublished,omitempty"`
}

// Creator holds name of Resource creator/author
type Creator struct {
	Name string `json:"name,omitempty"`
}

const (
	index = iota
	get
	create
	update
	delete
)

var templates = [...]string{
	"SELECT id, label, created_at, updated_at FROM %s LIMIT $1 OFFSET $2",  // index
	`SELECT id, label, fields, created_at, updated_at FROM %s WHERE id=$1`, // get
	`INSERT INTO %s (label, fields, created_at) VALUES ($1, $2, current_timestamp) 
		RETURNING  id, created_at`, // create
	`UPDATE %s SET label=$1, fields=$2, updated_at=current_timestamp 
		WHERE id=$3 RETURNING updated_at`, // update
	"DELETE FROM %s WHERE id=$1", // delete
}

var queries [4]string

var db *sql.DB

// Init sets up state for querying db
func Init(_db *sql.DB, table string) {
	db = _db
	for i := range queries {
		queries[i] = fmt.Sprintf(templates[i], table)
	}
}

// Scan implement the sql.Scanner interface on Resource
func (fs *FieldSet) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &fs)
}

// Value implements the driver.Valuer interface on Resource
func (fs FieldSet) Value() (driver.Value, error) {
	return json.Marshal(fs)
}

// List returns a slice of Resources
func List(limit int, offset int) ([]Resource, error) {
	var items []Resource

	rows, err := db.Query(queries[index], limit, offset)
	if err != nil {
		return items, err
	}
	defer rows.Close()
	for rows.Next() {
		item := Resource{}
		//"SELECT id, label, created_at, updated_at FROM %s LIMIT $1 OFFSET $2",
		err = rows.Scan(&item.ID, &item.Label, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return items, err
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return items, err
	}
	return items, nil
}

// Get returns pointer to Resource with the given ID.
// If not such Resource exists the pointer resolve to an empty
// Resource (no error is returned).
func Get(id string) (*Resource, error) {
	r := Resource{}
	err := db.QueryRow(queries[get], id).Scan(&r.ID, &r.Label, &r.Fields, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Create saves the Resource to the database and
// returns the new ID
func (r *Resource) Create() error {
	if r.ID != `` {
		return fmt.Errorf(`cannot create resource with existing id: %s`, r.ID)
	}
	err := db.QueryRow(queries[create], r.Label, r.Fields).Scan(&r.ID, &r.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

// Update updates the Resource on the database
func (r *Resource) Update() error {
	if r.ID == `` {
		return errors.New(`cannot update Resource without existing id`)
	}
	err := db.QueryRow(queries[update], r.Label, r.Fields, r.ID).Scan(&r.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}
