package recoveryalert

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestDailySeriesOrdersAndPreservesBuckets(t *testing.T) {
	result := dailySeries(map[string]int{
		"2026-05-14": 3,
		"2026-05-12": 1,
		"2026-05-13": 0,
	})
	if len(result) != 3 {
		t.Fatalf("dailySeries() len = %d, want 3", len(result))
	}
	if result[0].Date != "2026-05-12" || result[0].Count != 1 {
		t.Fatalf("dailySeries()[0] = %#v", result[0])
	}
	if result[1].Date != "2026-05-13" || result[1].Count != 0 {
		t.Fatalf("dailySeries()[1] = %#v", result[1])
	}
	if result[2].Date != "2026-05-14" || result[2].Count != 3 {
		t.Fatalf("dailySeries()[2] = %#v", result[2])
	}
}

func TestQueryServiceListFiltersRecoveryAlerts(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "recovery-alerts.ndjson")
	now := time.Now()
	within24h := now.Add(-2 * time.Hour).Format(time.RFC3339)
	within7d := now.Add(-48 * time.Hour).Format(time.RFC3339)
	older := now.Add(-10 * 24 * time.Hour).Format(time.RFC3339)
	content := "{\"recordedAt\":\"" + within24h + "\",\"challengeId\":\"c1\",\"challengeStatus\":\"expired\",\"userId\":1,\"username\":\"alice\",\"roleCode\":\"platform_admin\",\"channel\":\"email\",\"targetMasked\":\"al***@example.local\",\"error\":\"smtp timeout\",\"expiresAt\":\"2026-05-16T01:10:00+08:00\"}\n" +
		"{\"recordedAt\":\"" + within7d + "\",\"challengeId\":\"c2\",\"challengeStatus\":\"expired\",\"userId\":2,\"username\":\"bob\",\"roleCode\":\"org_admin\",\"channel\":\"sms\",\"targetMasked\":\"138****0000\",\"error\":\"provider down\",\"expiresAt\":\"2026-05-16T01:15:00+08:00\"}\n" +
		"{\"recordedAt\":\"" + older + "\",\"challengeId\":\"c3\",\"challengeStatus\":\"expired\",\"userId\":3,\"username\":\"carol\",\"roleCode\":\"platform_admin\",\"channel\":\"email\",\"targetMasked\":\"ca***@example.local\",\"error\":\"smtp timeout\",\"expiresAt\":\"2026-05-16T01:20:00+08:00\"}\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	service := NewQueryService(path)
	resp, err := service.List(Query{Channel: "sms", Limit: 20})
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if resp.Total != 1 || len(resp.Items) != 1 || resp.Items[0].Username != "bob" {
		t.Fatalf("List() = %#v, want one sms record for bob", resp)
	}
	summary, err := service.Summary(Query{})
	if err != nil {
		t.Fatalf("Summary() error = %v", err)
	}
	if summary.Total != 3 || summary.ByChannel["email"] != 2 || summary.ByErrorCategory["timeout"] != 2 || summary.ByErrorCategory["permanent"] != 1 {
		t.Fatalf("Summary() = %#v", summary)
	}
	if summary.Trend.Last24Hours != 1 || summary.Trend.Last7Days != 2 {
		t.Fatalf("Summary() trend = %#v", summary.Trend)
	}
	if len(summary.Daily) != 7 {
		t.Fatalf("Summary() daily length = %d, want 7", len(summary.Daily))
	}
	dailyTotal := 0
	for _, item := range summary.Daily {
		dailyTotal += item.Count
	}
	if dailyTotal != 2 {
		t.Fatalf("Summary() daily total = %d, want 2", dailyTotal)
	}
	if len(summary.TopErrors) == 0 || summary.TopErrors[0].Label != "smtp timeout" || summary.TopErrors[0].Count != 2 {
		t.Fatalf("Summary() top errors = %#v", summary.TopErrors)
	}
	if len(summary.TopUsernames) == 0 || summary.TopUsernames[0].Label != "alice" || summary.TopUsernames[0].Count != 1 {
		t.Fatalf("Summary() top usernames = %#v", summary.TopUsernames)
	}
	if len(summary.TopChannels) == 0 || summary.TopChannels[0].Label != "email" || summary.TopChannels[0].Count != 2 {
		t.Fatalf("Summary() top channels = %#v", summary.TopChannels)
	}
}
