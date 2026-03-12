package opsramp

import (
	"strings"
)

// Client simulates the OpsRamp API client. In production, this would make
// real HTTP calls to https://{tenant}.api.opsramp.com/api/v2/...
// For MVP, it operates on in-memory mock data with realistic filtering.
type Client struct {
	Alerts        []Alert
	Resources     []Resource
	Incidents     []Incident
	MetricHistory []MetricSeries
}

// NewClient creates a mock OpsRamp client pre-loaded with data.
func NewClient(alerts []Alert, resources []Resource, incidents []Incident, metricHistory []MetricSeries) *Client {
	return &Client{
		Alerts:        alerts,
		Resources:     resources,
		Incidents:     incidents,
		MetricHistory: metricHistory,
	}
}

// SearchAlerts filters alerts by state, priority, resource name, or free text.
// Mirrors: GET /api/v2/tenants/{tenantId}/alerts/search
func (c *Client) SearchAlerts(state, priority, resourceName, query string) []Alert {
	var results []Alert
	for _, a := range c.Alerts {
		if state != "" && !strings.EqualFold(a.CurrentState, state) {
			continue
		}
		if priority != "" && !strings.EqualFold(a.Priority, priority) {
			continue
		}
		if resourceName != "" && !containsIgnoreCase(a.Resource.Name, resourceName) &&
			!containsIgnoreCase(a.Device.Name, resourceName) {
			continue
		}
		if query != "" && !matchesAlertQuery(a, query) {
			continue
		}
		results = append(results, a)
	}
	return results
}

// GetAlertByID returns a single alert by its unique ID.
func (c *Client) GetAlertByID(id string) *Alert {
	for _, a := range c.Alerts {
		if strings.EqualFold(a.UniqueID, id) {
			return &a
		}
	}
	return nil
}

// SearchResources filters resources by cloud, region, type, state, tag, or free text.
// Mirrors: GET /api/v2/tenants/{tenantId}/resources/search
func (c *Client) SearchResources(cloud, region, resourceType, state, tag, query string) []Resource {
	var results []Resource
	for _, r := range c.Resources {
		if cloud != "" && !strings.EqualFold(r.Cloud, cloud) {
			continue
		}
		if region != "" && !strings.EqualFold(r.Region, region) {
			continue
		}
		if resourceType != "" && !containsIgnoreCase(r.Type, resourceType) &&
			!containsIgnoreCase(r.ResourceType, resourceType) {
			continue
		}
		if state != "" && !strings.EqualFold(r.State, state) {
			continue
		}
		if tag != "" && !matchesTag(r.Tags, tag) {
			continue
		}
		if query != "" && !matchesResourceQuery(r, query) {
			continue
		}
		results = append(results, r)
	}
	return results
}

// GetResourceByID returns a single resource by its ID.
func (c *Client) GetResourceByID(id string) *Resource {
	for _, r := range c.Resources {
		if strings.EqualFold(r.ID, id) {
			return &r
		}
	}
	return nil
}

// GetResourceByName returns the first resource matching the given name (partial match).
func (c *Client) GetResourceByName(name string) *Resource {
	for _, r := range c.Resources {
		if containsIgnoreCase(r.Name, name) || containsIgnoreCase(r.HostName, name) {
			return &r
		}
	}
	return nil
}

// SearchIncidents filters incidents by status, priority, or free text.
// Mirrors: GET /api/v2/tenants/{tenantId}/incidents/search
func (c *Client) SearchIncidents(status, priority, query string) []Incident {
	var results []Incident
	for _, inc := range c.Incidents {
		if status != "" && !strings.EqualFold(inc.Status, status) {
			continue
		}
		if priority != "" && !strings.EqualFold(inc.Priority, priority) {
			continue
		}
		if query != "" && !matchesIncidentQuery(inc, query) {
			continue
		}
		results = append(results, inc)
	}
	return results
}

// GetIncidentByID returns a single incident by its ID.
func (c *Client) GetIncidentByID(id string) *Incident {
	for _, inc := range c.Incidents {
		if strings.EqualFold(inc.ID, id) {
			return &inc
		}
	}
	return nil
}

// GetResourceMetrics returns metrics for a specific resource.
// Mirrors: GET /api/v2/tenants/{tenantId}/resources/{resourceId}/metrics
func (c *Client) GetResourceMetrics(resourceID string) *ResourceMetrics {
	r := c.GetResourceByID(resourceID)
	if r == nil {
		return nil
	}
	return &r.Metrics
}

