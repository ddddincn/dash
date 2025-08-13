package property

import "reflect"

var (
	TagsPrefix = Property{
		DefaultValue: "tags",
		KeyValue:     "tags_prefix",
		Kind:         reflect.String,
	}
	PathSuffix = Property{
		DefaultValue: "",
		KeyValue:     "path_suffix",
		Kind:         reflect.String,
	}
	CategoriesPrefix = Property{
		DefaultValue: "categories",
		KeyValue:     "category_prefix",
		Kind:         reflect.String,
	}
	PostPermalinkType = Property{
		DefaultValue: "DEFAULT",
		KeyValue:     "post_permalink_type",
		Kind:         reflect.String,
	}
	SheetPermalinkType = Property{
		DefaultValue: "SECONDARY",
		KeyValue:     "sheet_permalink_type",
		Kind:         reflect.String,
	}
	SheetPrefix = Property{
		DefaultValue: "s",
		KeyValue:     "sheet_prefix",
		Kind:         reflect.String,
	}
	ArchivesPrefix = Property{
		DefaultValue: "archives",
		KeyValue:     "archives_prefix",
		Kind:         reflect.String,
	}
)
