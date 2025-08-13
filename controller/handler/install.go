package handler

import (
	"context"

	"dash/model/param"
	"dash/service"
)

type InstallHandler struct {
	InstallService service.InstallService
	OptionService  service.OptionService
}

func NewInstallHandler(installService service.InstallService, optionService service.OptionService) *InstallHandler {
	return &InstallHandler{
		InstallService: installService,
		OptionService:  optionService,
	}
}

func (i *InstallHandler) InstallBlog(installParam *param.Install) (interface{}, error) {
	err := i.InstallService.InstallBlog(context.Background(), installParam)
	if err != nil {
		return nil, err
	}
	return "安装完成", nil
}

// func (i *InstallHandler) IsInstall(ctx *gin.Context) (interface{}, error) {
// 	return i.OptionService.GetOrByDefaultWithErr(ctx, property.IsInstalled, property.IsInstalled.DefaultValue)
// }
