package audit

import (
	"encoding/csv"
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

type Query struct {
	Action        string `form:"action"`
	Module        string `form:"module"`
	ActorID       int32  `form:"actorId"`
	ActorUsername string `form:"actorUsername"`
	TargetID      int32  `form:"targetId"`
	TraceID       string `form:"traceId"`
	StartTime     string `form:"startTime"`
	EndTime       string `form:"endTime"`
	Offset        int    `form:"offset"`
	Limit         int    `form:"limit"`
}

type ExportRow struct {
	Time          string
	Level         string
	Action        string
	Module        string
	ActorID       int32
	ActorUsername string
	ActorRole     string
	TargetID      int32
	TraceID       string
	Message       string
	Detail        string
}

type Record struct {
	Time          string         `json:"time"`
	Level         string         `json:"level"`
	Message       string         `json:"message"`
	Action        string         `json:"action"`
	Module        string         `json:"module"`
	TargetID      int32          `json:"targetId"`
	ActorID       int32          `json:"actorId"`
	ActorUsername string         `json:"actorUsername"`
	ActorRole     string         `json:"actorRole"`
	TraceID       string         `json:"traceId"`
	Detail        map[string]any `json:"detail,omitempty"`
	Raw           map[string]any `json:"raw,omitempty"`
}

type ListResponse struct {
	Items  []Record `json:"items"`
	Total  int      `json:"total"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
}

type logEntry struct {
	Time          string         `json:"time"`
	Level         string         `json:"level"`
	Message       string         `json:"msg"`
	Category      string         `json:"category"`
	Action        string         `json:"action"`
	Module        string         `json:"module"`
	TargetID      int32          `json:"target_id"`
	ActorID       int32          `json:"actor_id"`
	ActorUsername string         `json:"actor_username"`
	ActorRole     string         `json:"actor_role"`
	TraceID       string         `json:"trace_id"`
	Detail        map[string]any `json:"detail"`
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func normalizeExportLimit(limit int) int {
	if limit <= 0 {
		return 1000
	}
	if limit > 5000 {
		return 5000
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func parseAuditTime(value string) (time.Time, bool) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func parseTimeDescending(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func formatDetail(detail map[string]any) string {
	if len(detail) == 0 {
		return ""
	}
	payload, err := json.Marshal(detail)
	if err != nil {
		return ""
	}
	return string(payload)
}

func toExportRows(items []Record) []ExportRow {
	rows := make([]ExportRow, 0, len(items))
	for _, item := range items {
		rows = append(rows, ExportRow{
			Time:          item.Time,
			Level:         item.Level,
			Action:        item.Action,
			Module:        item.Module,
			ActorID:       item.ActorID,
			ActorUsername: item.ActorUsername,
			ActorRole:     item.ActorRole,
			TargetID:      item.TargetID,
			TraceID:       item.TraceID,
			Message:       item.Message,
			Detail:        formatDetail(item.Detail),
		})
	}
	return rows
}

func WriteCSV(writer *csv.Writer, rows []ExportRow) error {
	if err := writer.Write([]string{"time", "level", "action", "module", "actor_id", "actor_username", "actor_role", "target_id", "trace_id", "message", "detail"}); err != nil {
		return err
	}
	for _, row := range rows {
		if err := writer.Write([]string{
			row.Time,
			row.Level,
			row.Action,
			row.Module,
			strconv.FormatInt(int64(row.ActorID), 10),
			row.ActorUsername,
			row.ActorRole,
			strconv.FormatInt(int64(row.TargetID), 10),
			row.TraceID,
			row.Message,
			row.Detail,
		}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
