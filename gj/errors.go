package gj

import "errors"

// ErrFieldDataIncorrectType means the json data does not match the field
var ErrFieldDataIncorrectType = errors.New("Data is of incorrect type")

// ErrFieldDataOverflow means the json data would overflow the field
var ErrFieldDataOverflow = errors.New("Data would overflow field")

// ErrFieldUnsettable means the field is not settable
var ErrFieldUnsettable = errors.New("Field is not settable")

// ErrFieldIncorrectType means the field does not match the expcted type (impossible?)
var ErrFieldIncorrectType = errors.New("Field is of wrong type")

// ErrMemberFieldNotFound - the field defined in a template was not found on the struct
var ErrMemberFieldNotFound = errors.New("Member field not found on struct")

// ErrMemberFieldTypeMismatch - the field defined in a template does not match the type on the struct
var ErrMemberFieldTypeMismatch = errors.New("Serializer field type mismatch Member field")

// ErrDuplicateField - the same field was added multiple times to a template
var ErrDuplicateField = errors.New("Duplicate field")

// ErrDifferentType is returned if the target does not match the serializer type
var ErrDifferentType = errors.New("target is not of same type")

// ErrArrayNotSupported is returned if an attempt is made to deserialze a non-object json structure, e.g. `[1,2]`
var ErrArrayNotSupported = errors.New("(de)serialization of pure json arrays not supported")

// ErrNotAPointer - A pointer wasn't passed as target for deserialization
var ErrNotAPointer = errors.New("Deserializing into a non-pointer does not make sense")

// ErrNilPointer - A nil pointer was passed as target for deserialization
var ErrNilPointer = errors.New("Deserializing into a nil-pointer does not make sense")
