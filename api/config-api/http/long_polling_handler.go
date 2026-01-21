package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// LongPollingHandler 长轮询HTTP处理器
type LongPollingHandler struct {
	longPollingAppService *service.LongPollingAppService
}

// NewLongPollingHandler 创建长轮询处理器
func NewLongPollingHandler(longPollingAppService *service.LongPollingAppService) *LongPollingHandler {
	return &LongPollingHandler{
		longPollingAppService: longPollingAppService,
	}
}

// Watch 长轮询监听配置变更
// @Summary 长轮询监听配置变更
// @Tags 配置管理
// @Accept json
// @Produce json
// @Param request body request.LongPollingRequest true "长轮询请求"
// @Success 200 {object} types.Response{data=vo.LongPollingResponse}
// @Router /api/v1/configs/watch [post]
func (h *LongPollingHandler) Watch(ctx context.Context, c *app.RequestContext) {
	var req request.LongPollingRequest
	if err := c.BindAndValidate(&req); err != nil {
		panic(err)
	}

	resp, err := h.longPollingAppService.WaitForChanges(ctx, &req)
	if err != nil {
		panic(err)
	}

	// 根据是否有变更返回不同的响应
	if !resp.Changed {
		c.JSON(consts.StatusOK, types.SuccessWithMessage("配置未变更", resp))
		return
	}

	c.JSON(consts.StatusOK, types.SuccessWithMessage("配置已变更", resp))
}
