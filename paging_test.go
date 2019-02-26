package paging

import (
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	paginggorm "github.com/yansal/paging/gorm"
)

type project struct {
	ID           int64
	DateCreation time.Time
}

func setup(t *testing.T) *gorm.DB {
	db, err := gorm.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}

	if err := db.CreateTable(project{}).Error; err != nil {
		t.Fatal(err)
	}

	now := time.Now()
	for _, p := range []project{
		{ID: 1, DateCreation: now},
		{ID: 2, DateCreation: now.Add(time.Second)},
		{ID: 3, DateCreation: now.Add(2 * time.Second)},
	} {
		if err := db.Create(p).Error; err != nil {
			t.Fatal(err)
		}
	}
	return db
}

func assertf(t *testing.T, ok bool, msg string, args ...interface{}) {
	t.Helper()
	if !ok {
		t.Errorf(msg, args...)
	}
}

func TestOffset(t *testing.T) {
	db := setup(t)
	defer db.Close()

	store := paginggorm.New(db)
	var projects []project
	next, err := Paginate(store, Page{
		DBField: "date_creation",
		Limit:   2,
	}, &projects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(projects) == 2, "expected 2 projects, got %d", len(projects))
	assertf(t, projects[0].ID == 1, "expected first id to be 1, got %d", projects[0].ID)
	assertf(t, projects[1].ID == 2, "expected second id to be 2, got %d", projects[1].ID)
	assertf(t, next.HasNext, "expected to have a next page")
	assertf(t, next.Count == 3, "expected count to be 3, got %d", next.Count)

	var nextProjects []project
	next, err = Paginate(store, next, &nextProjects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(nextProjects) == 1, "expected 1 project, got %d", len(nextProjects))
	assertf(t, nextProjects[0].ID == 3, "expected first id to be 3, got %d", nextProjects[0].ID)
	assertf(t, !next.HasNext, "expected to not have a next page")
	assertf(t, next.Count == 3, "expected count to be 3, got %d", next.Count)
}

func TestOffsetReverse(t *testing.T) {
	db := setup(t)
	defer db.Close()

	store := paginggorm.New(db)
	var projects []project
	next, err := Paginate(store, Page{
		DBField: "date_creation",
		Reverse: true,
		Limit:   2,
	}, &projects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(projects) == 2, "expected 2 projects, got %d", len(projects))
	assertf(t, projects[0].ID == 3, "expected first id to be 3, got %d", projects[0].ID)
	assertf(t, projects[1].ID == 2, "expected second id to be 2, got %d", projects[1].ID)
	assertf(t, next.HasNext, "expected to have a next page")
	assertf(t, next.Count == 3, "expected count to be 3, got %d", next.Count)

	var nextProjects []project
	next, err = Paginate(store, next, &nextProjects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(nextProjects) == 1, "expected 1 project, got %d", len(nextProjects))
	assertf(t, nextProjects[0].ID == 1, "expected first id to be 1, got %d", nextProjects[0].ID)
	assertf(t, !next.HasNext, "expected to not have a next page")
	assertf(t, next.Count == 3, "expected count to be 3, got %d", next.Count)
}

func TestCursor(t *testing.T) {
	db := setup(t)
	defer db.Close()

	store := paginggorm.New(db)
	var projects []project
	next, err := Paginate(store, Page{
		Mode:    CursorMode,
		DBField: "id",
		Cursor: Cursor{
			StructField: "ID",
		},
		Limit: 2,
	}, &projects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(projects) == 2, "expected 2 projects, got %d", len(projects))
	assertf(t, projects[0].ID == 1, "expected first id to be 1, got %d", projects[0].ID)
	assertf(t, projects[1].ID == 2, "expected second id to be 2, got %d", projects[1].ID)
	assertf(t, next.HasNext, "expected to have a next page")
	nextValue, ok := next.Cursor.Value.(int64)
	assertf(t, ok,
		"expected next value type to be int64, got %T", next.Cursor.Value)
	assertf(t, nextValue == 2,
		"expected next cursor value to be 2, got %v", nextValue)

	var nextProjects []project
	next, err = Paginate(store, next, &nextProjects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(nextProjects) == 1, "expected 1 project, got %d", len(nextProjects))
	assertf(t, nextProjects[0].ID == 3, "expected first id to be 3, got %d", nextProjects[0].ID)
	assertf(t, !next.HasNext, "expected to not have a next page")
}
func TestCursorReverse(t *testing.T) {
	db := setup(t)
	defer db.Close()

	store := paginggorm.New(db)
	var projects []project
	next, err := Paginate(store, Page{
		Mode:    CursorMode,
		DBField: "id",
		Cursor: Cursor{
			StructField: "ID",
		},
		Reverse: true,
		Limit:   2,
	}, &projects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(projects) == 2, "expected 2 projects, got %d", len(projects))
	assertf(t, projects[0].ID == 3, "expected first id to be 3, got %d", projects[0].ID)
	assertf(t, projects[1].ID == 2, "expected second id to be 2, got %d", projects[1].ID)
	assertf(t, next.HasNext, "expected to have a next page")
	nextID, ok := next.Cursor.Value.(int64)
	assertf(t, ok,
		"expected next value type to be int64, got %T", next.Cursor.Value)
	assertf(t, nextID == 2,
		"expected next cursor value to be 2, got %v", nextID)

	var nextProjects []project
	next, err = Paginate(store, next, &nextProjects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(nextProjects) == 1, "expected 1 project, got %d", len(nextProjects))
	assertf(t, nextProjects[0].ID == 1, "expected first id to be 1, got %d", nextProjects[0].ID)
	assertf(t, !next.HasNext, "expected to not have a next page")
}

func TestCursorDateCreation(t *testing.T) {
	db := setup(t)
	defer db.Close()

	store := paginggorm.New(db)
	var projects []project
	next, err := Paginate(store, Page{
		Mode:    CursorMode,
		DBField: "date_creation",
		Cursor: Cursor{
			StructField: "DateCreation",
		},
		Limit: 2,
	}, &projects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(projects) == 2, "expected 2 projects, got %d", len(projects))
	assertf(t, projects[0].ID == 1, "expected first id to be 1, got %d", projects[0].ID)
	assertf(t, projects[1].ID == 2, "expected second id to be 2, got %d", projects[1].ID)
	assertf(t, next.HasNext, "expected to have a next page")
	nextValue, ok := next.Cursor.Value.(time.Time)
	assertf(t, ok,
		"expected next value type to be time.Time, got %T", next.Cursor.Value)
	assertf(t, nextValue.Equal(projects[1].DateCreation),
		"expected next cursor value to be %v, got %v", projects[1].DateCreation, nextValue)

	var nextProjects []project
	next, err = Paginate(store, next, &nextProjects)
	if err != nil {
		t.Fatal(err)
	}
	assertf(t, len(nextProjects) == 1, "expected 1 project, got %d", len(nextProjects))
	assertf(t, nextProjects[0].ID == 3, "expected first id to be 3, got %d", nextProjects[0].ID)
	assertf(t, !next.HasNext, "expected to not have a next page")
}