// GetAlertsForResource returns all alerts associated with a given resource.
func (c *Client) GetAlertsForResource(resourceID string) []Alert {
	var results []Alert
	for _, a := range c.Alerts {
		if strings.EqualFold(a.Resource.ID, resourceID) || strings.EqualFold(a.Device.ID, resourceID) {
			results = append(results, a)
		}
	}
	return results
}

// GetIncidentsForResource returns all incidents associated with a given resource.
func (c *Client) GetIncidentsForResource(resourceID string) []Incident {
	var results []Incident
	for _, inc := range c.Incidents {
		for _, rid := range inc.ResourceIDs {
			if strings.EqualFold(rid, resourceID) {
				results = append(results, inc)
				break
			}
		}
	}
	return results
}

// InvestigateResource returns a comprehensive investigation report for a resource,
// combining resource details, active alerts, incidents, and metrics.
func (c *Client) InvestigateResource(resourceNameOrID string) *InvestigationReport {
	// Try by ID first, then by name
	r := c.GetResourceByID(resourceNameOrID)
	if r == nil {
		r = c.GetResourceByName(resourceNameOrID)
	}
	if r == nil {
		return nil
	}

	alerts := c.GetAlertsForResource(r.ID)
	incidents := c.GetIncidentsForResource(r.ID)

	return &InvestigationReport{
		Resource:  *r,
		Alerts:    alerts,
		Incidents: incidents,
		Metrics:   r.Metrics,
	}
}

// InvestigationReport is a composite view used by the investigate tool.
type InvestigationReport struct {
	Resource  Resource        `json:"resource"`
	Alerts    []Alert         `json:"alerts"`
	Incidents []Incident      `json:"incidents"`
	Metrics   ResourceMetrics `json:"metrics"`
}

// GetSummary returns high-level counts across the entire monitored environment.
func (c *Client) GetSummary() EnvironmentSummary {
	critAlerts, warnAlerts, infoAlerts := 0, 0, 0
	for _, a := range c.Alerts {
		switch strings.ToLower(a.CurrentState) {
		case "critical":
			critAlerts++
		case "warning":
			warnAlerts++
		default:
			infoAlerts++
		}
	}
	openInc, resolvedInc := 0, 0
	for _, inc := range c.Incidents {
		if strings.EqualFold(inc.Status, "Open") {
			openInc++
		} else {
			resolvedInc++
		}
	}
	clouds := map[string]int{}
	for _, r := range c.Resources {
		clouds[r.Cloud]++
	}
	return EnvironmentSummary{
		TotalResources:    len(c.Resources),
		TotalAlerts:       len(c.Alerts),
		CriticalAlerts:    critAlerts,
		WarningAlerts:     warnAlerts,
		InfoAlerts:        infoAlerts,
		OpenIncidents:     openInc,
		ResolvedIncidents: resolvedInc,
		CloudBreakdown:    clouds,
	}
}

// EnvironmentSummary provides a high-level view of the monitored infrastructure.
type EnvironmentSummary struct {
	TotalResources    int            `json:"totalResources"`
	TotalAlerts       int            `json:"totalAlerts"`
	CriticalAlerts    int            `json:"criticalAlerts"`
	WarningAlerts     int            `json:"warningAlerts"`
	InfoAlerts        int            `json:"infoAlerts"`
	OpenIncidents     int            `json:"openIncidents"`
	ResolvedIncidents int            `json:"resolvedIncidents"`
	CloudBreakdown    map[string]int `json:"cloudBreakdown"`
}

// =============================================================================
// Helper functions
// =============================================================================

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func matchesAlertQuery(a Alert, query string) bool {
	q := strings.ToLower(query)
	return containsIgnoreCase(a.Subject, q) ||
		containsIgnoreCase(a.Description, q) ||
		containsIgnoreCase(a.CurrentState, q) ||
		containsIgnoreCase(a.Priority, q) ||
		containsIgnoreCase(a.ServiceName, q) ||
		containsIgnoreCase(a.ProblemArea, q) ||
		containsIgnoreCase(a.Component, q) ||
		containsIgnoreCase(a.Metric, q) ||
		containsIgnoreCase(a.Resource.Name, q) ||
		containsIgnoreCase(a.Device.Name, q)
}

