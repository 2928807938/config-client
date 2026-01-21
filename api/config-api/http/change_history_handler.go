package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ChangeHistoryHandler 变更历史HTTP处理器
type ChangeHistoryHandler struct {
	changeHistoryAppService *service.ChangeHistoryAppService
}

// NewChangeHistoryHandler 创建变更历史HTTP处理器
func NewChangeHistoryHandler(
	changeHistoryAppService *service.ChangeHistoryAppService,
) *ChangeHistoryHandler {
	return &ChangeHistoryHandler{
		changeHistoryAppService: changeHistoryAppService,
	}
}

// QueryHistory 分页查询变更历史
// @Summary 分页查询变更历史
// @Tags 变更管理
// @Accept json
// @Produce json
// @Param config_id query int false "配置ID"
// @Param namespace_id query int false "命名空间ID"
// @Param config_key query string false "配置键（模糊查询）"
// @Param operation query string false "操作类型：CREATE/UPDATE/DELETE/ROLLBACK"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param operator query string false "操作人（模糊查询）"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Success 200 {object} types.Response{data=vo.ChangeHistoryListVO}
// @Router /api/v1/history [get]
func (h *ChangeHistoryHandler) QueryHistory(ctx context.Context, c *app.RequestContext) {
	var req request.QueryHistoryRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	result, err := h.changeHistoryAppService.QueryHistory(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}

// GetHistoryByID 根据ID查询变更记录
// @Summary 根据ID查询变更记录
// @Tags 变更管理
// @Accept json
// @Produce json
// @Param request body request.GetHistoryByIDRequest true "获取变更记录请求"
// @Success 200 {object} types.Response{data=vo.ChangeHistoryVO}
// @Router /api/v1/history/get [post]
func (h *ChangeHistoryHandler) GetHistoryByID(ctx context.Context, c *app.RequestContext) {
	var req request.GetHistoryByIDRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	result, err := h.changeHistoryAppService.GetHistoryByID(ctx, req.HistoryID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}

// GetConfigHistory 获取指定配置的变更历史
// @Summary 获取配置变更历史
// @Tags 变更管理
// @Accept json
// @Produce json
// @Param config_id query int true "配置ID"
// @Param limit query int false "返回数量" default(50)
// @Success 200 {object} types.Response{data=vo.ChangeHistoryListVO}
// @Router /api/v1/history/config [get]
func (h *ChangeHistoryHandler) GetConfigHistory(ctx context.Context, c *app.RequestContext) {
	var req request.GetConfigHistoryRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	result, err := h.changeHistoryAppService.GetConfigHistory(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}

// CompareVersions 对比两个版本
// @Summary 对比两个版本
// @Tags 变更管理
// @Accept json
// @Produce json
// @Param request body request.CompareVersionsRequest true "版本对比请求"
// @Success 200 {object} types.Response{data=vo.VersionCompareVO}
// @Router /api/v1/history/compare [post]
func (h *ChangeHistoryHandler) CompareVersions(ctx context.Context, c *app.RequestContext) {
	var req request.CompareVersionsRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	result, err := h.changeHistoryAppService.CompareVersions(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}

// Rollback 回滚配置到指定版本
// @Summary 回滚配置
// @Tags 变更管理
// @Accept json
// @Produce json
// @Param request body request.RollbackRequest true "回滚请求"
// @Success 200 {object} types.Response
// @Router /api/v1/history/rollback [post]
func (h *ChangeHistoryHandler) Rollback(ctx context.Context, c *app.RequestContext) {
	var req request.RollbackRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	if err := h.changeHistoryAppService.Rollback(ctx, &req); err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("配置回滚成功", nil))
}

// GetStatistics 获取变更统计信息
// @Summary 获取变更统计
// @Tags 变更管理
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=vo.ChangeStatisticsVO}
// @Router /api/v1/history/statistics [get]
func (h *ChangeHistoryHandler) GetStatistics(ctx context.Context, c *app.RequestContext) {
	result, err := h.changeHistoryAppService.GetStatistics(ctx)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}
