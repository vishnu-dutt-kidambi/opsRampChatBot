package mockdata

import (
	"fmt"
	"opsramp-agent/opsramp"
)

// GetMetricHistory returns 30-day daily metric time-series for key resources.
// Each resource has CPU, memory, and disk series showing realistic growth trends.
// In production this would come from:
//
//	GET /api/v2/tenants/{tenantId}/resources/{resourceId}/metrics/{metricName}
func GetMetricHistory() []opsramp.MetricSeries {
	var series []opsramp.MetricSeries

	// ── web-server-prod-01 (res-001): CPU trending UP sharply ──────────
	series = append(series, generateSeries("res-001", "system.cpu.utilization", "%",
		// 30 daily values: CPU climbing from ~60% to 97%
		[]float64{
			60.2, 61.5, 63.1, 62.8, 65.0, 66.3, 68.5, 70.1, 72.3, 74.0,
			75.5, 76.2, 78.0, 79.5, 80.3, 82.0, 83.1, 85.5, 87.0, 88.2,
			89.5, 90.0, 91.3, 92.5, 93.0, 94.2, 95.0, 95.8, 96.5, 97.3,
		}))
	series = append(series, generateSeries("res-001", "system.memory.utilization", "%",
		[]float64{
			55.0, 55.2, 55.5, 56.0, 56.3, 56.8, 57.0, 57.5, 58.0, 58.3,
			58.5, 59.0, 59.2, 59.5, 60.0, 60.2, 60.5, 60.8, 61.0, 61.2,
			61.3, 61.5, 61.5, 61.8, 61.9, 62.0, 62.0, 62.1, 62.1, 62.1,
		}))
	series = append(series, generateSeries("res-001", "system.disk.utilization", "%",
		[]float64{
			38.0, 38.2, 38.5, 38.8, 39.0, 39.3, 39.5, 40.0, 40.3, 40.5,
			41.0, 41.2, 41.5, 41.8, 42.0, 42.3, 42.5, 43.0, 43.2, 43.5,
			43.8, 44.0, 44.2, 44.3, 44.5, 44.5, 44.8, 45.0, 45.0, 45.0,
		}))

	// ── db-primary-01 (res-005): Disk trending UP dangerously ──────────
	series = append(series, generateSeries("res-005", "system.cpu.utilization", "%",
		[]float64{
			55.0, 56.2, 57.0, 58.5, 59.0, 60.2, 61.0, 62.5, 63.0, 63.5,
			64.0, 64.5, 65.0, 65.2, 65.5, 66.0, 66.3, 66.5, 67.0, 67.2,
			67.5, 67.8, 68.0, 68.0, 68.0, 68.2, 68.2, 68.2, 68.2, 68.2,
		}))
	series = append(series, generateSeries("res-005", "system.memory.utilization", "%",
		[]float64{
			60.0, 61.0, 62.0, 62.5, 63.0, 64.0, 64.5, 65.5, 66.0, 67.0,
			67.5, 68.0, 69.0, 69.5, 70.0, 70.5, 71.0, 71.5, 72.0, 72.5,
			73.0, 73.5, 74.0, 74.0, 74.5, 74.5, 75.0, 75.0, 75.0, 75.0,
		}))
	series = append(series, generateSeries("res-005", "system.disk.utilization", "%",
		// Disk growing ~2%/day — currently at 92%, was 32% a month ago
		[]float64{
			32.0, 34.0, 36.0, 38.0, 40.0, 42.0, 44.0, 46.0, 48.0, 50.0,
			52.0, 54.0, 56.0, 58.0, 60.0, 62.0, 64.0, 66.0, 68.0, 70.0,
			72.0, 74.0, 76.0, 78.0, 80.0, 82.0, 84.0, 86.0, 89.0, 92.0,
		}))

	// ── app-server-prod-02 (res-003): Memory trending UP ──────────────
	series = append(series, generateSeries("res-003", "system.cpu.utilization", "%",
		[]float64{
			40.0, 41.0, 42.0, 43.0, 44.0, 45.0, 46.0, 47.0, 48.0, 49.0,
			50.0, 50.5, 51.0, 51.5, 52.0, 52.5, 53.0, 53.5, 54.0, 54.5,
			54.5, 55.0, 55.0, 55.0, 55.0, 55.0, 55.0, 55.0, 55.0, 55.0,
		}))
	series = append(series, generateSeries("res-003", "system.memory.utilization", "%",
		// Memory leak: steady climb from 58% to 88%
		[]float64{
			58.0, 59.0, 60.0, 61.0, 62.0, 63.0, 64.0, 65.5, 67.0, 68.0,
			69.5, 71.0, 72.0, 73.5, 74.5, 76.0, 77.0, 78.0, 79.5, 80.5,
			81.5, 82.5, 83.5, 84.5, 85.0, 86.0, 86.5, 87.0, 87.5, 88.0,
		}))
	series = append(series, generateSeries("res-003", "system.disk.utilization", "%",
		[]float64{
			35.0, 35.2, 35.5, 35.8, 36.0, 36.2, 36.5, 36.8, 37.0, 37.2,
			37.5, 37.8, 38.0, 38.2, 38.5, 38.8, 39.0, 39.2, 39.5, 39.8,
			40.0, 40.2, 40.5, 40.8, 41.0, 41.0, 41.2, 41.2, 41.5, 41.5,
		}))

	// ── k8s-node-04 (res-016): Healthy, stable resource ───────────────
	series = append(series, generateSeries("res-016", "system.cpu.utilization", "%",
		[]float64{
			40.0, 42.0, 38.0, 41.0, 39.0, 43.0, 40.0, 42.0, 38.0, 41.0,
			39.5, 40.5, 41.0, 39.0, 42.0, 40.0, 41.0, 39.0, 40.5, 41.0,
			40.0, 39.5, 41.0, 40.0, 40.5, 39.0, 41.0, 40.0, 40.5, 40.0,
		}))
	series = append(series, generateSeries("res-016", "system.memory.utilization", "%",
		[]float64{
			55.0, 56.0, 54.0, 55.5, 54.5, 56.5, 55.0, 56.0, 54.5, 55.5,
			55.0, 55.5, 56.0, 54.5, 55.0, 55.5, 55.0, 54.5, 55.0, 55.5,
			55.0, 55.0, 55.5, 55.0, 55.0, 55.5, 55.0, 55.0, 55.0, 55.0,
		}))
	series = append(series, generateSeries("res-016", "system.disk.utilization", "%",
		[]float64{
			30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0,
			30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0,
			30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0, 30.0,
		}))

	// ── azure-sql-prod-01 (res-017): DTU/CPU growing ──────────────────
	series = append(series, generateSeries("res-017", "system.cpu.utilization", "%",
		[]float64{
			50.0, 51.5, 53.0, 54.0, 55.5, 57.0, 58.0, 59.5, 60.0, 61.5,
			63.0, 64.0, 65.0, 66.5, 67.0, 68.5, 69.0, 70.0, 71.0, 72.0,
			73.0, 73.5, 74.0, 74.5, 75.0, 75.5, 76.0, 76.5, 77.0, 78.0,
		}))
	series = append(series, generateSeries("res-017", "system.memory.utilization", "%",
		[]float64{
			60.0, 60.0, 60.5, 60.5, 61.0, 61.0, 61.5, 61.5, 62.0, 62.0,
			62.5, 62.5, 63.0, 63.0, 63.5, 63.5, 64.0, 64.0, 64.5, 64.5,
			65.0, 65.0, 65.0, 65.5, 65.5, 65.5, 66.0, 66.0, 66.0, 66.0,
		}))
	series = append(series, generateSeries("res-017", "system.disk.utilization", "%",
		[]float64{
			40.0, 40.5, 41.0, 41.5, 42.0, 42.5, 43.0, 43.5, 44.0, 44.5,
			45.0, 45.5, 46.0, 46.5, 47.0, 47.5, 48.0, 48.5, 49.0, 49.5,
			50.0, 50.5, 51.0, 51.5, 52.0, 52.5, 53.0, 53.5, 54.0, 55.0,
		}))

	return series
}

// generateSeries builds a MetricSeries with daily data points over the last 30 days.
// Day 0 = 30 days ago, day 29 = today (2026-02-20).
func generateSeries(resourceID, metricName, unit string, values []float64) opsramp.MetricSeries {
	points := make([]opsramp.DataPoint, len(values))
	for i, v := range values {
		// Day 0 = Jan 22, day 29 = Feb 20 (today)
		day := 22 + i // January 22 + offset
		var ts string
		if day <= 31 {
			ts = fmt.Sprintf("2026-01-%02dT00:00:00+0000", day)
		} else {
			ts = fmt.Sprintf("2026-02-%02dT00:00:00+0000", day-31)
		}
		points[i] = opsramp.DataPoint{Timestamp: ts, Value: v}
	}
	return opsramp.MetricSeries{
		MetricName: metricName,
		ResourceID: resourceID,
		DataPoints: points,
		Unit:       unit,
	}
}
