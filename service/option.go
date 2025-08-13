package service

import (
	"context"
	"dash/model/param"
	"dash/model/property"
)

type OptionService interface {
	GetOrByDefault(ctx context.Context, p property.Property) interface{}
	GetOrByDefaultWithErr(ctx context.Context, p property.Property, defaultValue interface{}) (interface{}, error)
	GetIndexPageSize(ctx context.Context) int
	GetPostSort(ctx context.Context) param.Sort
	IsEnabledAbsolutePath(ctx context.Context) (bool, error)
	GetBlogBaseURL(ctx context.Context) (string, error)
	GetPathSuffix(ctx context.Context) (string, error)
	GetPostSummaryLength(ctx context.Context) int
	GetArchivePrefix(ctx context.Context) (string, error)

	OptionMap() map[string]property.Property
	Save(ctx context.Context, saveMap map[string]string) (err error)
}
