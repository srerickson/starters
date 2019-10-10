package resource

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

const envDB = `DB_TEST_URL`

func TestMain(m *testing.M) {
	dbURL := os.Getenv(envDB)
	if dbURL == `` {
		log.Fatalf(`%s not set`, envDB)
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("database error: %s", err)
	}
	Init(db, `api.resources`)
	val := m.Run()
	db.Close()
	os.Exit(val)
}

func TestCreate(t *testing.T) {
	resc := Resource{
		Label: `test`,
		Fields: &FieldSet{
			Title: `This is a something`,
		},
	}
	err := resc.Create()
	if err != nil {
		t.Error(err)
	}
	label := `test`
	if resc.Label != label {
		t.Errorf(`expected label %s, got %s`, label, resc.Label)
	}
	if resc.ID == `` {
		t.Error(`expected Create to assign ID Resource`)
	}
	if resc.CreatedAt == nil {
		t.Error(`expected Create to set CreatedAt`)
	}
	if resc.UpdatedAt != nil {
		t.Error(`expected UpdatedAt to be nil`)
	}
}

func TestGet(t *testing.T) {
	resc := Resource{
		Label: `test`,
		Fields: &FieldSet{
			Title: `This is a something`,
		},
	}
	err := resc.Create()
	if err != nil {
		t.Error(err)
	}
	r, err := Get(resc.ID)
	if err != nil {
		t.Error(err)
	}
	if resc.ID != r.ID || r.ID == `` {
		t.Error(`expected Get to return Resource with matching ID`)
	}
}

func TestList(t *testing.T) {
	for i := 0; i < 50; i++ {
		r := Resource{
			Label: fmt.Sprintf(`test-%d`, i),
			Fields: &FieldSet{
				Title: `This is a something`,
			},
		}
		err := r.Create()
		if err != nil {
			t.Error(err)
			break
		}
	}
	rescs, err := List(50, 0)
	if err != nil {
		t.Error(err)
	}
	if len(rescs) != 50 {
		t.Errorf(`expected 50 rescs, got %d`, len(rescs))
	}
}

func TestUpdate(t *testing.T) {
	resc := Resource{
		Label: `test`,
		Fields: &FieldSet{
			Title: `This is a something`,
		},
	}
	err := resc.Create()
	if err != nil {
		t.Error(err)
	}

	// update
	resc.Label = `New Label`
	resc.Fields.Description = `Add a description`
	err = resc.Update()

	if err != nil {
		t.Error(err)
	}
	if resc.UpdatedAt == nil {
		t.Error(`expected Update to update UpdatedAt`)
	}

	resc2, err := Get(resc.ID)
	if err != nil {
		t.Error(err)
	}
	if resc2.Label != resc.Label {
		t.Error(`expected values to match`)
	}
	if resc2.Fields.Description != resc.Fields.Description {
		t.Error(`expected values to match`)
	}

}
