package constants

// ContextKey 上下文键类型
type ContextKey string

const (
	// OperatorKey 操作人上下文键
	OperatorKey ContextKey = "operator"

	// OperatorIPKey 操作人IP上下文键
	OperatorIPKey ContextKey = "operator_ip"
)
