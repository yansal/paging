package paging

import (
	"fmt"
	"reflect"
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

type Store interface {
	Where(query string, args ...interface{})
	Order(order string)
	Limit(limit int64)
	Offset(offset int64)

	Find(dest interface{}) error
	Count(count *int64) error
}

func Paginate(store Store, page Page, dest interface{}) (next Page, err error) {
	switch page.Mode {
	case OffsetMode:
		return paginateOffset(store, page, dest)
	case CursorMode:
		return paginateCursor(store, page, dest)
	}
	return page, fmt.Errorf("unknown mode %d", page.Mode)
}

func paginateOffset(store Store, page Page, dest interface{}) (Page, error) {
	order := page.DBField
	if page.Reverse {
		order = order + " desc"
	}

	store.Order(order)
	store.Limit(page.Limit)
	store.Offset(page.Offset)
	if err := store.Find(dest); err != nil {
		return page, err
	}

	next := page
	next.Offset += next.Limit
	if int64(getLen(dest)) == page.Limit {
		next.HasNext = true
	} else {
		next.HasNext = false
	}

	store.Limit(-1)
	store.Offset(-1)
	if err := store.Count(&next.Count); err != nil {
		return next, err
	}

	return next, nil
}

func paginateCursor(store Store, page Page, dest interface{}) (Page, error) {
	if page.Cursor.Value != nil {
		op := ">"
		if page.Reverse {
			op = "<"
		}
		store.Where(page.DBField+" "+op+" ?", page.Cursor.Value)
	}

	order := page.DBField
	if page.Reverse {
		order = order + " desc"
	}

	store.Order(order)
	store.Limit(page.Limit + 1) // + 1 to compute HasNext
	if err := store.Find(dest); err != nil {
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
