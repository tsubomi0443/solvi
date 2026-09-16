package page

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/labstack/echo/v5"
)

const logDir = "logs"

func (h *Handler) ManagementLogsPage(c *echo.Context) error {
	data := h.baseData(c, "management-logs")
	data["LogDatesJSON"] = mustJSON(listAvailableLogDates())
	return c.Render(http.StatusOK, "management_logs.html", data)
}

func listAvailableLogDates() []string {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return []string{}
	}

	dates := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "log_") || !strings.HasSuffix(name, ".log") || strings.Contains(name, "_access") {
			continue
		}
		compact := strings.TrimSuffix(strings.TrimPrefix(name, "log_"), ".log")
		if len(compact) != 8 {
			continue
		}
		if !hasLogPair(compact) {
			continue
		}
		dates[formatLogDate(compact)] = struct{}{}
	}

	result := make([]string, 0, len(dates))
	for date := range dates {
		result = append(result, date)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(result)))
	return result
}

func hasLogPair(compact string) bool {
	logPath := filepath.Join(logDir, fmt.Sprintf("log_%s.log", compact))
	accessPath := filepath.Join(logDir, fmt.Sprintf("log_%s_access.log", compact))
	if _, err := os.Stat(logPath); err != nil {
		return false
	}
	if _, err := os.Stat(accessPath); err != nil {
		return false
	}
	return true
}

func formatLogDate(compact string) string {
	return compact[:4] + "-" + compact[4:6] + "-" + compact[6:8]
}
