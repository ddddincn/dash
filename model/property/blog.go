package property

import "reflect"

var (
	BlogURL = Property{
		KeyValue:     "blog_url",
		DefaultValue: "",
		Kind:         reflect.String,
	}
	BlogTitle = Property{
		KeyValue:     "blog_title",
		DefaultValue: "",
		Kind:         reflect.String,
	}
)
