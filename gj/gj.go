package gj

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

type SerializerTemplate struct {
	fields []Field
}

type Serializer struct {
	forType  reflect.Type
	template *SerializerTemplate
	// some sort of map
	fieldmap  map[string]Field
	fieldmap2 map[string]reflect.Value
}

var ErrMemberFieldNotFound = errors.New("Member field not found on struct")
var ErrMemberFieldTypeMismatch = errors.New("Serializer field type mismatch Member field")
var ErrDuplicateField = errors.New("Duplicate field")

// Serializer builds a serializer out of this template
func (st *SerializerTemplate) Serializer(d interface{}) (*Serializer, error) {
	// Validate nested serializers
	// Everything will probably be recursive, e.g. Fields will validate as well

	s := &Serializer{template: st,
		fieldmap:  make(map[string]Field),
		fieldmap2: make(map[string]reflect.Value)}

	e := reflect.ValueOf(d).Elem()
	s.forType = reflect.TypeOf(d)
	// Iterate over the serializer fields and store them in a map
	for _, f := range st.fields {
		if _, exists := s.fieldmap[f.FromName()]; exists {
			return nil, ErrDuplicateField
		}
		s.fieldmap[f.FromName()] = f
		ef := e.FieldByName(f.FromName())
		if !ef.IsValid() {
			return nil, ErrMemberFieldNotFound
		}
		if !f.typeMatch(ef.Kind()) {
			return nil, ErrMemberFieldTypeMismatch
		}
		s.fieldmap2[f.FromName()] = ef
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
		r += "\"" + f.ToName() + "\":\"" + string(val) + "\","
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
		if v, found := targetMap[f.ToName()]; found {
			if err := f.setProp(target, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// NewSerializerTemplate creates a new SerializerTemplate based on the supplied (fields) config
func NewSerializerTemplate(fields ...Field) *SerializerTemplate {
	return &SerializerTemplate{
		fields: fields,
	}
}
