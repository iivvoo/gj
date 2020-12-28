package gj

import (
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
}

func StructField(f, t string, s *Serializer) *structField {
	return &structField{&BaseField{f, t}, s}
}

func (f *structField) Encode(v interface{}) (interface{}, error) {
	return f.s.EncodeBase(v)
}
func (f *structField) Decode(target interface{}, val interface{}) error {

	// If the value is nil (so null in the json), we're done. We could also make this
	// optional and make the field decode null into an empty struct
	if val == nil {
		return nil
	}

	// The value is not nil, so there's data that needs to be decoded. If we need to decode
	// into a struct, we need to make sure there's a struct instantiated in target to hold the data

	// We need the Value to set the new value, and the Type to create a new instance, if necessary

	ps := reflect.ValueOf(target)
	s := ps.Elem() // Assumes pointer, which it must be anyway if we want to be able to change it

	fieldValue := s.FieldByName(f.f)
	if !fieldValue.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			fieldType, found := reflect.TypeOf(target).Elem().FieldByName(f.FromName())
			if !found { // should be unlikely at this point
				return ErrMemberFieldNotFound
			}
			fieldValue.Set(reflect.New(fieldType.Type.Elem()))
		}
	} else {
		// We'll need a pointer anyway
		fieldValue = fieldValue.Addr()
	}
	f.s.DecodeBase(val, fieldValue.Interface()) // swapped order, weird?

	return nil
}

func (f *structField) typeMatch(k reflect.Type) bool {
	// So it's a struct, but is it the expected type? E.g. main.FooStruct
	if k.Kind() == reflect.Ptr {
		k = k.Elem()
	}
	return k.Kind() == reflect.Struct
}
