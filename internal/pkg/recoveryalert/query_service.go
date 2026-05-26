package recoveryalert

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Query struct {
	Username     string `form:"username"`
	Channel      string `form:"channel"`
	ChallengeID  string `form:"challengeId"`
	TargetMasked string `form:"targetMasked"`
	Offset       int    `form:"offset"`
	Limit        int    `form:"limit"`
}

type Record struct {
	RecordedAt      string `json:"recordedAt"`
	ChallengeID     string `json:"challengeId"`
	ChallengeStatus string `json:"challengeStatus"`
	UserID          int32  `json:"userId"`
	Username        string `json:"username"`
	RoleCode        string `json:"roleCode"`
	Channel         string `json:"channel"`
	TargetMasked    string `json:"targetMasked"`
	Error           string `json:"error"`
	ExpiresAt       string `json:"expiresAt"`
}

type ListResponse struct {
	Items  []Record `json:"items"`
	Total  int      `json:"total"`
	Offset int      `json:"offset"`
	Limit  int      `json:"limit"`
}

type SummaryResponse struct {
	Total           int            `json:"total"`
	ByChannel       map[string]int `json:"byChannel"`
	ByErrorCategory map[string]int `json:"byErrorCategory"`
	Trend           TrendSummary   `json:"trend"`
	TopErrors       []TopItem      `json:"topErrors"`
	TopUsernames    []TopItem      `json:"topUsernames"`
	TopChannels     []TopItem      `json:"topChannels"`
	Daily           []DailyPoint   `json:"daily"`
}

type TrendSummary struct {
	Last24Hours int `json:"last24Hours"`
	Last7Days   int `json:"last7Days"`
}

type DailyPoint struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type TopItem struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type QueryService interface {
	List(query Query) (*ListResponse, error)
	Summary(query Query) (*SummaryResponse, error)
}

type fileQueryService struct {
	path string
}

func NewQueryService(path string) QueryService {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		trimmed = filepath.Join("logs", "recovery-alerts.ndjson")
	}
	return &fileQueryService{path: trimmed}
}

func (s *fileQueryService) List(query Query) (*ListResponse, error) {
	items, err := s.loadRecords(query)
	if err != nil {
		return nil, err
	}
	limit := normalizeLimit(query.Limit)
	offset := normalizeOffset(query.Offset)
	total := len(items)
	if offset > total {
		offset = total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return &ListResponse{Items: items[offset:end], Total: total, Offset: offset, Limit: limit}, nil
}

func (s *fileQueryService) Summary(query Query) (*SummaryResponse, error) {
	items, err := s.loadRecords(query)
	if err != nil {
		return nil, err
	}
	byChannel := map[string]int{}
	byErrorCategory := map[string]int{}
	errorCounts := map[string]int{}
	usernameCounts := map[string]int{}
	channelCounts := map[string]int{}
	dailyCounts := initializeDailyBuckets(7)
	trend := TrendSummary{}
	now := time.Now()
	for _, item := range items {
		channel := strings.TrimSpace(item.Channel)
		if channel == "" {
			channel = "unknown"
		}
		errorCategory := classifyError(item.Error)
		byChannel[channel]++
		byErrorCategory[errorCategory]++
		channelCounts[channel]++
		username := strings.TrimSpace(item.Username)
		if username == "" {
			username = "unknown"
		}
		usernameCounts[username]++
		errorLabel := strings.TrimSpace(item.Error)
		if errorLabel == "" {
			errorLabel = "unknown"
		}
		errorCounts[errorLabel]++
		recordedAt := parseTime(item.RecordedAt)
		if !recordedAt.IsZero() {
			if recordedAt.After(now.Add(-24 * time.Hour)) {
				trend.Last24Hours++
			}
			if recordedAt.After(now.Add(-7 * 24 * time.Hour)) {
				trend.Last7Days++
			}
			key := recordedAt.Format("2006-01-02")
			if _, ok := dailyCounts[key]; ok {
				dailyCounts[key]++
			}
		}
	}
	return &SummaryResponse{
		Total:           len(items),
		ByChannel:       byChannel,
		ByErrorCategory: byErrorCategory,
		Trend:           trend,
		TopErrors:       topN(errorCounts, 5),
		TopUsernames:    topN(usernameCounts, 5),
		TopChannels:     topN(channelCounts, 5),
		Daily:           dailySeries(dailyCounts),
	}, nil
}

func (s *fileQueryService) loadRecords(query Query) ([]Record, error) {
	items := make([]Record, 0, normalizeLimit(query.Limit))
	file, err := os.Open(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Record{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var item Record
		if err := json.Unmarshal([]byte(line), &item); err != nil {
			continue
		}
		if query.Username != "" && item.Username != query.Username {
			continue
		}
		if query.Channel != "" && item.Channel != query.Channel {
			continue
		}
		if query.ChallengeID != "" && item.ChallengeID != query.ChallengeID {
			continue
		}
		if query.TargetMasked != "" && !strings.Contains(item.TargetMasked, query.TargetMasked) {
			continue
		}
		items = append(items, item)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return parseTime(items[i].RecordedAt).After(parseTime(items[j].RecordedAt))
	})
	return items, nil
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func normalizeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}

func parseTime(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func classifyError(message string) string {
	msg := strings.ToLower(strings.TrimSpace(message))
	switch {
	case strings.Contains(msg, "timeout"):
		return "timeout"
	case strings.Contains(msg, "config"), strings.Contains(msg, "auth"), strings.Contains(msg, "username"), strings.Contains(msg, "password"):
		return "config"
	case strings.Contains(msg, "temporar"), strings.Contains(msg, "refused"), strings.Contains(msg, "reset by peer"), strings.Contains(msg, "broken pipe"):
		return "temporary"
	case msg == "":
		return "unknown"
	default:
		return "permanent"
	}
}

func topN(source map[string]int, limit int) []TopItem {
	items := make([]TopItem, 0, len(source))
	for label, count := range source {
		items = append(items, TopItem{Label: label, Count: count})
	}
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Label < items[j].Label
		}
		return items[i].Count > items[j].Count
	})
	if limit > 0 && len(items) > limit {
		items = items[:limit]
	}
	return items
}

func initializeDailyBuckets(days int) map[string]int {
	result := make(map[string]int, days)
	today := time.Now()
	for i := days - 1; i >= 0; i-- {
		day := today.AddDate(0, 0, -i).Format("2006-01-02")
		result[day] = 0
	}
	return result
}

func dailySeries(source map[string]int) []DailyPoint {
	keys := make([]string, 0, len(source))
	for key := range source {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]DailyPoint, 0, len(keys))
	for _, key := range keys {
		result = append(result, DailyPoint{Date: key, Count: source[key]})
	}
	return result
}
