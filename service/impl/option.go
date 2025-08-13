package impl

import (
	"context"
	"dash/cache"
	"dash/config"
	"dash/dal"
	"dash/log"
	"dash/model/entity"
	"dash/model/param"
	"dash/model/property"
	"dash/service"
	"dash/utils/xerr"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"go.uber.org/zap"
)

type optionServiceImpl struct {
	Config *config.Config
	Logger *zap.Logger
}

func NewOptionService(config *config.Config, logger *zap.Logger) service.OptionService {
	return &optionServiceImpl{
		Config: config,
		Logger: logger,
	}
}

// GetIndexPageSize 获取首页分页大小
func (o *optionServiceImpl) GetIndexPageSize(ctx context.Context) int {
	p := property.IndexPageSize
	value, err := o.getFromCacheMissFromDB(ctx, p)
	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
		return p.DefaultValue.(int)
	} else if err != nil {
		log.CtxErrorf(ctx, "query option err=%v", err)
		return p.DefaultValue.(int)
	}
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return p.DefaultValue.(int)
	}
}

// GetPostSort 获取文章排序方式
func (o *optionServiceImpl) GetPostSort(ctx context.Context) param.Sort {
	p := property.IndexSort
	value, err := o.getFromCacheMissFromDB(ctx, p)
	sort := p.DefaultValue.(string)

	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
	} else if err != nil {
		log.CtxErrorf(ctx, "query option err=%v", err)
	} else {
		sort = value.(string)
	}
	//
	return param.Sort{
		Fields: []string{"topPriority,desc", sort + ",desc", "id,desc"},
	}
}

func (o *optionServiceImpl) GetOrByDefault(ctx context.Context, p property.Property) interface{} {
	value, err := o.getFromCacheMissFromDB(ctx, p)
	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
		return p.DefaultValue
	}
	if err != nil {
		o.Logger.Error("get option", zap.String("key", p.KeyValue), zap.Error(err))
		return p.DefaultValue
	}
	if reflect.ValueOf(value).Kind() == reflect.Float64 {
		switch p.Kind {
		case reflect.Int64:
			value = int64(value.(float64))
		case reflect.Int32:
			value = int32(value.(float64))
		case reflect.Int:
			value = int(value.(float64))
		}
	}
	return value
}

func (o *optionServiceImpl) GetOrByDefaultWithErr(ctx context.Context, p property.Property, defaultValue interface{}) (interface{}, error) {
	value, err := o.getFromCacheMissFromDB(ctx, p)
	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, defaultValue)
		return defaultValue, nil
	}
	if err != nil {
		return nil, xerr.WithStatus(err, xerr.StatusInternalServerError)
	}
	if reflect.ValueOf(value).Kind() == reflect.Float64 {
		switch p.Kind {
		case reflect.Int64:
			value = int64(value.(float64))
		case reflect.Int32:
			value = int32(value.(float64))
		case reflect.Int:
			value = int(value.(float64))
		}

	}
	return value, nil
}

func (o *optionServiceImpl) IsEnabledAbsolutePath(ctx context.Context) (bool, error) {
	isEnabled, err := o.GetOrByDefaultWithErr(ctx, property.GlobalAbsolutePathEnabled, true)
	if err != nil {
		return true, err
	}
	return isEnabled.(bool), nil
}

func (o *optionServiceImpl) GetBlogBaseURL(ctx context.Context) (string, error) {
	blogURL, err := o.GetOrByDefaultWithErr(ctx, property.BlogURL, "")
	if err != nil {
		return "", err
	}
	if blogURL != "" {
		return blogURL.(string), nil
	}
	if o.Config.Server.Host == "0.0.0.0" {
		return fmt.Sprintf("http://127.0.0.1:%s", o.Config.Server.Port), nil
	} else {
		return fmt.Sprintf("http://%s:%s", o.Config.Server.Host, o.Config.Server.Port), nil
	}
}

func (o *optionServiceImpl) GetPathSuffix(ctx context.Context) (string, error) {
	p := property.PathSuffix
	value, err := o.getFromCacheMissFromDB(ctx, p)
	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
		return p.DefaultValue.(string), nil
	} else if err != nil {
		return "", err
	}
	return value.(string), nil
}

// getFromCacheMissFromDB 从缓存中获取配置值，如果不存在则从数据库中查询并缓存
func (o *optionServiceImpl) getFromCacheMissFromDB(ctx context.Context, p property.Property) (interface{}, error) {
	value, ok, err := cache.Get(p.KeyValue) // 从缓存中获取配置值
	if err != nil {
		return nil, err
	}
	if ok {
		return value, nil
	}

	optionDAL := dal.GetQueryByCtx(ctx).Option // 获取数据库查询实例
	// 从数据库中查询配置值
	option, err := optionDAL.WithContext(ctx).Where(optionDAL.OptionKey.Eq(p.KeyValue)).Take()
	if err != nil {
		return nil, WrapDBErr(err)
	}
	// 转换配置值为指定类型
	value, err = o.convert(option, p)
	if err != nil {
		return nil, err
	}
	// 更新缓存中的配置值
	cache.SetDefault(p.KeyValue, value)
	return value, nil
}

