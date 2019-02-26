package paging

import (
	"fmt"
	"reflect"

	"github.com/jinzhu/gorm"
)

type Mode int

const (
	OffsetMode Mode = iota
	CursorMode
)

type Page struct {
	Mode    Mode
	Offset  int64
	Cursor  Cursor
	DBField string // used for WHERE in cursor mode and for ORDER BY in both mode
	Reverse bool
	Limit   int64

	Count   int64 // only valid in offset mode
	HasNext bool
}

type Cursor struct {
	Value       interface{}
	StructField string // used to compute next cursor value
}

func GORM(db *gorm.DB, page Page, dest interface{}) (next Page, err error) {
	switch page.Mode {
	case OffsetMode:
		return gormOffset(db, page, dest)
	case CursorMode:
		return gormCursor(db, page, dest)
	}
	return page, fmt.Errorf("unknown mode %d", page.Mode)
}

func gormOffset(db *gorm.DB, page Page, dest interface{}) (Page, error) {
	order := page.DBField
	if page.Reverse {
		order = order + " desc"
	}

	db = db.
		Order(order).
		Limit(page.Limit).
		Offset(page.Offset).
		Find(dest)
	if err := db.Error; err != nil {
		return page, err
	}

	next := page
	next.Offset += next.Limit
	if int64(getLen(dest)) == page.Limit {
		next.HasNext = true
	} else {
		next.HasNext = false
	}

	if err := db.Limit(-1).Offset(-1).
		Count(&next.Count).
		Error; err != nil {
		return next, err
	}

	return next, nil
}

func gormCursor(db *gorm.DB, page Page, dest interface{}) (Page, error) {
	if page.Cursor.Value != nil {
		op := ">"
		if page.Reverse {
			op = "<"
		}
		db = db.Where(page.DBField+" "+op+" ?", page.Cursor.Value)
	}

	order := page.DBField
	if page.Reverse {
		order = order + " desc"
	}

	if err := db.
		Order(order).
		Limit(page.Limit + 1). // + 1 to compute HasNext
		Find(dest).
		Error; err != nil {
		return page, err
	}

	next := page
	if int64(getLen(dest)) == page.Limit+1 {
		_, dest = popLastElement(dest)
		next.HasNext = true
	} else {
		next.HasNext = false
	}
	next.Cursor.Value = getLastElementField(dest, page.Cursor.StructField)
	return next, nil
}

// the three functions below are copied from https://github.com/ulule/paging
func getLastElementField(array interface{}, fieldname string) interface{} {
	value := reflect.ValueOf(array)
	kind := value.Kind()
	if kind == reflect.Ptr {
		value = value.Elem()
		kind = value.Kind()
	}

	if kind != reflect.Array && kind != reflect.Slice {
		panic(fmt.Sprintf("can't get last element of a value of type %T", array))
	}

	if value.Len() == 0 {
		return nil
	}

	last := value.Index(value.Len() - 1)
	if last.Kind() != reflect.Struct {
		panic(fmt.Sprintf("can't get fieldname %q of an element of type %T", fieldname, last.Interface()))
	}

	return last.FieldByName(fieldname).Interface()
}

func getLen(array interface{}) int {
	value := reflect.ValueOf(array)
	kind := value.Kind()
	if kind == reflect.Ptr {
		value = value.Elem()
		kind = value.Kind()
	}

	if kind != reflect.Array && kind != reflect.Slice {
		panic(fmt.Sprintf("can't get len of a value of type %T", array))
	}

	return value.Len()
}

func popLastElement(arrayPtr interface{}) (last, remaining interface{}) {
	ptr := reflect.ValueOf(arrayPtr)
	if ptr.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected pointer type, got %T", arrayPtr))
	}

	array := ptr.Elem()
	if array.Kind() != reflect.Array && array.Kind() != reflect.Slice {
		panic(fmt.Sprintf("can't pop last element of a value of type %T", arrayPtr))
	}

	len := array.Len()
	if len == 0 {
		return nil, arrayPtr
	}

	last = array.Index(len - 1).Interface()

	array.Set(array.Slice(0, len-1))
	remaining = array.Addr().Interface()

	return last, remaining
}
