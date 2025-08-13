package property

import "reflect"

var (
	IsInstalled = Property{
		KeyValue:     "is_installed",
		DefaultValue: false,
		Kind:         reflect.Bool,
	}
	BirthDay = Property{
		KeyValue:     "birthday",
		DefaultValue: int64(0),
		Kind:         reflect.Int64,
	}
)
