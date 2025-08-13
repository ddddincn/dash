package property

import (
	"dash/model/entity"
	"reflect"
	"strconv"
)

type Property struct {
	DefaultValue interface{}
	KeyValue     string
	Kind         reflect.Kind
}

func (p *Property) ConvertToOption() *entity.Option {
	var value string
	switch p.Kind {
	case reflect.Bool:
		value = strconv.FormatBool(p.DefaultValue.(bool))
	case reflect.Int:
		value = strconv.FormatInt(int64(p.DefaultValue.(int)), 10)
	case reflect.Int32:
		value = strconv.FormatInt(int64(p.DefaultValue.(int32)), 10)
	case reflect.Int64:
		value = strconv.FormatInt(p.DefaultValue.(int64), 10)
	case reflect.String:
		if p.DefaultValue != nil {
			value = p.DefaultValue.(string)
		}
	}
	return &entity.Option{
		OptionKey:   p.KeyValue,
		OptionValue: value,
	}
}

var AllProperty = []Property{
	BlogTitle,
	BlogURL,
	PostPermalinkType,
	SheetPermalinkType,
	CategoriesPrefix,
	TagsPrefix,
	ArchivesPrefix,
	SheetPrefix,
	PathSuffix,
	IsInstalled,
	BirthDay,
	SummaryLength,
	IndexPageSize,
	ArchivePageSize,
	IndexSort,
	JWTAccessSecret,
	JWTRefreshSecret,
}
