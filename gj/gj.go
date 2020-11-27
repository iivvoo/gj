package gj

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

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

var x = errors.New("la")

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

type SerializerTemplate struct {
	fields []*Field
}

type Serializer struct {
	forType  reflect.Type
	template *SerializerTemplate
	// some sort of map
	fieldmap  map[string]*Field
	fieldmap2 map[string]reflect.Value
}

var ErrMemberFieldNotFound = errors.New("Member field not found on struct")
var ErrMemberFieldTypeMismatch = errors.New("Serializer field type mismatch Member field")
var ErrDuplicateField = errors.New("Duplicate field")

// Serializer builds a serializer out of this template
func (st *SerializerTemplate) Serializer(d interface{}) (*Serializer, error) {
	// Validate if fields exist
	// Validate if types are compatible
	// Validate nested serializers
	// Everything will probably be recursive, e.g. Fields will validate as well

	s := &Serializer{template: st,
		fieldmap:  make(map[string]*Field),
		fieldmap2: make(map[string]reflect.Value)}

	e := reflect.ValueOf(d).Elem()
	s.forType = reflect.TypeOf(d)
	// Iterate over the serializer fields and store them in a map
	for _, f := range st.fields {
		if _, exists := s.fieldmap[f.f]; exists {
			return nil, ErrDuplicateField
		}
		s.fieldmap[f.f] = f
		ef := e.FieldByName(f.f)
		if !ef.IsValid() {
			return nil, ErrMemberFieldNotFound
		}
		if !f.typeMatch(ef.Kind()) {
			return nil, ErrMemberFieldTypeMismatch
		}
		s.fieldmap2[f.f] = ef
	}

	return s, nil
}

func (s *Serializer) Encode(d interface{}) ([]byte, error) {
	// Assume always struct for now
	// probbaly check if d is of same type we validated for
	r := "{"

	e := reflect.ValueOf(d).Elem()
	// typeOfE := e.Type()

	for name, f := range s.fieldmap {
		_, found := s.fieldmap2[name]
		if !found {
			panic("Encode not found " + name)
		}
		ff := e.FieldByName(name)
		val, err := f.Value(ff.Interface())
		if err != nil {
			panic(err)
		}
		r += "\"" + f.t + "\":\"" + string(val) + "\","
	}
	r += "}"
	return []byte(r), nil
}

// ErrDifferentType is returned if the target does not match the serializer type
var ErrDifferentType = errors.New("target is not of same type")

// ErrArrayNotSupported is returned if an attempt is made to deserialze a non-object json structure, e.g. `[1,2]`
var ErrArrayNotSupported = errors.New("(de)serialization of pure json arrays not supported")

// Decode decodes `raw` into `target` which must be the same type as where the serialized
// was created for
func (s *Serializer) Decode(raw []byte, target interface{}) error {
	// we could probably create an instance of the type ourselves but,
	// - you may want to (partially) deserialize into an existing object
	// - you don't want to type assert the generic interface{} return value
	var any interface{}
	if err := json.Unmarshal(raw, &any); err != nil {
		return err
	}
	targetMap, ok := any.(map[string]interface{})
	if !ok {
		return ErrArrayNotSupported
	}

	targetType := reflect.TypeOf(target)

	fmt.Printf("TYPE %T %v -- %T %v\n", s.forType, s.forType, targetType, targetType)
	if targetType != s.forType {
		return ErrDifferentType
	}
	// Start decoding. Look specifically at the fields in the serializer in stead of everything
	// in the returned json.

	for _, f := range s.fieldmap {
		if v, found := targetMap[f.t]; found {
			if err := f.setProp(target, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewSerializerTemplate creates a new SerializerTemplate based on the supplied (fields) config
func NewSerializerTemplate(fields ...*Field) *SerializerTemplate {
	return &SerializerTemplate{
		fields: fields,
	}
}
