package audit

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"edu-schedule-system/configs"
)

type QueryService interface {
	List(query Query) (*ListResponse, error)
	Export(query Query) ([]ExportRow, error)
	Meta() map[string][]string
}

type logQueryService struct {
	logPaths []string
}

func NewQueryService(logPaths ...string) QueryService {
	paths := make([]string, 0, len(logPaths))
	for _, item := range logPaths {
		if strings.TrimSpace(item) != "" {
			paths = append(paths, item)
		}
	}
	if len(paths) == 0 {
		paths = []string{
			configs.ProjectAccessLogFile,
			filepath.Join("logs", configs.ProjectName+"-access.log"),
			filepath.Join("logs", "backend-runtime.log"),
			filepath.Join("logs", "backend-live.log"),
			filepath.Join("logs", "backend-manual.log"),
		}
	}
	return &logQueryService{logPaths: paths}
}

func (s *logQueryService) List(query Query) (*ListResponse, error) {
	items, total, offset, limit, err := s.queryRecords(query, normalizeLimit(query.Limit))
	if err != nil {
		return nil, err
	}
	return &ListResponse{Items: items, Total: total, Offset: offset, Limit: limit}, nil
}

func (s *logQueryService) Export(query Query) ([]ExportRow, error) {
	query.Offset = 0
	items, _, _, _, err := s.queryRecords(query, normalizeExportLimit(query.Limit))
	if err != nil {
		return nil, err
	}
	return toExportRows(items), nil
}

func (s *logQueryService) queryRecords(query Query, normalizedLimit int) ([]Record, int, int, int, error) {
	limit := normalizedLimit
	offset := normalizeOffset(query.Offset)
	items := make([]Record, 0, limit)
	startTime, hasStartTime := parseAuditTime(query.StartTime)
	endTime, hasEndTime := parseAuditTime(query.EndTime)
	for _, path := range s.logPaths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var entry logEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.Category != "audit" {
				continue
			}
			if query.Action != "" && entry.Action != query.Action {
				continue
			}
			if query.Module != "" && entry.Module != query.Module {
				continue
			}
			if query.ActorID > 0 && entry.ActorID != query.ActorID {
				continue
			}
			if query.ActorUsername != "" && entry.ActorUsername != query.ActorUsername {
				continue
			}
			if query.TargetID > 0 && entry.TargetID != query.TargetID {
				continue
			}
			if query.TraceID != "" && entry.TraceID != query.TraceID {
				continue
			}
			entryTime := parseTimeDescending(entry.Time)
			if hasStartTime && entryTime.Before(startTime) {
				continue
			}
			if hasEndTime && entryTime.After(endTime) {
				continue
			}
			items = append(items, Record{
				Time:          entry.Time,
				Level:         entry.Level,
				Message:       entry.Message,
				Action:        entry.Action,
				Module:        entry.Module,
				TargetID:      entry.TargetID,
				ActorID:       entry.ActorID,
				ActorUsername: entry.ActorUsername,
				ActorRole:     entry.ActorRole,
				TraceID:       entry.TraceID,
				Detail:        entry.Detail,
			})
		}
		_ = file.Close()
	}
	sort.SliceStable(items, func(i, j int) bool {
		return parseTimeDescending(items[i].Time).After(parseTimeDescending(items[j].Time))
	})
	total := len(items)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	items = items[offset:end]
	return items, total, offset, limit, nil
}

func (s *logQueryService) Meta() map[string][]string {
	actionsSet := map[string]struct{}{}
	modulesSet := map[string]struct{}{}
	for _, path := range s.logPaths {
		file, err := os.Open(path)
		if err != nil {
			continue
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			var entry logEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}
			if entry.Category != "audit" {
				continue
			}
			if strings.TrimSpace(entry.Action) != "" {
				actionsSet[entry.Action] = struct{}{}
			}
			if strings.TrimSpace(entry.Module) != "" {
				modulesSet[entry.Module] = struct{}{}
			}
		}
		_ = file.Close()
	}
	actions := make([]string, 0, len(actionsSet))
	for item := range actionsSet {
		actions = append(actions, item)
	}
	modules := make([]string, 0, len(modulesSet))
	for item := range modulesSet {
		modules = append(modules, item)
	}
	sort.Strings(actions)
	sort.Strings(modules)
	return map[string][]string{"actions": actions, "modules": modules}
}
