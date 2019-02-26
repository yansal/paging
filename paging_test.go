package paging

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)

type project struct {
	ID           int64
	DateCreation time.Time
}

var (
	now          = time.Now()
	inOneSecond  = now.Add(time.Second)
	inTwoSeconds = now.Add(2 * time.Second)
)

func setup(t *testing.T) *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateTable(project{}).Error; err != nil {
		t.Fatal(err)
	}
	for _, p := range []project{
		{ID: 1, DateCreation: now},
		{ID: 2, DateCreation: inOneSecond},
		{ID: 3, DateCreation: inTwoSeconds},
	} {
		if err := db.Create(p).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func TestOffset(t *testing.T) {
	db := setup(t)
	defer db.Close()

	var projects []project
	next, err := GORM(db,
		Page{
			Mode:    OffsetMode,
			DBField: "date_creation",
			Limit:   2,
		},
		&projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != 1 {
		t.Errorf("expected first id to be 1, got %d", projects[0].ID)
	}
	if projects[1].ID != 2 {
		t.Errorf("expected second id to be 2, got %d", projects[1].ID)
	}
	if !next.HasNext {
		t.Error("expected to have a next page")
	}

	projects = nil
	next, err = GORM(db, next, &projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != 3 {
		t.Errorf("expected first id to be 3, got %d", projects[0].ID)
	}
	if next.HasNext {
		t.Error("expected to not have a next page")
	}
	if next.Count != 3 {
		t.Errorf("expected count to be 3, got %d", next.Count)
	}
}

func TestOffsetReverse(t *testing.T) {
	db := setup(t)
	defer db.Close()

	var projects []project
	next, err := GORM(db,
		Page{
			Mode:    OffsetMode,
			DBField: "date_creation",
			Reverse: true,
			Limit:   2,
		},
		&projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != 3 {
		t.Errorf("expected first id to be 3, got %d", projects[0].ID)
	}
	if projects[1].ID != 2 {
		t.Errorf("expected second id to be 2, got %d", projects[1].ID)
	}
	if !next.HasNext {
		t.Error("expected to have a next page")
	}
	if next.Count != 3 {
		t.Errorf("expected count to be 3, got %d", next.Count)
	}

	projects = nil
	next, err = GORM(db, next, &projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != 1 {
		t.Errorf("expected first id to be 1, got %d", projects[0].ID)
	}
	if next.HasNext {
		t.Error("expected to not have a next page")
	}
}

func TestCursor(t *testing.T) {
	db := setup(t)
	defer db.Close()

	var projects []project
	next, err := GORM(db,
		Page{
			Mode:    CursorMode,
			DBField: "id",
			Cursor: Cursor{
				StructField: "ID",
			},
			Limit: 2,
		},
		&projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != 1 {
		t.Errorf("expected first id to be 1, got %d", projects[0].ID)
	}
	if projects[1].ID != 2 {
		t.Errorf("expected second id to be 2, got %d", projects[1].ID)

	}
	if !next.HasNext {
		t.Error("expected to have a next page")
	}

	projects = nil
	next, err = GORM(db, next, &projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != 3 {
		t.Errorf("expected first id to be 3, got %d", projects[0].ID)
	}
	if next.HasNext {
		t.Error("expected to not have a next page")
	}
}
func TestCursorReverse(t *testing.T) {
	db := setup(t)
	defer db.Close()

	var projects []project
	next, err := GORM(db,
		Page{
			Mode:    CursorMode,
			DBField: "id",
			Cursor: Cursor{
				StructField: "ID",
			},
			Reverse: true,
			Limit:   2,
		},
		&projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != 3 {
		t.Errorf("expected first id to be 3, got %d", projects[0].ID)
	}
	if projects[1].ID != 2 {
		t.Errorf("expected second id to be 2, got %d", projects[1].ID)
	}
	if !next.HasNext {
		t.Error("expected to have a next page")
	}

	projects = nil
	next, err = GORM(db, next, &projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != 1 {
		t.Errorf("expected first id to be 1, got %d", projects[0].ID)
	}
	if next.HasNext {
		t.Error("expected to not have a next page")
	}
}

func TestCursorDateCreation(t *testing.T) {
	db := setup(t)
	defer db.Close()

	var projects []project
	next, err := GORM(db,
		Page{
			Mode:    CursorMode,
			DBField: "date_creation",
			Cursor: Cursor{
				StructField: "DateCreation",
			},
			Limit: 2,
		},
		&projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != 1 {
		t.Errorf("expected first id to be 1, got %d", projects[0].ID)
	}
	if projects[1].ID != 2 {
		t.Errorf("expected second id to be 2, got %d", projects[1].ID)
	}
	if !next.HasNext {
		t.Error("expected to have a next page")
	}

	projects = nil
	next, err = GORM(db, next, &projects)
	if err != nil {
		t.Fatal(err)
	}
	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
	if projects[0].ID != 3 {
		t.Errorf("expected first id to be 3, got %d", projects[0].ID)
	}
	if next.HasNext {
		t.Error("expected to not have a next page")
	}
}