// convert 转换数据库中的配置值为指定类型
func (o *optionServiceImpl) convert(option *entity.Option, p property.Property) (interface{}, error) {
	var err error
	var result interface{}
	switch p.Kind {
	case reflect.Bool:
		result, err = strconv.ParseBool(option.OptionValue)
	case reflect.Int:
		result, err = strconv.Atoi(option.OptionValue)
	case reflect.Int32:
		v, e := strconv.ParseInt(option.OptionValue, 10, 32)
		result = int32(v)
		err = e
	case reflect.Int64:
		result, err = strconv.ParseInt(option.OptionValue, 10, 64)
	case reflect.String:
		result, err = option.OptionValue, nil
	}
	if err != nil {
		return nil, xerr.BadParam.Wrapf(err, "option 类型错误 optionValue=%v kind=%v", option.OptionValue, p.Kind)
	}
	return result, nil
}

func (o *optionServiceImpl) GetPostSummaryLength(ctx context.Context) int {
	p := property.SummaryLength
	value, err := o.getFromCacheMissFromDB(ctx, p)

	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
		return p.DefaultValue.(int)
	} else if err != nil {
		log.CtxErrorf(ctx, "query option key=%v err=%v", p.KeyValue, err)
		return p.DefaultValue.(int)
	}
	switch v := value.(type) {
	case int:
		return v
	case float64:
		return int(v)
	default:
		return p.DefaultValue.(int)
	}
}

func (o *optionServiceImpl) GetArchivePrefix(ctx context.Context) (string, error) {
	p := property.ArchivesPrefix
	value, err := o.getFromCacheMissFromDB(ctx, p)
	if xerr.GetType(err) == xerr.NoRecord {
		cache.SetDefault(p.KeyValue, p.DefaultValue)
		return p.DefaultValue.(string), nil
	} else if err != nil {
		return "", err
	}
	return value.(string), nil
}

func (o *optionServiceImpl) OptionMap() map[string]property.Property {
	result := make(map[string]property.Property)
	for _, p := range property.AllProperty {
		result[p.KeyValue] = p
	}
	return result
}

func (o *optionServiceImpl) Save(ctx context.Context, saveMap map[string]string) (err error) {
	propertyMap := o.OptionMap()
	optionDAL := dal.GetQueryByCtx(ctx).Option
	options, err := optionDAL.WithContext(ctx).Find()
	if err != nil {
		return WrapDBErr(err)
	}
	optionKeyMap := make(map[string]*entity.Option)
	for _, option := range options {
		optionKeyMap[option.OptionKey] = option
	}
	toCreates := make([]*entity.Option, 0)
	toUpdates := make([]*entity.Option, 0)
	now := time.Now()
	for key, value := range saveMap {
		p, ok := propertyMap[key]
		if !ok {
			return xerr.BadParam.New("key=%v", key).WithMsg("option key not exist").WithStatus(xerr.StatusBadRequest)
		}
		temp := &entity.Option{
			OptionKey:   key,
			OptionValue: value,
		}
		// check type
		_, err := o.convert(temp, p)
		if err != nil {
			return err
		}
		option, ok := optionKeyMap[key]
		if ok {
			option.OptionValue = value
			option.UpdateTime = &now
			toUpdates = append(toUpdates, option)
		} else {
			toCreates = append(toCreates, &entity.Option{
				CreateTime:  now,
				OptionKey:   key,
				OptionValue: value,
			})
		}
	}

	// Update the database before deleting the cache.
	// Although there is a very small probability that this will lead to temporary cache and database data inconsistency,
	// but it is acceptable

	deleteKeys := make([]string, 0, len(toUpdates)+len(toCreates))
	for _, option := range toCreates {
		deleteKeys = append(deleteKeys, option.OptionKey)
	}
	for _, option := range toUpdates {
		deleteKeys = append(deleteKeys, option.OptionKey)
	}
	err = cache.BatchDelete(deleteKeys)
	if err != nil {
		return err
	}

	err = dal.GetQueryByCtx(ctx).Transaction(func(tx *dal.Query) error {
		optionDAL := tx.Option
		for _, toUpdate := range toUpdates {
			_, err := optionDAL.WithContext(ctx).Where(optionDAL.ID.Eq(toUpdate.ID), optionDAL.OptionKey.Eq(toUpdate.OptionKey)).UpdateColumnSimple(optionDAL.OptionValue.Value(toUpdate.OptionValue))
			if err != nil {
				return WrapDBErr(err)
			}
		}
		err := optionDAL.WithContext(ctx).Create(toCreates...)
		if err != nil {
			return WrapDBErr(err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
