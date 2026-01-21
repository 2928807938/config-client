package http

import (
	"context"
	"strconv"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// ReleaseHandler 发布管理HTTP处理器
type ReleaseHandler struct {
	releaseAppService *service.ReleaseAppService
}

// NewReleaseHandler 创建发布管理HTTP处理器
func NewReleaseHandler(releaseAppService *service.ReleaseAppService) *ReleaseHandler {
	return &ReleaseHandler{
		releaseAppService: releaseAppService,
	}
}

// CreateRelease 创建发布版本
// @Summary 创建发布版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param request body request.CreateReleaseRequest true "创建发布版本请求"
// @Success 200 {object} types.Response{data=vo.ReleaseVO}
// @Router /api/v1/releases [post]
func (h *ReleaseHandler) CreateRelease(ctx context.Context, c *app.RequestContext) {
	var req request.CreateReleaseRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	releaseVO, err := h.releaseAppService.CreateRelease(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("发布版本创建成功", releaseVO))
}

// PublishFull 全量发布
// @Summary 全量发布
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param request body request.PublishFullRequest true "全量发布请求"
// @Success 200 {object} types.Response
// @Router /api/v1/releases/publish-full [post]
func (h *ReleaseHandler) PublishFull(ctx context.Context, c *app.RequestContext) {
	var req request.PublishFullRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	err := h.releaseAppService.PublishFull(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("全量发布成功", nil))
}

// PublishCanary 灰度发布
// @Summary 灰度发布
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param request body request.PublishCanaryRequest true "灰度发布请求"
// @Success 200 {object} types.Response
// @Router /api/v1/releases/publish-canary [post]
func (h *ReleaseHandler) PublishCanary(ctx context.Context, c *app.RequestContext) {
	var req request.PublishCanaryRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	err := h.releaseAppService.PublishCanary(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("灰度发布成功", nil))
}

// Rollback 回滚到指定版本
// @Summary 回滚版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param request body request.ReleaseRollbackRequest true "回滚请求"
// @Success 200 {object} types.Response
// @Router /api/v1/releases/rollback [post]
func (h *ReleaseHandler) Rollback(ctx context.Context, c *app.RequestContext) {
	var req request.ReleaseRollbackRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	err := h.releaseAppService.Rollback(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("回滚成功", nil))
}

// GetReleaseByID 根据ID查询发布版本
// @Summary 根据ID查询发布版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param id path int true "发布版本ID"
// @Param include_snapshot query bool false "是否包含配置快照"
// @Success 200 {object} types.Response{data=vo.ReleaseVO}
// @Router /api/v1/releases/{id} [get]
func (h *ReleaseHandler) GetReleaseByID(ctx context.Context, c *app.RequestContext) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.SuccessWithMessage("无效的发布版本ID", nil))
		return
	}

	includeSnapshot := c.DefaultQuery("include_snapshot", "false") == "true"

	releaseVO, err := h.releaseAppService.GetReleaseByID(ctx, id, includeSnapshot)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(releaseVO))
}

// GetLatestPublishedRelease 获取最新已发布版本
// @Summary 获取最新已发布版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param namespace_id query int true "命名空间ID"
// @Param environment query string true "环境"
// @Success 200 {object} types.Response{data=vo.ReleaseVO}
// @Router /api/v1/releases/latest [get]
func (h *ReleaseHandler) GetLatestPublishedRelease(ctx context.Context, c *app.RequestContext) {
	namespaceIDStr := c.Query("namespace_id")
	namespaceID, err := strconv.Atoi(namespaceIDStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.SuccessWithMessage("无效的命名空间ID", nil))
		return
	}

	environment := c.Query("environment")
	if environment == "" {
		c.JSON(consts.StatusBadRequest, types.SuccessWithMessage("环境参数不能为空", nil))
		return
	}

	releaseVO, err := h.releaseAppService.GetLatestPublishedRelease(ctx, namespaceID, environment)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(releaseVO))
}

// ListReleasesByNamespace 查询命名空间下的所有发布版本
// @Summary 查询命名空间下的所有发布版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param namespace_id query int true "命名空间ID"
// @Param environment query string true "环境"
// @Success 200 {object} types.Response{data=[]vo.ReleaseVO}
// @Router /api/v1/releases/list [get]
func (h *ReleaseHandler) ListReleasesByNamespace(ctx context.Context, c *app.RequestContext) {
	namespaceIDStr := c.Query("namespace_id")
	namespaceID, err := strconv.Atoi(namespaceIDStr)
	if err != nil {
		c.JSON(consts.StatusBadRequest, types.SuccessWithMessage("无效的命名空间ID", nil))
		return
	}

	environment := c.Query("environment")
	if environment == "" {
		c.JSON(consts.StatusBadRequest, types.SuccessWithMessage("环境参数不能为空", nil))
		return
	}

	releases, err := h.releaseAppService.ListReleasesByNamespace(ctx, namespaceID, environment)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(releases))
}

// QueryReleases 分页查询发布版本
// @Summary 分页查询发布版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param namespace_id query int false "命名空间ID"
// @Param environment query string false "环境"
// @Param status query string false "状态"
// @Param release_type query string false "发布类型"
// @Param version_name query string false "版本名称"
// @Param page query int false "页码"
// @Param size query int false "每页数量"
// @Param order_by query string false "排序字段"
// @Success 200 {object} types.Response{data=vo.ReleaseListVO}
// @Router /api/v1/releases [get]
func (h *ReleaseHandler) QueryReleases(ctx context.Context, c *app.RequestContext) {
	var req request.QueryReleaseRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	releaseListVO, err := h.releaseAppService.QueryReleases(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(releaseListVO))
}

// CompareReleases 对比两个版本
// @Summary 对比两个版本
// @Tags 发布管理
// @Accept json
// @Produce json
// @Param request body request.CompareReleasesRequest true "对比版本请求"
// @Success 200 {object} types.Response{data=vo.ReleaseCompareVO}
// @Router /api/v1/releases/compare [post]
func (h *ReleaseHandler) CompareReleases(ctx context.Context, c *app.RequestContext) {
	var req request.CompareReleasesRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	compareVO, err := h.releaseAppService.CompareReleases(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(compareVO))
}
