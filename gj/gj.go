package gj

// la

import (
	"encoding/json"
	"reflect"
)

type SerializerTemplate struct {
	fields []Field
}

// NewSerializerTemplate creates a new SerializerTemplate based on the supplied (fields) config
func NewSerializerTemplate(fields ...Field) (*SerializerTemplate, error) {
	st := &SerializerTemplate{}
	for _, f := range fields {
		if err := st.Add(f); err != nil {
			return nil, err
		}
	}
	return st, nil
}

// Add adds a field to the Template. It checks if the field is not a duplicate
func (st *SerializerTemplate) Add(newField Field) error {
	for _, f := range st.fields {
		if f.FromName() == newField.FromName() {
			// would be nice to add the duplicate field key to the error
			return ErrDuplicateField
		}
	}
	st.fields = append(st.fields, newField)
	return nil
}

// Serializer builds a serializer out of this template
func (st *SerializerTemplate) Serializer(d interface{}) (*Serializer, error) {
	// Validate nested serializers

	// Elem() assumes it's a pointer.
	e := reflect.TypeOf(d).Elem()
	s := &Serializer{forType: reflect.TypeOf(d), template: st}

	// Check if this template is suitable for the type `d` passed, which means the fields must
	// exist and be type-compatible
	for _, f := range st.fields {
		if ef, found := e.FieldByName(f.FromName()); !found {
			// it would be nice to include which field wasn't found
			return nil, ErrMemberFieldNotFound
		} else if !f.typeMatch(ef.Type) {
			return nil, ErrMemberFieldTypeMismatch
		}
	}

	return s, nil
}

type Serializer struct {
	forType  reflect.Type
	template *SerializerTemplate
}

func (s *Serializer) EncodeBase(d interface{}) (interface{}, error) {
	targetType := reflect.TypeOf(d)

	if targetType != s.forType {
		return nil, ErrDifferentType
	}

	e := reflect.ValueOf(d).Elem()

	collector := make(map[string]interface{})

	for _, f := range s.template.fields {
		ff := e.FieldByName(f.FromName())
		// Do not recurse into pointers if they're nil
		if ff.Kind() != reflect.Ptr || !ff.IsNil() {
			val, err := f.Encode(ff.Interface())
			if err != nil {
				return nil, err
			}
			collector[f.ToName()] = val
		} else {
			collector[f.ToName()] = nil
		}
	}
	return collector, nil
}

func (s *Serializer) Encode(d interface{}) ([]byte, error) {
	// Assume always struct for now
	// probbaly check if d is of same type we validated for

	raw, err := s.EncodeBase(d)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(raw)
	return encoded, err
}

func (s *Serializer) DecodeBase(val interface{}, target interface{}) error {

	targetMap, ok := val.(map[string]interface{})
	if !ok {
		return ErrArrayNotSupported
	}

	// XXX Deal with PTR?
	targetType := reflect.TypeOf(target)

	if targetType != s.forType {
		return ErrDifferentType
	}
	// Start decoding. Look specifically at the fields in the serializer in stead of everything
	// in the returned json.

	for _, f := range s.template.fields {
		// targetMap is effectively what json.Unmarshal produced as a map[string]interface{}
		if v, found := targetMap[f.ToName()]; found {
			// So v is the actual value decoded
			if err := f.Decode(target, v); err != nil {
				return err
			}
		}
	}

	return nil
}

// Decode decodes `raw` into `target` which must be the same type as where the serialized
// was created for
func (s *Serializer) Decode(raw []byte, target interface{}) error {
	// we could probably create an instance of the type ourselves but,
	// - you may want to (partially) deserialize into an existing object
	// - you don't want to type assert the generic interface{} return value

	// passing anything else than a pointer does not make sense
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		return ErrNotAPointer
	}
	if reflect.ValueOf(target).IsNil() {
		return ErrNilPointer
	}
	var any interface{}
	if err := json.Unmarshal(raw, &any); err != nil {
		return err
	}
	return s.DecodeBase(any, target)
}
