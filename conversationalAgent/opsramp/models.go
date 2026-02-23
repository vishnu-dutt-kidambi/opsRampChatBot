package opsramp

// =============================================================================
// OpsRamp API Data Models
// =============================================================================
//
// These structs mirror the real OpsRamp API v2 response schemas from
// https://develop.opsramp.com/v2. Field names, JSON tags, and enums are
// aligned with the actual API so swapping from mock data to real calls
// requires minimal changes.
//
// Schema sources:
//   Alerts:    GET /api/v2/tenants/{tenantId}/alerts/search
//   Resources: GET /api/v2/tenants/{tenantId}/resources/search
//   Incidents: GET /api/v2/tenants/{clientId}/incidents/search
//   Tickets:   POST /api/v2/tenants/{clientId}/incidents (Create)
// =============================================================================

// Alert represents an OpsRamp monitoring alert.
// API: GET /api/v2/tenants/{tenantId}/alerts/search
type Alert struct {
	UniqueID              string   `json:"uniqueId"`
	Subject               string   `json:"subject"`
	Description           string   `json:"description"`
	CurrentState          string   `json:"currentState"` // Ok, Warning, Critical, Info
	OldState              string   `json:"oldState,omitempty"`
	ServiceName           string   `json:"serviceName"`
	ProblemArea           string   `json:"problemArea"`
	Acknowledged          bool     `json:"acknowledged"`
	Suppressed            bool     `json:"suppressed"`
	PermanentlySuppressed bool     `json:"permanentlySuppressed"`
	Closed                bool     `json:"closed"`
	Ticketed              bool     `json:"ticketed"`
	ClientUniqueID        string   `json:"clientUniqueId,omitempty"`
	AlertType             string   `json:"alertType"` // Monitoring, Maintenance, Agent, etc.
	App                   string   `json:"app"`
	Component             string   `json:"component"`
	AlertTime             string   `json:"alertTime"`
	Device                Device   `json:"device"`
	Resource              Resource `json:"resource"`
	Metric                string   `json:"metric"`
	MonitorName           string   `json:"monitorName"`
	RepeatCount           string   `json:"repeatCount"`
	TenantID              int      `json:"tenantId,omitempty"`
	Status                string   `json:"status"`   // New, Acknowledged, Ticketed, Closed
	Priority              string   `json:"priority"` // P0, P1, P2, P3, P4, P5, N/A
	ElapsedTime           string   `json:"elapsedTimeString"`
	CreatedDate           string   `json:"createdDate"`
	UpdatedTime           string   `json:"updatedTime"`
	RBA                   bool     `json:"rba"`                  // Run book automation applied
	EntityType            string   `json:"entityType,omitempty"` // RESOURCE, CLIENT
	EventType             string   `json:"eventType,omitempty"`  // ALERT, RCA, Inference Alert
	IncidentID            string   `json:"incidentId,omitempty"` // set when ticketed
}

// AlertSearchResponse represents a paginated alert search result.
type AlertSearchResponse struct {
	Results         []Alert `json:"results"`
	TotalCount      int     `json:"totalResults"`
	PageNo          int     `json:"pageNo"`
	PageSize        int     `json:"pageSize"`
	OrderBy         string  `json:"orderBy"`
	TotalPages      int     `json:"totalPages"`
	NextPage        bool    `json:"nextPage"`
	DescendingOrder bool    `json:"descendingOrder"`
}

// Device represents the device associated with an alert.
// Matches the device object in the alert search response.
type Device struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	HostName     string `json:"hostName"`
	IPAddress    string `json:"ipAddress"`
	Type         string `json:"type"` // DEVICE, CLOUD
	AliasName    string `json:"aliasName,omitempty"`
	ResourceName string `json:"resourceName,omitempty"`
	ResourceType string `json:"resourceType"`
}

