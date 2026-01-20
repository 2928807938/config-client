package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/errors"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// NamespaceHandler 命名空间HTTP处理器
type NamespaceHandler struct {
	namespaceAppService *service.NamespaceAppService
}

// NewNamespaceHandler 创建命名空间HTTP处理器
func NewNamespaceHandler(namespaceAppService *service.NamespaceAppService) *NamespaceHandler {
	return &NamespaceHandler{
		namespaceAppService: namespaceAppService,
	}
}

// CreateNamespace 创建命名空间
// @Summary 创建命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.CreateNamespaceRequest true "创建命名空间请求"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces [post]
func (h *NamespaceHandler) CreateNamespace(ctx context.Context, c *app.RequestContext) {
	var req request.CreateNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceVO, err := h.namespaceAppService.CreateNamespace(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("命名空间创建成功", namespaceVO))
}

// UpdateNamespace 更新命名空间
// @Summary 更新命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.UpdateNamespaceRequest true "更新命名空间请求"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces [put]
func (h *NamespaceHandler) UpdateNamespace(ctx context.Context, c *app.RequestContext) {
	var req request.UpdateNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceVO, err := h.namespaceAppService.UpdateNamespace(ctx, req.ID, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("命名空间更新成功", namespaceVO))
}

// DeleteNamespace 删除命名空间（软删除）
// @Summary 删除命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.DeleteNamespaceRequest true "删除命名空间请求"
// @Success 200 {object} types.Response
// @Router /api/v1/namespaces [delete]
func (h *NamespaceHandler) DeleteNamespace(ctx context.Context, c *app.RequestContext) {
	var req request.DeleteNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	if err := h.namespaceAppService.DeleteNamespace(ctx, req.ID); err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("命名空间删除成功", nil))
}

// ActivateNamespace 激活命名空间
// @Summary 激活命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.ActivateNamespaceRequest true "激活命名空间请求"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces/activate [put]
func (h *NamespaceHandler) ActivateNamespace(ctx context.Context, c *app.RequestContext) {
	var req request.ActivateNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceVO, err := h.namespaceAppService.ActivateNamespace(ctx, req.ID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("命名空间激活成功", namespaceVO))
}

// DeactivateNamespace 停用命名空间
// @Summary 停用命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.DeactivateNamespaceRequest true "停用命名空间请求"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces/deactivate [put]
func (h *NamespaceHandler) DeactivateNamespace(ctx context.Context, c *app.RequestContext) {
	var req request.DeactivateNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceVO, err := h.namespaceAppService.DeactivateNamespace(ctx, req.ID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("命名空间停用成功", namespaceVO))
}

// QueryNamespaces 分页查询命名空间
// @Summary 分页查询命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param name query string false "命名空间名称（模糊查询）"
// @Param is_active query bool false "是否激活"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} types.Response{data=vo.NamespaceListVO}
// @Router /api/v1/namespaces [get]
func (h *NamespaceHandler) QueryNamespaces(ctx context.Context, c *app.RequestContext) {
	var req request.QueryNamespaceRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceListVO, err := h.namespaceAppService.QueryNamespaces(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceListVO))
}

// GetNamespaceByID 根据ID获取命名空间
// @Summary 根据ID获取命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param request body request.GetNamespaceByIDRequest true "获取命名空间请求"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces/get [post]
func (h *NamespaceHandler) GetNamespaceByID(ctx context.Context, c *app.RequestContext) {
	var req request.GetNamespaceByIDRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	namespaceVO, err := h.namespaceAppService.GetNamespaceByID(ctx, req.ID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceVO))
}

// GetNamespaceByName 根据名称获取命名空间
// @Summary 根据名称获取命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param name query string true "命名空间名称"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces/name [get]
func (h *NamespaceHandler) GetNamespaceByName(ctx context.Context, c *app.RequestContext) {
	name := c.Query("name")
	if name == "" {
		panic(errors.ErrBadRequest("命名空间名称不能为空"))
	}

	namespaceVO, err := h.namespaceAppService.GetNamespaceByName(ctx, name)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceVO))
}

// GetActiveNamespace 获取激活的命名空间
// @Summary 获取激活的命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Param name query string true "命名空间名称"
// @Success 200 {object} types.Response{data=vo.NamespaceVO}
// @Router /api/v1/namespaces/active [get]
func (h *NamespaceHandler) GetActiveNamespace(ctx context.Context, c *app.RequestContext) {
	name := c.Query("name")
	if name == "" {
		panic(errors.ErrBadRequest("命名空间名称不能为空"))
	}

	namespaceVO, err := h.namespaceAppService.GetActiveNamespace(ctx, name)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceVO))
}

// ListAllNamespaces 获取所有命名空间（不分页）
// @Summary 获取所有命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=[]vo.NamespaceVO}
// @Router /api/v1/namespaces/all [get]
func (h *NamespaceHandler) ListAllNamespaces(ctx context.Context, c *app.RequestContext) {
	namespaceVOs, err := h.namespaceAppService.ListAllNamespaces(ctx)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceVOs))
}

// ListActiveNamespaces 获取所有激活的命名空间（不分页）
// @Summary 获取所有激活的命名空间
// @Tags 命名空间管理
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=[]vo.NamespaceVO}
// @Router /api/v1/namespaces/active/all [get]
func (h *NamespaceHandler) ListActiveNamespaces(ctx context.Context, c *app.RequestContext) {
	namespaceVOs, err := h.namespaceAppService.ListActiveNamespaces(ctx)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(namespaceVOs))
}
