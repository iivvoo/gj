package gj

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
	t.Run("Regular String Decode", func(t *testing.T) {
		assert := assert.New(t)
		var target string

		assert.NoError(NStringField().Decode(&target, "Hello"))
		assert.Equal("Hello", target)
	})
	t.Run("String Decode wrong target type", func(t *testing.T) {
		assert := assert.New(t)
		var target int64

		err := NStringField().Decode(&target, "Hello")
		assert.Error(err)
		assert.Equal(ErrFieldIncorrectType, err)
	})
	t.Run("String Decode No Pointer", func(t *testing.T) {
		assert := assert.New(t)
		var target string

		err := NStringField().Decode(target, "Hello")
		assert.Error(err)
		assert.Equal(ErrFieldUnsettable, err)
	})
	t.Run("String Decode nil", func(t *testing.T) {
		assert := assert.New(t)

		err := NStringField().Decode(nil, "Hello")
		assert.Error(err)
		assert.Equal(ErrFieldUnsettable, err)
	})
	t.Run("String Decode wrong value", func(t *testing.T) {
		assert := assert.New(t)
		var target string

		err := NStringField().Decode(&target, 42)
		assert.Error(err)
		assert.Equal(ErrFieldDataIncorrectType, err)
	})
}
