package zap_log

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 自定义 Context Key 类型（避免与其他包的 key 冲突）
type ctxKey string

// 预定义需要存储的上下文字段
const (
	CtxKeyRequestID ctxKey = "request_id" // 请求 ID
	CtxKeyUserID    ctxKey = "user_id"    // 用户 ID
)

var ZapLogger *zap.Logger

// 初始化 Zap Logger
func init() {
	// 生产环境：JSON 格式 + 只输出 Info 及以上级别
	prodConfig := zap.NewProductionConfig()
	prodConfig.EncoderConfig.TimeKey = "timestamp"                   // 时间字段名改为 timestamp
	prodConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder // 时间格式：ISO8601
	var err error
	ZapLogger, err = prodConfig.Build()
	if err != nil {
		panic("初始化 Zap Logger 失败: " + err.Error())
	}
	defer ZapLogger.Sync() // 程序退出时刷新日志缓存
}

// CtxInfo 封装带 ctx 的 Info 级别日志
func CtxInfo(ctx context.Context, msg string, fields ...zap.Field) {
	// 从 ctx 提取上下文字段，添加到日志字段中
	logFields := extractCtxFields(ctx)
	logFields = append(logFields, fields...) // 追加用户自定义字段
	ZapLogger.Info(msg, logFields...)
}

// CtxError 封装带 ctx 的 Error 级别日志
func CtxError(ctx context.Context, msg string, err error, fields ...zap.Field) {
	logFields := extractCtxFields(ctx)
	logFields = append(logFields, zap.Error(err)) // 附加错误信息
	logFields = append(logFields, fields...)
	ZapLogger.Error(msg, logFields...)
}

// extractCtxFields 从 ctx 提取固定上下文字段
func extractCtxFields(ctx context.Context) []zap.Field {
	var fields []zap.Field
	// 提取请求 ID（若 ctx 中不存在，字段值为 ""，不影响日志）
	if reqID, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		fields = append(fields, zap.String("request_id", reqID))
	}
	// 提取用户 ID
	if userID, ok := ctx.Value(CtxKeyUserID).(string); ok {
		fields = append(fields, zap.String("user_id", userID))
	}
	return fields
}
