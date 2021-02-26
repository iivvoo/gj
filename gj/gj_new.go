package gj

import (
	"fmt"
	"reflect"
)

// Why Encode to interface, Decode from interface?
// Because it's recursive, and we decode from/encode to map[string]interface{}

type NSerializer interface {
	Encode(interface{}) (interface{}, error)
	Decode(target interface{}, val interface{}) error
}

type memberField struct {
	fromName   string
	toName     string
	serializer NSerializer
}

// MemberField??
func MemberField(f, t string, s NSerializer) *memberField {
	return &memberField{f, t, s}
}

type nstructField struct {
	fields []*memberField
}

func NStructField(f ...*memberField) NSerializer {
	return &nstructField{f}
}

func (st *nstructField) Encode(interface{}) (interface{}, error) {
	return nil, nil
}
func (st *nstructField) Decode(interface{}, interface{}) error {
	return nil
}

// A template adds validation, encoding to Seralizer

type template struct {
	s NSerializer
}

func Template(s NSerializer) *template {
	return &template{s}
}

func (t *template) Serialize(interface{}) ([]byte, error) {
	return nil, nil
}
func (t *template) Deserialize(interface{}, []byte) error {
	return nil
}

// ***** FIELDS ******

type nstringField struct{}

func NStringField() *nstringField {
	return &nstringField{}
}

func (sf *nstringField) Encode(value interface{}) (interface{}, error) {
	vv, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("Could not convert to string: %v", value)
	}

	return vv, nil
}
func (sf *nstringField) Decode(target interface{}, val interface{}) error {
	// ps := reflect.ValueOf(target)
	// s := ps.Elem()

	// Must be a Ptr
	t := reflect.ValueOf(target)
	if t.Kind() != reflect.Ptr {
		return ErrFieldUnsettable
	}
	if t.IsNil() {
		return ErrFieldUnsettable
	}
	v := t.Elem()
	if !v.CanSet() {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldUnsettable
	}

	vv, ok := val.(string)
	if !ok {
		return ErrFieldDataIncorrectType
	}
	if v.Kind() != reflect.String {
		// unlikely if we properly validate when creating the serializer
		return ErrFieldIncorrectType
	}
	v.SetString(vv)

	return nil
}

// NumberField? Does that include float? Negative?
type nintField struct{}

func NIntField() *nintField {
	return &nintField{}
}

func (inf *nintField) Encode(value interface{}) (interface{}, error) {
	vv, ok := value.(int)
	if !ok {
		return nil, fmt.Errorf("Could not convert to int: %v", value)
	}
	return vv, nil
}

func (inf *nintField) Decode(interface{}, interface{}) error {
	return nil
}
