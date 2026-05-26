package audit

import (
	"strings"

	"edu-schedule-system/internal/pkg/core"
	"edu-schedule-system/internal/proposal"

	"go.uber.org/zap"
)

type Event struct {
	Action   string
	Module   string
	TargetID int32
	Detail   map[string]interface{}
}

func Log(ctx core.Context, event Event) {
	if ctx == nil {
		return
	}
	logger := ctx.Logger()
	if logger == nil {
		return
	}
	actor := ctx.SessionUserInfo()
	fields := []zap.Field{
		zap.String("category", "audit"),
		zap.String("action", strings.TrimSpace(event.Action)),
		zap.String("module", strings.TrimSpace(event.Module)),
		zap.Int32("target_id", event.TargetID),
		zap.Int32("actor_id", actor.Id),
		zap.String("actor_username", actor.UserName),
		zap.String("actor_role", actor.RoleCode),
		zap.String("trace_id", traceID(ctx)),
	}
	if len(event.Detail) > 0 {
		fields = append(fields, zap.Any("detail", event.Detail))
	}
	logger.Info("audit-event", fields...)
}

func traceID(ctx core.Context) string {
	if ctx == nil || ctx.Trace() == nil {
		return ""
	}
	return ctx.Trace().ID()
}

func Detail(kv ...interface{}) map[string]interface{} {
	if len(kv) == 0 {
		return nil
	}
	result := make(map[string]interface{}, len(kv)/2)
	for i := 0; i+1 < len(kv); i += 2 {
		key, ok := kv[i].(string)
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		result[key] = kv[i+1]
	}
	return result
}

func ActorSummary(actor proposal.SessionUserInfo) map[string]interface{} {
	return map[string]interface{}{
		"actorId":       actor.Id,
		"actorUsername": actor.UserName,
		"actorRole":     actor.RoleCode,
	}
}
