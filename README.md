# gj - a flexible json encoder/decoder

gj is a flexible json encoder/decoder. It allows you to define your json (de)serialization onto
existing data type. This enables you to serialized data types outside of your control,
to have multiple (different) serializations on the same data type and to reuse the same definition for different data types.

## Motivation

The default way to define json (de)serialization is to use struct tags. I think this is a messy and inflexible way to define your serialization

- it becomes messy quickly, especially when mixed with other tag based configurations (e.g. 
  gorm and json on the same struct)
- there's no real validation on the tag definitions, it's easy to make mistakes/typo's unnoticed
- you cannot have different serializations for the same struct (e.g. the same model on both the API and a json store like CouchDB)
- you can not (easily) modify/update existing definitions

on top of that, it's really difficult to change the existing, builtin serialization of types, e.g. serialize time.Time differently.

### but not..

Speed. This project does not aim to be the fastest json (de)serializer. Though, because (de)serialization might be more efficient since you can more selectively chose what to (de)serialize and what not, it might perform better than average eventually.

## An example

```
type Q struct {
	A string
	B int
}
type S struct {
	A string
	Q *Q
}

qTemplate, _ := NewSerializerTemplate(StringField("A", "aa"))
qSerializer, _ := qTemplate.Serializer(&Q{})
sTemplate, _ := NewSerializerTemplate(StringField("A", "a"), StructField("Q", "q", qSerializer))
sSerializer, _ := sTemplate.Serializer(&S{})
s := S{A: "I'm a", Q: &Q{A: "old value", B: 42}}
err = sSerializer.Decode([]byte(`{"a":"Hello","q":{"aa":"World!"}}`), &s)
fmt.Println(s.Q.A, s.Q.B)
# will print "World!" 42
```

## Implementation

The toolkit is currently implemented on top of json.Marshal/json.Unmarshal, which (de)serializes into `map[string]interface{}`. This currently also means embedded json fields will still be respected and lower-cased (non-exported) fields will be ignored. This may change when full (de)serialization is implemented.

## License

See [LICENSE](LICENSE)

## Current status, history

The code is currently pre-alpha and in a prototype stage. It's not suitable for production or development use. Many features are missing, API's will change and bugs will be present.

A first usable alpha version should be available early januari 2021