// Resource represents a monitored infrastructure resource.
// API: GET /api/v2/tenants/{tenantId}/resources/search
type Resource struct {
	ID             string            `json:"id"`
	Name           string            `json:"name"`
	HostName       string            `json:"hostName"`
	IPAddress      string            `json:"ipAddress"`
	Type           string            `json:"type"`
	State          string            `json:"state"`
	OSName         string            `json:"os"`
	AgentInstalled bool              `json:"agentInstalled"`
	Status         string            `json:"status"`
	ResourceType   string            `json:"resourceType"`
	Cloud          string            `json:"cloud"`
	Region         string            `json:"region"`
	InstanceSize   string            `json:"instanceSize,omitempty"`
	Tags           map[string]string `json:"tags,omitempty"`
	Metrics        ResourceMetrics   `json:"metrics,omitempty"`
}

// ResourceMetrics holds current metric values for a resource.
type ResourceMetrics struct {
	CPUUtilization    float64 `json:"cpuUtilization"`
	MemoryUtilization float64 `json:"memoryUtilization"`
	DiskUtilization   float64 `json:"diskUtilization"`
	NetworkIn         float64 `json:"networkInMbps"`
	NetworkOut        float64 `json:"networkOutMbps"`
}

// Incident represents an OpsRamp incident/ticket.
// API: GET /api/v2/tenants/{tenantId}/incidents/search
// API: POST /api/v2/tenants/{clientId}/incidents (Create)
type Incident struct {
	ID            string      `json:"id"`
	Subject       string      `json:"subject"`
	Description   string      `json:"description"`
	Status        string      `json:"status"` // New, Open, Pending, Resolved, Closed, On Hold
	OldStatus     string      `json:"oldStatus,omitempty"`
	SubStatus     string      `json:"subStatus,omitempty"` // In Progress, Waiting, Escalated, etc.
	Priority      string      `json:"priority"`            // Very Low, Low, Normal, High, Urgent
	AssignedTo    User        `json:"assignedUser"`
	AssigneeGroup *Group      `json:"assigneeGroup,omitempty"`
	CreatedBy     User        `json:"createdUser"`
	CreatedDate   string      `json:"createdDate"`
	UpdatedTime   string      `json:"updatedDate"` // API uses updatedDate
	ResolvedDate  string      `json:"resolvedDate,omitempty"`
	DueDate       string      `json:"dueDate,omitempty"`
	Source        string      `json:"source,omitempty"` // PORTAL, ALERT, EMAIL, INTEGRATION
	Category      *Category   `json:"category,omitempty"`
	SubCategory   *Category   `json:"subCategory,omitempty"`
	SLADetails    *SLADetails `json:"slaDetails,omitempty"`
	AlertIDs      []string    `json:"alertIds,omitempty"`
	ResourceIDs   []string    `json:"resourceIds,omitempty"`
	ReOpenCount   int         `json:"reOpenCount,omitempty"`
}

// Category represents a ticket category or sub-category.
// Real API returns {"id": "1", "uniqueId": "SCAT-...", "name": "Infrastructure"}
type Category struct {
	ID       string `json:"id"`
	UniqueID string `json:"uniqueId,omitempty"`
	Name     string `json:"name"`
}

// Group represents an assignee group.
type Group struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// SLADetails represents SLA information attached to an incident.
// Mirrors the real API slaDetails object.
type SLADetails struct {
	ResolutionBreach bool `json:"resolutionBreach"`
	ResponseBreach   bool `json:"responseBreach"`
	ResolutionTime   int  `json:"resolutionTime,omitempty"` // seconds
	ResponseTime     int  `json:"responseTime,omitempty"`   // seconds
}

// User represents an OpsRamp user.
type User struct {
	ID        string `json:"id"`
	LoginName string `json:"loginName"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
	Email     string `json:"email"`
	Name      string `json:"name,omitempty"` // convenience field for display (not in real API)
}

// MetricSeries represents a time-series of metric data points.
type MetricSeries struct {
	MetricName string      `json:"metricName"`
	ResourceID string      `json:"resourceId"`
	DataPoints []DataPoint `json:"dataPoints"`
	Unit       string      `json:"unit"`
}

// DataPoint is a single metric measurement at a point in time.
type DataPoint struct {
	Timestamp string  `json:"timestamp"`
	Value     float64 `json:"value"`
}