func matchesResourceQuery(r Resource, query string) bool {
	q := strings.ToLower(query)
	return containsIgnoreCase(r.Name, q) ||
		containsIgnoreCase(r.HostName, q) ||
		containsIgnoreCase(r.IPAddress, q) ||
		containsIgnoreCase(r.OSName, q) ||
		containsIgnoreCase(r.Type, q) ||
		containsIgnoreCase(r.Cloud, q) ||
		containsIgnoreCase(r.Region, q) ||
		matchesTag(r.Tags, q)
}

func matchesIncidentQuery(inc Incident, query string) bool {
	q := strings.ToLower(query)

	// Handle common natural language queries that map to status filters.
	// Users say "recent incidents" or "latest incidents" meaning open/active ones.
	recentTerms := []string{"recent", "latest", "new", "current", "active", "today"}
	for _, term := range recentTerms {
		if strings.Contains(q, term) {
			// Treat as "show open incidents"
			if strings.EqualFold(inc.Status, "Open") {
				return true
			}
		}
	}

	catName := ""
	if inc.Category != nil {
		catName = inc.Category.Name
	}
	subCatName := ""
	if inc.SubCategory != nil {
		subCatName = inc.SubCategory.Name
	}
	return containsIgnoreCase(inc.Subject, q) ||
		containsIgnoreCase(inc.Description, q) ||
		containsIgnoreCase(catName, q) ||
		containsIgnoreCase(subCatName, q) ||
		containsIgnoreCase(inc.AssignedTo.Name, q)
}

func matchesTag(tags map[string]string, query string) bool {
	q := strings.ToLower(query)
	for k, v := range tags {
		if containsIgnoreCase(k, q) || containsIgnoreCase(v, q) {
			return true
		}
	}
	return false
}

// GetMetricHistoryForResource returns all metric series for a resource.
func (c *Client) GetMetricHistoryForResource(resourceID string) []MetricSeries {
	var results []MetricSeries
	for _, s := range c.MetricHistory {
		if strings.EqualFold(s.ResourceID, resourceID) {
			results = append(results, s)
		}
	}
	return results
}

// PredictCapacity runs forecasting on a resource's metrics.
// If metric is empty, all available metrics are forecasted.
// threshold defaults to 90 if 0.
func (c *Client) PredictCapacity(resourceNameOrID, metric string, threshold float64) []ForecastResult {
	// Resolve resource
	r := c.GetResourceByID(resourceNameOrID)
	if r == nil {
		r = c.GetResourceByName(resourceNameOrID)
	}
	if r == nil {
		return nil
	}

	series := c.GetMetricHistoryForResource(r.ID)
	if len(series) == 0 {
		return nil
	}

	var results []ForecastResult
	for _, s := range series {
		if metric != "" && !containsIgnoreCase(s.MetricName, metric) {
			continue
		}
		results = append(results, CapacityForecast(s, r.Name, threshold))
	}
	return results
}

// PredictAllCapacity runs forecasting across ALL resources that have metric history.
// Returns a slice of forecasts sorted with the most critical (fewest days to threshold) first.
func (c *Client) PredictAllCapacity(metric string, threshold float64) []ForecastResult {
	if threshold == 0 {
		threshold = 90
	}

	// Collect unique resource IDs from metric history
	seen := make(map[string]bool)
	var resourceIDs []string
	for _, s := range c.MetricHistory {
		if !seen[s.ResourceID] {
			seen[s.ResourceID] = true
			resourceIDs = append(resourceIDs, s.ResourceID)
		}
	}

	var results []ForecastResult
	for _, rid := range resourceIDs {
		r := c.GetResourceByID(rid)
		if r == nil {
			continue
		}
		series := c.GetMetricHistoryForResource(rid)
		for _, s := range series {
			if metric != "" && !containsIgnoreCase(s.MetricName, metric) {
				continue
			}
			fc := CapacityForecast(s, r.Name, threshold)
			results = append(results, fc)
		}
	}

	// Sort: rising trends first, then by days to threshold (soonest first)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			swap := false
			if results[i].Trend != "Rising" && results[j].Trend == "Rising" {
				swap = true
			} else if results[i].Trend == results[j].Trend {
				di := results[i].DaysToThresh
				dj := results[j].DaysToThresh
				if di < 0 {
					di = 9999
				}
				if dj < 0 {
					dj = 9999
				}
				if dj < di {
					swap = true
				}
			}
			if swap {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}
