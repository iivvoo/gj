package gj

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type xField interface {
}

type FieldType uint

const (
	StringFieldType FieldType = 1
	NumberFieldType FieldType = 2
)

type Field struct {
	ftype FieldType
	f     string
	t     string
}

func StringField(f, t string) *Field {
	return &Field{StringFieldType, f, t}
}

func NumberField(f, t string) *Field {
	return &Field{NumberFieldType, f, t}
}

func (f *Field) Value(v interface{}) ([]byte, error) {
	switch f.ftype {
	case StringFieldType:
		vv, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("Could not convert to string: %v", v)
		}
		return []byte(vv), nil
	case NumberFieldType:
		// Bwuh, need to deal with all types of numbers
		vv, ok := v.(int)
		if !ok {
			return nil, fmt.Errorf("Could not convert to int: %v", v)
		}
		return []byte(strconv.Itoa(vv)), nil
	default:
		return nil, fmt.Errorf("Unknown type: %d", f.ftype)
	}
}

// ErrFieldDataIncorrectType means the json data does not match the field
var ErrFieldDataIncorrectType = errors.New("Data is of incorrect type")

// ErrFieldDataOverflow means the json data would overflow the field
var ErrFieldDataOverflow = errors.New("Data would overflow field")

// ErrFieldUnsettable means the field is not settable
var ErrFieldUnsettable = errors.New("Field is not settable")

// ErrFieldIncorrectType means the field does not match the expcted type (impossible?)
var ErrFieldIncorrectType = errors.New("Field is of wrong type")

func (f *Field) setProp(target interface{}, val interface{}) error {
	ps := reflect.ValueOf(target)
	s := ps.Elem()

	// So this is an addressable field, not a struct field
	structField := s.FieldByName(f.f)
	if !structField.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

	switch f.ftype {
	case StringFieldType:
		vv, ok := val.(string)
		if !ok {
			return ErrFieldDataIncorrectType
		}
		if structField.Kind() != reflect.String {
			// unlikely if we properly validate when creating the serializer
			return ErrFieldIncorrectType
		}
		structField.SetString(vv)
	case NumberFieldType:
		vv, ok := val.(int64)
		if !ok {
			return ErrFieldDataIncorrectType
		}
		if structField.Kind() == reflect.Int {
			if structField.OverflowInt(vv) {
				return ErrFieldDataOverflow
			}
			structField.SetInt(vv)
		} else {
			return ErrFieldIncorrectType
		}
		return nil
	default:
		return fmt.Errorf("Unknown type: %d", f.ftype)
	}

	return nil
}
func (f *Field) typeMatch(k reflect.Kind) bool {
	switch f.ftype {
	case StringFieldType:
		return k == reflect.String
	case NumberFieldType:
		return k == reflect.Int
	}
	return false
}
