package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// SubscriptionHandler 订阅管理HTTP处理器
type SubscriptionHandler struct {
	subscriptionAppService *service.SubscriptionAppService
}

// NewSubscriptionHandler 创建订阅管理HTTP处理器
func NewSubscriptionHandler(subscriptionAppService *service.SubscriptionAppService) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionAppService: subscriptionAppService,
	}
}

// QuerySubscriptions 分页查询订阅
// @Summary 分页查询订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param namespace_id query int false "命名空间ID"
// @Param environment query string false "环境"
// @Param client_id query string false "客户端ID（模糊查询）"
// @Param is_active query bool false "是否激活"
// @Param page query int false "页码" default(1)
// @Param size query int false "每页数量" default(20)
// @Param order_by query string false "排序字段"
// @Success 200 {object} types.Response{data=vo.SubscriptionListVO}
// @Router /api/v1/subscriptions [get]
func (h *SubscriptionHandler) QuerySubscriptions(ctx context.Context, c *app.RequestContext) {
	var req request.QuerySubscriptionRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	result, err := h.subscriptionAppService.QuerySubscriptions(ctx, &req)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}

// DeactivateSubscription 停用订阅
// @Summary 停用订阅
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Param request body request.DeactivateSubscriptionRequest true "停用订阅请求"
// @Success 200 {object} types.Response{data=vo.SubscriptionVO}
// @Router /api/v1/subscriptions/deactivate [post]
func (h *SubscriptionHandler) DeactivateSubscription(ctx context.Context, c *app.RequestContext) {
	var req request.DeactivateSubscriptionRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	subscription, err := h.subscriptionAppService.DeactivateSubscription(ctx, req.ID)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("订阅已停用", subscription))
}

// GetStatistics 获取订阅统计
// @Summary 获取订阅统计
// @Tags 订阅管理
// @Accept json
// @Produce json
// @Success 200 {object} types.Response{data=vo.SubscriptionStatisticsVO}
// @Router /api/v1/subscriptions/statistics [get]
func (h *SubscriptionHandler) GetStatistics(ctx context.Context, c *app.RequestContext) {
	result, err := h.subscriptionAppService.GetStatistics(ctx)
	if err != nil {
		panic(err)
	}

	c.JSON(consts.StatusOK, types.Success(result))
}
