package gj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type WrongTarget interface{}

func FieldDecodeTests(t *testing.T, s NSerializer, target, value interface{}) {
	sp := fmt.Sprintf
	t.Run(sp("%#v Decode wrong target type", s), func(t *testing.T) {
		assert := assert.New(t)
		var target WrongTarget

		err := s.Decode(&target, value)
		assert.Error(err)
		assert.Equal(ErrFieldIncorrectType, err)
	})
	t.Run(sp("%#v Decode No Pointer", s), func(t *testing.T) {
		assert := assert.New(t)

		err := s.Decode(target, value)
		assert.Error(err)
		assert.Equal(ErrFieldUnsettable, err)
	})
	t.Run(sp("%#v Decode nil", s), func(t *testing.T) {
		assert := assert.New(t)

		err := s.Decode(nil, value)
		assert.Error(err)
		assert.Equal(ErrFieldUnsettable, err)
	})
	t.Run(sp("%#v Decode wrong value", s), func(t *testing.T) {
		assert := assert.New(t)

		// Not sure if this will always work
		err := s.Decode(&target, struct{}{})
		assert.Error(err)
		assert.Equal(ErrFieldDataIncorrectType, err)
	})
}

func TestGJNew(t *testing.T) {
	t.Run("Random", func(t *testing.T) {
		assert := assert.New(t)

		type Foo struct {
			A string
			B int
		}

		FooSer := NStructField(
			MemberField("A", "a", NStringField()),
			MemberField("B", "frop", NIntField()),
		)
		res, err := Template(FooSer).Serialize(&Foo{"Hello", 42})
		assert.NoError(err)
		fmt.Println(string(res))
		// assert.True(false)
	})

	t.Run("StringField Regular Decode", func(t *testing.T) {
		assert := assert.New(t)
		var target string

		assert.NoError(NStringField().Decode(&target, "Hello"))
		assert.EqualValues("Hello", target)
	})
	FieldDecodeTests(t, NStringField(), "", "Hello")
}
