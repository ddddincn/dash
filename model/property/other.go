package property

import "reflect"

var (
	GlobalAbsolutePathEnabled = Property{
		DefaultValue: true,
		KeyValue:     "global_absolute_path_enabled",
		Kind:         reflect.Bool,
	}
	JWTAccessSecret = Property{
		DefaultValue: "1234567890abcdefghijklmnopqrstuvwxyz",
		KeyValue:     "jwt_access_secret",
		Kind:         reflect.String,
	}
	JWTRefreshSecret = Property{
		DefaultValue: "zyxwvutsrqponmlkjihgfedcba0987654321",
		KeyValue:     "jwt_refresh_secret",
		Kind:         reflect.String,
	}
)
