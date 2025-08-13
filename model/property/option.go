package property

import "reflect"

var (
	IndexPageSize = Property{
		KeyValue:     "post_index_page_size",
		DefaultValue: 10,
		Kind:         reflect.Int,
	}
	IndexSort = Property{
		KeyValue:     "post_index_sort",
		DefaultValue: "create_time",
		Kind:         reflect.String,
	}
)
