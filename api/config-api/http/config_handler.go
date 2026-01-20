package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ConfigHandler 配置HTTP处理器
type ConfigHandler struct {
	configAppService *service.ConfigAppService
}

// NewConfigHandler 创建配置HTTP处理器
func NewConfigHandler(configAppService *service.ConfigAppService) *ConfigHandler {
	return &ConfigHandler{
		configAppService: configAppService,
	}
}

// CreateConfig 创建配置
// @Summary 创建配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param request body request.CreateConfigRequest true "创建配置请求"
// @Success 200 {object} types.Response{data=vo.ConfigVO}
// @Router /api/v1/configs [post]
func (h *ConfigHandler) CreateConfig(ctx context.Context, c *app.RequestContext) {
	var req request.CreateConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	configVO, err := h.configAppService.CreateConfig(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("配置创建成功", configVO))
}

// UpdateConfig 更新配置
// @Summary 更新配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param request body request.UpdateConfigRequest true "更新配置请求"
// @Success 200 {object} types.Response{data=vo.ConfigVO}
// @Router /api/v1/configs [put]
func (h *ConfigHandler) UpdateConfig(ctx context.Context, c *app.RequestContext) {
	var req request.UpdateConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	configVO, err := h.configAppService.UpdateConfig(ctx, req.ID, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("配置更新成功", configVO))
}

// QueryConfigs 分页查询配置
// @Summary 分页查询配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param namespace_id query int false "命名空间ID"
// @Param key query string false "配置键（模糊查询）"
// @Param group_name query string false "配置分组"
// @Param environment query string false "环境"
// @Param is_active query bool false "是否激活"
// @Param is_released query bool false "是否已发布"
// @Param value_type query string false "值类型"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(10)
// @Param order_by query string false "排序字段"
// @Success 200 {object} types.Response{data=vo.ConfigListVO}
// @Router /api/v1/configs [get]
func (h *ConfigHandler) QueryConfigs(ctx context.Context, c *app.RequestContext) {
	var req request.QueryConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	configListVO, err := h.configAppService.QueryConfigs(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(configListVO))
}

// GetConfigByID 根据ID获取配置
// @Summary 根据ID获取配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param request body request.GetConfigByIDRequest true "获取配置请求"
// @Success 200 {object} types.Response{data=vo.ConfigVO}
// @Router /api/v1/configs/get [post]
func (h *ConfigHandler) GetConfigByID(ctx context.Context, c *app.RequestContext) {
	var req request.GetConfigByIDRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	configVO, err := h.configAppService.GetConfigByID(ctx, req.ID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(configVO))
}

// DeleteConfig 删除配置（逻辑删除）
// @Summary 删除配置
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param request body request.DeleteConfigRequest true "删除配置请求"
// @Success 200 {object} types.Response
// @Router /api/v1/configs [delete]
func (h *ConfigHandler) DeleteConfig(ctx context.Context, c *app.RequestContext) {
	var req request.DeleteConfigRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	if err := h.configAppService.DeleteConfig(ctx, req.ID); err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("配置删除成功", nil))
}
