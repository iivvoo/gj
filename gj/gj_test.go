package gj

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// some randomly complex structure that we want to mess with
type NestedStruct struct {
	Name    string
	Numbers []int
	When    time.Time
}
type SampleStruct struct {
	Name string
	Num  int
	Nest *NestedStruct
}

func TestSerializerCreation(t *testing.T) {
	t.Run("Test simple success", func(t *testing.T) {
		assert := assert.New(t)

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		_, err = tpl.Serializer(&struct{ A string }{})
		assert.NoError(err)
	})
	t.Run("Test duplicate field", func(t *testing.T) {
		assert := assert.New(t)

		_, err := NewSerializerTemplate(StringField("A", "a"), StringField("A", "b"))
		assert.Error(err)
		assert.EqualValues(ErrDuplicateField, err)
	})
	t.Run("Test missing member", func(t *testing.T) {
		assert := assert.New(t)

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		_, err = tpl.Serializer(&struct{}{})
		assert.Error(err)
		assert.EqualValues(ErrMemberFieldNotFound, err)
	})
	t.Run("Test type mismatch", func(t *testing.T) {
		assert := assert.New(t)

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		_, err = tpl.Serializer(&struct{ A int }{})
		assert.Error(err)
		assert.EqualValues(ErrMemberFieldTypeMismatch, err)
	})
	// Test that fields can be added
	// Test that duplication cannot be added
}

func TestSerialization(t *testing.T) {
	t.Run("Test simple success", func(t *testing.T) {
		assert := assert.New(t)

		type S struct {
			A string
			B int
		}
		tpl, err := NewSerializerTemplate(StringField("A", "a"), NumberField("B", "x"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&S{})
		assert.NoError(err)

		res, err := serializer.Encode(&S{"Hello", 42})
		assert.NoError(err)

		// This randomly fails because map order is random. In stead, create assert that matches json?
		assert.JSONEq(`{"a":"Hello","x":42}`, string(res))
	})

	t.Run("Test nesting", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
		}
		type S struct {
			A string
			Q *Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)
		assert.NotNil(qSerializer)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)
		assert.NotNil(sSerializer)

		res, err := sSerializer.Encode(&S{"Hello", &Q{"World"}})
		assert.NoError(err)

		assert.Equal(`{"a":"Hello","q":{"aa":"World"}}`, string(res))
	})
	t.Run("Test nesting, nil", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
		}
		type S struct {
			A string
			Q *Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)
		assert.NotNil(qSerializer)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)
		assert.NotNil(sSerializer)

		res, err := sSerializer.Encode(&S{"Hello", nil})
		assert.NoError(err)

		assert.Equal(`{"a":"Hello","q":null}`, string(res))
	})
}
func TestSerializerTypeMatch(t *testing.T) {
	type A struct{ A string }
	type B struct{ A string }

	t.Run("Test type match", func(t *testing.T) {
		assert := assert.New(t)

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)
		a := A{}
		err = serializer.Decode([]byte(`{"a":"x"}`), &a)
		assert.NoError(err)
	})
	t.Run("Test type mismatch", func(t *testing.T) {
		assert := assert.New(t)

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)
		b := B{}
		err = serializer.Decode([]byte(`{"a":"x"}`), &b)
		assert.Error(err)
		assert.EqualValues(ErrDifferentType, err)
	})
}

func TestSerializerDeserialize(t *testing.T) {

	t.Run("Test trivial case", func(t *testing.T) {
		assert := assert.New(t)
		type A struct{ A string }

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)

		a := A{}

		err = serializer.Decode([]byte(`{"a":"A"}`), &a)
		assert.NoError(err)
		assert.Equal("A", a.A)
	})

	t.Run("Test deserialize into non-pointer", func(t *testing.T) {
		assert := assert.New(t)
		type A struct{ A string }

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)

		err = serializer.Decode([]byte(`{"a":"A"}`), A{})
		assert.Error(err)
		assert.Equal(err, ErrNotAPointer)
	})
	t.Run("Test deserialize into nil", func(t *testing.T) {
		assert := assert.New(t)
		type A struct{ A string }

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)

		var a *A

		err = serializer.Decode([]byte(`{"a":"A"}`), a)
		assert.Error(err)
		assert.Equal(err, ErrNilPointer)
	})

	t.Run("Test wrong data type", func(t *testing.T) {
		assert := assert.New(t)
		type A struct{ A string }

		tpl, err := NewSerializerTemplate(StringField("A", "a"))
		assert.NoError(err)
		serializer, err := tpl.Serializer(&A{})
		assert.NoError(err)

		a := A{}

		// Can't deserialize int into string
		err = serializer.Decode([]byte(`{"a":1}`), &a)
		assert.Error(err)
		assert.EqualValues(ErrFieldDataIncorrectType, err)
	})
	t.Run("Test nesting with existing data", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
			B int
		}
		type S struct {
			A string
			Q *Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)

		s := S{A: "I'm a", Q: &Q{A: "old value", B: 42}}
		err = sSerializer.Decode([]byte(`{"a":"Hello","q":{"aa":"World!"}}`), &s)
		assert.NoError(err)
		assert.Equal("World!", s.Q.A)
		assert.EqualValues(42, s.Q.B, "Field not in serializer should remain untouched")
	})

	t.Run("Test nesting with nil", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
		}
		type S struct {
			A string
			Q *Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)

		s := S{}
		err = sSerializer.Decode([]byte(`{"a":"Hello","q":{"aa":"World!"}}`), &s)
		assert.NoError(err)
		assert.Equal("World!", s.Q.A)
	})
	t.Run("Test nesting, not a pointer embed ", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
		}
		type S struct {
			A string
			Q Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)

		s := S{}
		err = sSerializer.Decode([]byte(`{"a":"Hello","q":{"aa":"World!"}}`), &s)
		assert.NoError(err)
		assert.Equal("World!", s.Q.A)
	})
	t.Run("Test nesting, nil ", func(t *testing.T) {
		assert := assert.New(t)

		type Q struct {
			A string
		}
		type S struct {
			A string
			Q *Q
		}
		qTemplate, err := NewSerializerTemplate(StringField("A", "aa"))
		assert.NoError(err)
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)

		sTemplate, err := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		assert.NoError(err)
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)

		s := S{}
		err = sSerializer.Decode([]byte(`{"a":"Hello","q":null}`), &s)
		assert.NoError(err)
		assert.Nil(s.Q)
	})
}
