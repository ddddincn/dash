package utils

import (
	"dash/utils/xerr"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ParamString(ctx *gin.Context, key string) (string, error) {
	str := ctx.Param(key)
	if str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func ParamInt32(ctx *gin.Context, key string) (int32, error) {
	str := ctx.Param(key)
	if str == "" {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}

func MustGetQueryString(ctx *gin.Context, key string) (string, error) {
	str, ok := ctx.GetQuery(key)
	if !ok || str == "" {
		return "", xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	return str, nil
}

func MustGetQueryInt32(ctx *gin.Context, key string) (int32, error) {
	str, ok := ctx.GetQuery(key)
	if !ok {
		return 0, xerr.WithStatus(nil, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("%s parameter does not exisit", key))
	}
	value, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0, xerr.WithStatus(err, xerr.StatusBadRequest).WithMsg(fmt.Sprintf("The parameter %s type is incorrect", key))
	}
	return int32(value), nil
}
