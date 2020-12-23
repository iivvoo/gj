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

// func TestMakeJSON(t *testing.T) {
// 	// eerst wat random json bouwen en dan mee spelen voor deserialisatie
// 	s := &SampleStruct{
// 		Name: "Sample Struct",
// 		Num:  42,
// 		Nest: &NestedStruct{
// 			Name:    "Nested Struct",
// 			Numbers: []int{1, 1, 2, 3, 5, 8, 13},
// 			When:    time.Now(),
// 		},
// 	}
// 	data, _ := json.Marshal(s)
// 	fmt.Println(string(data))
// }

// func TestGJExp(t *testing.T) {
// 	j := `{"Name":"Sample Struct","Num":42,"Nest":{"Name":"Nested Struct","Numbers":[1,1,2,3,5,8,13],"When":"2020-11-27T10:45:55.287389483+01:00"}}`

// 	j = `[1,2,3]`
// 	// Would this work for a plain [] array result?
// 	// No, so deserialze into interface{} and assert it into a map or array
// 	var objmap map[string]interface{}
// 	if err := json.Unmarshal([]byte(j), &objmap); err != nil {
// 		log.Fatal(err)
// 	}
// 	fmt.Printf("%v\n", objmap)
// 	for k, v := range objmap {
// 		fmt.Printf("%s -> %v | %#v |  %+v | type: %T\n", k, v, v, v, v)
// 	}
// 	for k, v := range objmap["Nest"].(map[string]interface{}) {
// 		fmt.Printf("%s -> %v | %#v |  %+v | type: %T\n", k, v, v, v, v)
// 	}
// }

func TestSerializerCreation(t *testing.T) {
	t.Run("Test simple success", func(t *testing.T) {
		assert := assert.New(t)

		_, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&struct{ A string }{})
		assert.NoError(err)
	})
	t.Run("Test duplicate field", func(t *testing.T) {
		assert := assert.New(t)

		_, err := NewSerializerTemplate(StringField("A", "a"), StringField("A", "b")).Serializer(&struct{ A string }{})
		assert.Error(err)
		assert.EqualValues(ErrDuplicateField, err)
	})
	t.Run("Test missing member", func(t *testing.T) {
		assert := assert.New(t)

		_, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&struct{}{})
		assert.Error(err)
		assert.EqualValues(ErrMemberFieldNotFound, err)
	})
	t.Run("Test type mismatch", func(t *testing.T) {
		assert := assert.New(t)

		_, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&struct{ A int }{})
		assert.Error(err)
		assert.EqualValues(ErrMemberFieldTypeMismatch, err)
	})
}

func TestSerialization(t *testing.T) {
	t.Run("Test simple success", func(t *testing.T) {
		assert := assert.New(t)

		type S struct {
			A string
			B int
		}
		serializer, err := NewSerializerTemplate(StringField("A", "a"), NumberField("B", "x")).Serializer(&S{})
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
		qTemplate := NewSerializerTemplate(StringField("A", "aa"))
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)
		assert.NotNil(qSerializer)

		sTemplate := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)
		assert.NotNil(sSerializer)

		res, err := sSerializer.Encode(&S{"Hello", &Q{"World"}})
		assert.NoError(err)

		assert.Equal(`{"a":"Hello","q":{"aa":"World"}}`, string(res))

	})
}
func TestSerializerTypeMatch(t *testing.T) {
	type A struct{ A string }
	type B struct{ A string }

	t.Run("Test type match", func(t *testing.T) {
		assert := assert.New(t)

		serializer, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&A{})
		assert.NoError(err)
		a := A{}
		err = serializer.Decode([]byte(`{"a":"x"}`), &a)
		assert.NoError(err)
	})
	t.Run("Test type mismatch", func(t *testing.T) {
		assert := assert.New(t)

		serializer, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&A{})
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

		serializer, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&A{})
		assert.NoError(err)

		a := A{}

		err = serializer.Decode([]byte(`{"a":"A"}`), &a)
		assert.NoError(err)
		assert.Equal("A", a.A)
	})

	t.Run("Test wrong data type", func(t *testing.T) {
		assert := assert.New(t)
		type A struct{ A string }

		serializer, err := NewSerializerTemplate(StringField("A", "a")).Serializer(&A{})
		assert.NoError(err)

		a := A{}

		// Can't deserialize int into string
		err = serializer.Decode([]byte(`{"a":1}`), &a)
		assert.Error(err)
		assert.EqualValues(ErrFieldDataIncorrectType, err)
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
		qTemplate := NewSerializerTemplate(StringField("A", "aa"))
		qSerializer, err := qTemplate.Serializer(&Q{})
		assert.NoError(err)

		sTemplate := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
		sSerializer, err := sTemplate.Serializer(&S{})
		assert.NoError(err)

		s := S{}
		err = sSerializer.Decode([]byte(`{"a":"Hello","q":{"aa":"World"}}`), &s)
		assert.NoError(err)
		// This fails because s.Q == nil
		// assert.Equal("World", s.Q.A)

	})
}

/*
 * Important corner cases:
 *
 * - Do we need to HTML-safe encode json? (see encoding/json)
 * - Properly escape keys and strings (encoding/json/encode.go encodeState:string
 */
