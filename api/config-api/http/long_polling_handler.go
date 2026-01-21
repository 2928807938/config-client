package http

import (
	"context"

	"config-client/api/config-api/dto/request"
	"config-client/api/config-api/service"
	"config-client/share/errors"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// LongPollingHandler 长轮询HTTP处理器
type LongPollingHandler struct {
	longPollingService *service.LongPollingService
}

// NewLongPollingHandler 创建长轮询处理器
func NewLongPollingHandler(longPollingService *service.LongPollingService) *LongPollingHandler {
	return &LongPollingHandler{
		longPollingService: longPollingService,
	}
}

// Watch 长轮询监听配置变更
// POST /api/v1/configs/watch
func (h *LongPollingHandler) Watch(c context.Context, ctx *app.RequestContext) {
	// 1. 绑定请求参数
	var req request.LongPollingRequest
	if err := ctx.BindAndValidate(&req); err != nil {
		hlog.Errorf("绑定长轮询请求参数失败: %v", err)
		errors.HandleError(ctx, errors.NewValidationError("请求参数无效", err))
		return
	}

	// 2. 调用应用服务等待配置变更
	resp, err := h.longPollingService.WaitForChanges(c, &req)
	if err != nil {
		hlog.Errorf("长轮询等待配置变更失败: %v", err)
		errors.HandleError(ctx, errors.NewInternalError("长轮询失败", err))
		return
	}

	// 3. 根据是否有变更返回不同的状态码
	if !resp.Changed {
		// 304 Not Modified - 配置未变更
		ctx.JSON(consts.StatusNotModified, map[string]interface{}{
			"changed": false,
			"message": "配置未变更",
		})
		return
	}

	// 4. 200 OK - 配置已变更
	ctx.JSON(consts.StatusOK, map[string]interface{}{
		"success": true,
		"message": "配置已变更",
		"data":    resp,
	})
}
