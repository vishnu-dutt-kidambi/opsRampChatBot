package opsramp

import (
	"fmt"
	"math"
)

// =============================================================================
// Capacity Forecasting
// =============================================================================
//
// Performs linear regression on historical metric data to predict when a
// resource will reach a configurable threshold (default 90%).
//
// The math runs in Go — the LLM's job is to decide when to call this tool
// and then summarize the result for the user. We NEVER ask the LLM to do
// the regression itself.
// =============================================================================

// ForecastResult contains the output of a capacity prediction.
type ForecastResult struct {
	ResourceID     string  `json:"resourceId"`
	ResourceName   string  `json:"resourceName"`
	MetricName     string  `json:"metricName"`
	Unit           string  `json:"unit"`
	CurrentValue   float64 `json:"currentValue"`
	DailyGrowth    float64 `json:"dailyGrowthRate"` // slope of line per day
	Threshold      float64 `json:"threshold"`       // target threshold (e.g., 90)
	DaysToThresh   int     `json:"daysToThreshold"` // -1 if declining or already exceeded
	PredictedDate  string  `json:"predictedDate"`   // "" if not applicable
	Trend          string  `json:"trend"`           // Rising, Stable, Declining
	Confidence     float64 `json:"rSquared"`        // R² goodness of fit (0.0-1.0)
	Recommendation string  `json:"recommendation"`
	DataPoints     int     `json:"dataPointsAnalyzed"`
}

// CapacityForecast runs linear regression on a metric series and predicts
// when it will breach the given threshold.
//
// threshold: the utilization % to predict against (typically 90 or 95).
// If threshold is 0, defaults to 90.
func CapacityForecast(series MetricSeries, resourceName string, threshold float64) ForecastResult {
	if threshold == 0 {
		threshold = 90.0
	}

	n := len(series.DataPoints)
	result := ForecastResult{
		ResourceID:   series.ResourceID,
		ResourceName: resourceName,
		MetricName:   series.MetricName,
		Unit:         series.Unit,
		Threshold:    threshold,
		DataPoints:   n,
	}

	if n < 2 {
		result.Trend = "Insufficient Data"
		result.DaysToThresh = -1
		result.Recommendation = "Not enough data points for forecasting (need at least 2 days)."
		return result
	}

	// Current value is the last data point
	result.CurrentValue = series.DataPoints[n-1].Value

	// Simple linear regression: y = slope*x + intercept
	// x = day index (0, 1, 2, ..., n-1)
	slope, intercept, rSquared := linearRegression(series.DataPoints)

	result.DailyGrowth = math.Round(slope*100) / 100 // round to 2 decimals
	result.Confidence = math.Round(rSquared*1000) / 1000

	// Classify trend
	switch {
	case slope > 0.5:
		result.Trend = "Rising"
	case slope < -0.5:
		result.Trend = "Declining"
	default:
		result.Trend = "Stable"
	}

	// Predict when threshold will be breached
	if result.CurrentValue >= threshold {
		result.DaysToThresh = 0
		result.PredictedDate = "Already exceeded"
		result.Recommendation = fmt.Sprintf(
			"%s is already at %.1f%% (threshold: %.0f%%). Immediate action required: scale up, add capacity, or clean up resources.",
			resourceName, result.CurrentValue, threshold)
	} else if slope <= 0 {
		result.DaysToThresh = -1
		result.PredictedDate = "Not projected to breach"
		result.Recommendation = fmt.Sprintf(
			"%s is %s at %.1f%% with a daily change of %.2f%%/day. No capacity risk detected.",
			resourceName, result.Trend, result.CurrentValue, slope)
	} else {
		// Days until regression line reaches threshold
		// threshold = slope * (n-1 + daysAhead) + intercept
		// daysAhead = (threshold - intercept)/slope - (n-1)
		currentX := float64(n - 1)
		threshX := (threshold - intercept) / slope
		daysAhead := threshX - currentX

		if daysAhead <= 0 {
			result.DaysToThresh = 0
			result.PredictedDate = "Imminent"
		} else {
			result.DaysToThresh = int(math.Ceil(daysAhead))
			result.PredictedDate = predictDate(result.DaysToThresh)
		}

		result.Recommendation = buildRecommendation(resourceName, series.MetricName,
			result.CurrentValue, slope, result.DaysToThresh, threshold)
	}

	return result
}

// linearRegression computes slope, intercept, and R² for y = slope*x + intercept
// where x = index (0, 1, 2, ..., n-1) and y = data point values.
func linearRegression(points []DataPoint) (slope, intercept, rSquared float64) {
	n := float64(len(points))
	var sumX, sumY, sumXY, sumX2 float64

	for i, p := range points {
		x := float64(i)
		sumX += x
		sumY += p.Value
		sumXY += x * p.Value
		sumX2 += x * x
	}

	// Slope and intercept
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n, 0
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n

	// R² (coefficient of determination)
	meanY := sumY / n
	var ssTot, ssRes float64
	for i, p := range points {
		predicted := slope*float64(i) + intercept
		ssRes += (p.Value - predicted) * (p.Value - predicted)
		ssTot += (p.Value - meanY) * (p.Value - meanY)
	}
	if ssTot == 0 {
		rSquared = 1.0
	} else {
		rSquared = 1.0 - ssRes/ssTot
	}

	return slope, intercept, rSquared
}

// predictDate returns a human-readable date string daysAhead from "today" (2026-02-20).
func predictDate(daysAhead int) string {
	// Simple month arithmetic from Feb 20, 2026
	day := 20 + daysAhead
	month := 2
	year := 2026

	daysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	for day > daysInMonth[month] {
		day -= daysInMonth[month]
		month++
		if month > 12 {
			month = 1
			year++
		}
	}

	months := []string{"", "Jan", "Feb", "Mar", "Apr", "May", "Jun",
		"Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	return fmt.Sprintf("%s %d, %d", months[month], day, year)
}

// buildRecommendation generates actionable advice based on the metric and trend.
func buildRecommendation(resourceName, metric string, current, slope float64, daysToThresh int, threshold float64) string {
	urgency := "Plan capacity expansion"
	if daysToThresh <= 3 {
		urgency = "CRITICAL — immediate action required"
	} else if daysToThresh <= 7 {
		urgency = "URGENT — action needed this week"
	} else if daysToThresh <= 14 {
		urgency = "WARNING — plan capacity within 2 weeks"
	}

	var action string
	switch {
	case contains(metric, "cpu"):
		action = fmt.Sprintf("Consider scaling up the instance, adding horizontal replicas, or investigating high-CPU processes on %s.", resourceName)
	case contains(metric, "disk"):
		action = fmt.Sprintf("Consider expanding disk volume, enabling log rotation, archiving old data, or cleaning up temp files on %s.", resourceName)
	case contains(metric, "memory"):
		action = fmt.Sprintf("Consider increasing instance memory, investigating memory leaks, or tuning application memory settings on %s.", resourceName)
	default:
		action = fmt.Sprintf("Review %s usage patterns on %s and plan for additional capacity.", metric, resourceName)
	}

	return fmt.Sprintf("%s. At %.2f%%/day growth, %s will reach %.0f%% in ~%d days. %s",
		urgency, slope, metric, threshold, daysToThresh, action)
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
