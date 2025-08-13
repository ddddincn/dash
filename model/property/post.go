package property

import "reflect"

var (
	SummaryLength = Property{
		KeyValue:     "post_summary_length",
		DefaultValue: 150,
		Kind:         reflect.Int,
	}
	ArchivePageSize = Property{
		KeyValue:     "post_archives_page_size",
		DefaultValue: 10,
		Kind:         reflect.Int,
	}
	CategoryPageSize = Property{
		KeyValue:     "post_category_page_size",
		DefaultValue: 10,
		Kind:         reflect.Int,
	}
	TagPageSize = Property{
		KeyValue:     "post_tag_page_size",
		DefaultValue: 10,
		Kind:         reflect.Int,
	}
)
