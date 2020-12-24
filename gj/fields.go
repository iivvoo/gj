package gj

import (
	"errors"
	"fmt"
	"reflect"
)

type Field interface {
	Encode(interface{}) (interface{}, error)
	Decode(target interface{}, val interface{}) error // Decode
	typeMatch(k reflect.Type) bool
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

func (f *stringField) Decode(target interface{}, val interface{}) error {
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
func (f *stringField) typeMatch(k reflect.Type) bool {
	return k.Kind() == reflect.String
}

type numberField struct {
	*BaseField
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

func (f *numberField) Decode(target interface{}, val interface{}) error {
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
}

// Could be in BaseField, with value stored in struct data
func (f *numberField) typeMatch(k reflect.Type) bool {
	return k.Kind() == reflect.Int
}

type structField struct {
	*BaseField
	s *Serializer
	t reflect.Type
}

func StructField(f, t string, s *Serializer) *structField {
	return &structField{&BaseField{f, t}, s, nil}
}

func (f *structField) Encode(v interface{}) (interface{}, error) {
	return f.s.EncodeBase(v)
}
func (f *structField) Decode(target interface{}, val interface{}) error {

	ps := reflect.ValueOf(target)
	s := ps.Elem() // Assumes pointer

	structField := s.FieldByName(f.f)
	if !structField.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

	if structField.Kind() == reflect.Ptr {
		if structField.IsNil() {
			structField.Set(reflect.New(f.t))
			// We will want to set this value on s
		}
	} else {
		// We'll need a pointer anyway
		structField = structField.Addr()
	}
	f.s.DecodeBase(val, structField.Interface()) // swapped order, weird?

	return nil
}

func (f *structField) typeMatch(k reflect.Type) bool {
	// So it's a struct, but is it the expected type? E.g. main.FooStruct
	// could even be a while?
	if k.Kind() == reflect.Ptr {
		fmt.Println("Elem")
		k = k.Elem()
	}
	fmt.Printf("%T %v %v\n", k, k, k.Kind())
	f.t = k
	return k.Kind() == reflect.Struct
}
