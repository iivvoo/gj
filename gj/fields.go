package gj

import (
	"errors"
	"fmt"
	"reflect"
)

type Field interface {
	Encode(interface{}) (interface{}, error) // Rename to Encode
	SetMember(target interface{}, val interface{}) error
	typeMatch(k reflect.Kind) bool
	FromName() string
	ToName() string
}

// ErrFieldDataIncorrectType means the json data does not match the field
var ErrFieldDataIncorrectType = errors.New("Data is of incorrect type")

// ErrFieldDataOverflow means the json data would overflow the field
var ErrFieldDataOverflow = errors.New("Data would overflow field")

// ErrFieldUnsettable means the field is not settable
var ErrFieldUnsettable = errors.New("Field is not settable")

// ErrFieldIncorrectType means the field does not match the expcted type (impossible?)
var ErrFieldIncorrectType = errors.New("Field is of wrong type")

type BaseField struct {
	f string
	t string
}

func (b *BaseField) FromName() string {
	return b.f
}
func (b *BaseField) ToName() string {
	return b.t
}

type stringField struct {
	*BaseField
}
type numberField struct {
	*BaseField
}

func StringField(f, t string) *stringField {
	return &stringField{&BaseField{f, t}}
}

func (f *stringField) Encode(v interface{}) (interface{}, error) {
	vv, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("Could not convert to string: %v", v)
	}

	return vv, nil
}

func (f *stringField) SetMember(target interface{}, val interface{}) error {
	// move to BaseField?
	ps := reflect.ValueOf(target)
	s := ps.Elem()

	// So this is an addressable field, not a struct field
	structField := s.FieldByName(f.f)
	if !structField.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

	vv, ok := val.(string)
	if !ok {
		return ErrFieldDataIncorrectType
	}
	if structField.Kind() != reflect.String {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldIncorrectType
	}
	structField.SetString(vv)

	return nil
}

// Could be in BaseField, with value stored in struct data
func (f *stringField) typeMatch(k reflect.Kind) bool {
	return k == reflect.String
}

func NumberField(f, t string) *numberField {
	return &numberField{&BaseField{f, t}}
}

func (f numberField) Encode(v interface{}) (interface{}, error) {
	// Bwuh, need to deal with all types of numbers
	vv, ok := v.(int)
	if !ok {
		return nil, fmt.Errorf("Could not convert to int: %v", v)
	}
	return vv, nil
}

func (f *numberField) SetMember(target interface{}, val interface{}) error {
	ps := reflect.ValueOf(target)
	s := ps.Elem()

	// So this is an addressable field, not a struct field
	structField := s.FieldByName(f.f)
	if !structField.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

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

	return nil
}

// Could be in BaseField, with value stored in struct data
func (f *numberField) typeMatch(k reflect.Kind) bool {
	return k == reflect.Int
}
