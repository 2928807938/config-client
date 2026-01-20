package middleware

import (
	"context"
	"runtime/debug"

	"config-client/share/errors"
	"config-client/share/types"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
)

// Recovery 统一错误恢复和处理中间件
// 捕获所有 panic 和业务异常，统一返回错误响应
func Recovery() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		defer func() {
			if err := recover(); err != nil {
				// 记录 panic 堆栈
				hlog.CtxErrorf(ctx, "Panic recovered: %v\nStack: %s", err, string(debug.Stack()))

				// 判断是否为 error 类型
				if e, ok := err.(error); ok {
					// 使用统一错误处理
					errors.HandleError(ctx, c, e)
				} else {
					// 非 error 类型的 panic，返回内部错误
					c.JSON(consts.StatusInternalServerError, types.Error(errors.InternalError, "服务器内部错误"))
				}

				// 终止后续处理
				c.Abort()
			}
		}()

		// 继续处理请求
		c.Next(ctx)
	}
}
