package logpolicy

import (
	"bufio"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type RetentionPolicy struct {
	LogDir        string
	Pattern       string
	MaxAgeDays    int
	MaxBytes      int64
	RotationCount int
}

type CleanupResult struct {
	Archived string
	Removed  []string
}

func Apply(policy RetentionPolicy) (*CleanupResult, error) {
	logDir := strings.TrimSpace(policy.LogDir)
	if logDir == "" {
		logDir = "logs"
	}
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, err
	}
	pattern := strings.TrimSpace(policy.Pattern)
	if pattern == "" {
		pattern = "*.ndjson"
	}
	matches, err := filepath.Glob(filepath.Join(logDir, pattern))
	if err != nil {
		return nil, err
	}
	result := &CleanupResult{Removed: []string{}}
	now := time.Now()
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}
		if policy.MaxBytes > 0 && info.Size() > policy.MaxBytes {
			archived, err := rotateTrimFile(path, policy.MaxBytes)
			if err == nil && archived != "" {
				result.Archived = archived
			}
		}
		if policy.MaxAgeDays > 0 && info.ModTime().Before(now.Add(-time.Duration(policy.MaxAgeDays)*24*time.Hour)) {
			if err := os.Remove(path); err == nil {
				result.Removed = append(result.Removed, path)
			}
		}
	}
	if policy.RotationCount > 0 {
		trimRotations(logDir, policy.RotationCount)
	}
	return result, nil
}

func rotateTrimFile(path string, maxBytes int64) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return "", err
	}
	if info.Size() <= maxBytes || maxBytes <= 0 {
		return "", nil
	}
	lines := make([]string, 0, 128)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	keep := make([]string, 0, len(lines))
	var total int64
	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i] + "\n"
		if total+int64(len(line)) > maxBytes && len(keep) > 0 {
			break
		}
		keep = append([]string{strings.TrimRight(line, "\n")}, keep...)
		total += int64(len(line))
	}
	archivePath := path + "." + time.Now().Format("20060102150405")
	if err := os.Rename(path, archivePath); err != nil {
		return "", err
	}
	body := strings.Join(keep, "\n")
	if body != "" {
		body += "\n"
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return archivePath, err
	}
	return archivePath, nil
}

func trimRotations(logDir string, keep int) {
	rotations, _ := filepath.Glob(filepath.Join(logDir, "*.ndjson.*"))
	if len(rotations) <= keep {
		return
	}
	sort.Slice(rotations, func(i, j int) bool {
		left, _ := os.Stat(rotations[i])
		right, _ := os.Stat(rotations[j])
		if left == nil || right == nil {
			return rotations[i] > rotations[j]
		}
		return left.ModTime().After(right.ModTime())
	})
	for _, path := range rotations[keep:] {
		_ = os.Remove(path)
	}
}
