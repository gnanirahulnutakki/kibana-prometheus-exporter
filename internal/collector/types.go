package collector

// KibanaStatus represents the response from /api/status
type KibanaStatus struct {
	Name    string        `json:"name"`
	UUID    string        `json:"uuid"`
	Version VersionInfo   `json:"version"`
	Status  StatusInfo    `json:"status"`
	Metrics MetricsInfo   `json:"metrics"`
}

// VersionInfo contains version details
type VersionInfo struct {
	Number        string `json:"number"`
	BuildHash     string `json:"build_hash"`
	BuildNumber   int    `json:"build_number"`
	BuildSnapshot bool   `json:"build_snapshot"`
}

// StatusInfo contains overall and service status
type StatusInfo struct {
	Overall  OverallStatus           `json:"overall"`
	Core     map[string]*ServiceStatus `json:"core"`
	Plugins  map[string]*ServiceStatus `json:"plugins"`
}

// OverallStatus represents the overall system status
type OverallStatus struct {
	Level   string `json:"level"`
	Summary string `json:"summary"`
}

// ServiceStatus represents individual service status
type ServiceStatus struct {
	Level   string `json:"level"`
	Summary string `json:"summary"`
}

// MetricsInfo contains all metrics data
type MetricsInfo struct {
	CollectedAt           string                 `json:"collected_at"`
	ConcurrentConnections *int64                 `json:"concurrent_connections"`
	Process               ProcessMetrics         `json:"process"`
	OS                    *OSMetrics             `json:"os"`
	Requests              *RequestMetrics        `json:"requests"`
	ResponseTimes         *ResponseTimeMetrics   `json:"response_times"`
}

// ProcessMetrics contains process-level metrics
type ProcessMetrics struct {
	Memory         *MemoryMetrics `json:"memory"`
	EventLoopDelay *float64       `json:"event_loop_delay"`
	Uptime         *float64       `json:"uptime_in_millis"`
}

// MemoryMetrics contains memory usage details
type MemoryMetrics struct {
	Heap     *HeapMetrics `json:"heap"`
	Resident *int64       `json:"resident_set_size_in_bytes"`
}

// HeapMetrics contains heap memory details
type HeapMetrics struct {
	TotalBytes int64 `json:"total_in_bytes"`
	UsedBytes  int64 `json:"used_in_bytes"`
	SizeLimit  int64 `json:"size_limit"`
}

// OSMetrics contains operating system metrics
type OSMetrics struct {
	CPU    *CPUMetrics       `json:"cpu"`
	Load   *LoadMetrics      `json:"load"`
	Memory *OSMemoryMetrics  `json:"memory"`
}

// CPUMetrics contains CPU usage details
type CPUMetrics struct {
	ControlGroup *ControlGroupCPU `json:"cgroup"`
}

// ControlGroupCPU contains cgroup CPU metrics
type ControlGroupCPU struct {
	CPUPercent *float64 `json:"cpu_percent"`
}

// LoadMetrics contains system load averages
type LoadMetrics struct {
	Load1m  *float64 `json:"1m"`
	Load5m  *float64 `json:"5m"`
	Load15m *float64 `json:"15m"`
}

// OSMemoryMetrics contains OS memory details
type OSMemoryMetrics struct {
	TotalBytes *int64 `json:"total_in_bytes"`
	FreeBytes  *int64 `json:"free_in_bytes"`
	UsedBytes  *int64 `json:"used_in_bytes"`
}

// RequestMetrics contains HTTP request metrics
type RequestMetrics struct {
	Total       *int64         `json:"total"`
	Disconnects *int64         `json:"disconnects"`
	StatusCodes map[string]int `json:"status_codes"`
}

// ResponseTimeMetrics contains response time statistics
type ResponseTimeMetrics struct {
	Avg *float64 `json:"avg_in_millis"`
	Max *float64 `json:"max_in_millis"`
}
