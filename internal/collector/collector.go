package collector

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const namespace = "kibana"

// Config holds the exporter configuration
type Config struct {
	KibanaURL          string
	Username           string
	Password           string
	Timeout            time.Duration
	InsecureSkipVerify bool
}

// KibanaCollector collects metrics from Kibana
type KibanaCollector struct {
	config Config
	client *http.Client
	mutex  sync.Mutex

	// Metrics
	up                 *prometheus.Desc
	statusOverall      *prometheus.Desc
	statusCore         *prometheus.Desc
	statusElastic      *prometheus.Desc
	statusSavedObjects *prometheus.Desc

	// Performance metrics
	heapTotal      *prometheus.Desc
	heapUsed       *prometheus.Desc
	heapSizeLimit  *prometheus.Desc
	residentSet    *prometheus.Desc
	eventLoop      *prometheus.Desc
	requestsTotal  *prometheus.Desc
	responseTime   *prometheus.Desc
	concurrentConn *prometheus.Desc

	// Process metrics
	uptime           *prometheus.Desc
	processMemory    *prometheus.Desc
	osCPUPercent     *prometheus.Desc
	osLoadAvg1m      *prometheus.Desc
	osLoadAvg5m      *prometheus.Desc
	osLoadAvg15m     *prometheus.Desc
	osMemTotal       *prometheus.Desc
	osMemFree        *prometheus.Desc
	osMemUsed        *prometheus.Desc

	// Scrape metrics
	scrapeDuration *prometheus.Desc
	scrapeSuccess  *prometheus.Desc
}

// NewKibanaCollector creates a new collector
func NewKibanaCollector(config Config) *KibanaCollector {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: config.InsecureSkipVerify,
		},
	}

	client := &http.Client{
		Timeout:   config.Timeout,
		Transport: transport,
	}

	return &KibanaCollector{
		config: config,
		client: client,

		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", "up"),
			"Was the last scrape of Kibana successful",
			nil, nil,
		),
		statusOverall: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", "overall"),
			"Kibana overall status (1=green, 0.5=yellow, 0=red, -1=unknown)",
			nil, nil,
		),
		statusCore: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", "core"),
			"Kibana core status (1=available, 0=unavailable)",
			[]string{"name"}, nil,
		),
		statusElastic: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", "elasticsearch"),
			"Elasticsearch connection status (1=available, 0=unavailable)",
			nil, nil,
		),
		statusSavedObjects: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "status", "saved_objects"),
			"Saved objects status (1=available, 0=unavailable)",
			nil, nil,
		),

		// Heap metrics
		heapTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "heap", "total_bytes"),
			"Total heap size in bytes",
			nil, nil,
		),
		heapUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "heap", "used_bytes"),
			"Used heap size in bytes",
			nil, nil,
		),
		heapSizeLimit: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "heap", "size_limit_bytes"),
			"Heap size limit in bytes",
			nil, nil,
		),
		residentSet: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "memory", "resident_set_bytes"),
			"Resident set size in bytes",
			nil, nil,
		),
		eventLoop: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "event_loop", "delay_seconds"),
			"Event loop delay in seconds",
			nil, nil,
		),

		// Request metrics
		requestsTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "requests", "total"),
			"Total number of requests",
			[]string{"status"}, nil,
		),
		responseTime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "response_time", "seconds"),
			"Response time statistics",
			[]string{"quantile"}, nil,
		),
		concurrentConn: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "concurrent_connections", "total"),
			"Number of concurrent connections",
			nil, nil,
		),

		// Process metrics
		uptime: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process", "uptime_seconds"),
			"Kibana process uptime in seconds",
			nil, nil,
		),
		processMemory: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "process", "memory_bytes"),
			"Kibana process memory usage",
			[]string{"type"}, nil,
		),

		// OS metrics
		osCPUPercent: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "cpu_percent"),
			"OS CPU usage percentage",
			nil, nil,
		),
		osLoadAvg1m: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "load_average_1m"),
			"OS load average 1 minute",
			nil, nil,
		),
		osLoadAvg5m: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "load_average_5m"),
			"OS load average 5 minutes",
			nil, nil,
		),
		osLoadAvg15m: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "load_average_15m"),
			"OS load average 15 minutes",
			nil, nil,
		),
		osMemTotal: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "memory_total_bytes"),
			"OS total memory in bytes",
			nil, nil,
		),
		osMemFree: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "memory_free_bytes"),
			"OS free memory in bytes",
			nil, nil,
		),
		osMemUsed: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "os", "memory_used_bytes"),
			"OS used memory in bytes",
			nil, nil,
		),

		// Scrape metrics
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "duration_seconds"),
			"Duration of Kibana scrape",
			nil, nil,
		),
		scrapeSuccess: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "scrape", "success"),
			"Was the last scrape successful",
			nil, nil,
		),
	}
}

// Describe implements prometheus.Collector
func (c *KibanaCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.statusOverall
	ch <- c.statusCore
	ch <- c.statusElastic
	ch <- c.statusSavedObjects
	ch <- c.heapTotal
	ch <- c.heapUsed
	ch <- c.heapSizeLimit
	ch <- c.residentSet
	ch <- c.eventLoop
	ch <- c.requestsTotal
	ch <- c.responseTime
	ch <- c.concurrentConn
	ch <- c.uptime
	ch <- c.processMemory
	ch <- c.osCPUPercent
	ch <- c.osLoadAvg1m
	ch <- c.osLoadAvg5m
	ch <- c.osLoadAvg15m
	ch <- c.osMemTotal
	ch <- c.osMemFree
	ch <- c.osMemUsed
	ch <- c.scrapeDuration
	ch <- c.scrapeSuccess
}

// Collect implements prometheus.Collector
func (c *KibanaCollector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	status, err := c.scrapeKibana()
	duration := time.Since(start).Seconds()

	ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, duration)

	if err != nil {
		log.WithError(err).Error("Failed to scrape Kibana")
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, 0)
		return
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
	ch <- prometheus.MustNewConstMetric(c.scrapeSuccess, prometheus.GaugeValue, 1)

	// Export metrics from status
	c.exportStatus(ch, status)
}

// CheckHealth checks if Kibana is reachable
func (c *KibanaCollector) CheckHealth() error {
	req, err := http.NewRequest("GET", c.config.KibanaURL+"/api/status", nil)
	if err != nil {
		return err
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	req.Header.Set("kbn-xsrf", "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("kibana returned status %d", resp.StatusCode)
	}

	return nil
}

func (c *KibanaCollector) scrapeKibana() (*KibanaStatus, error) {
	req, err := http.NewRequest("GET", c.config.KibanaURL+"/api/status", nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if c.config.Username != "" {
		req.SetBasicAuth(c.config.Username, c.config.Password)
	}
	req.Header.Set("kbn-xsrf", "true")

	log.WithField("url", c.config.KibanaURL+"/api/status").Debug("Scraping Kibana")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	var status KibanaStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &status, nil
}

func (c *KibanaCollector) exportStatus(ch chan<- prometheus.Metric, status *KibanaStatus) {
	// Overall status
	statusValue := -1.0
	switch status.Status.Overall.Level {
	case "available", "green":
		statusValue = 1.0
	case "degraded", "yellow":
		statusValue = 0.5
	case "unavailable", "red":
		statusValue = 0.0
	}
	ch <- prometheus.MustNewConstMetric(c.statusOverall, prometheus.GaugeValue, statusValue)

	// Core services status
	for name, svc := range status.Status.Core {
		value := 0.0
		if svc.Level == "available" {
			value = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.statusCore, prometheus.GaugeValue, value, name)
	}

	// Elasticsearch status
	if status.Status.Core["elasticsearch"] != nil {
		value := 0.0
		if status.Status.Core["elasticsearch"].Level == "available" {
			value = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.statusElastic, prometheus.GaugeValue, value)
	}

	// Saved objects status
	if status.Status.Core["savedObjects"] != nil {
		value := 0.0
		if status.Status.Core["savedObjects"].Level == "available" {
			value = 1.0
		}
		ch <- prometheus.MustNewConstMetric(c.statusSavedObjects, prometheus.GaugeValue, value)
	}

	// Process memory metrics
	if status.Metrics.Process.Memory != nil {
		mem := status.Metrics.Process.Memory
		if mem.Heap != nil {
			ch <- prometheus.MustNewConstMetric(c.heapTotal, prometheus.GaugeValue, float64(mem.Heap.TotalBytes))
			ch <- prometheus.MustNewConstMetric(c.heapUsed, prometheus.GaugeValue, float64(mem.Heap.UsedBytes))
			ch <- prometheus.MustNewConstMetric(c.heapSizeLimit, prometheus.GaugeValue, float64(mem.Heap.SizeLimit))
		}
		if mem.Resident != nil {
			ch <- prometheus.MustNewConstMetric(c.residentSet, prometheus.GaugeValue, float64(*mem.Resident))
		}
	}

	// Event loop delay
	if status.Metrics.Process.EventLoopDelay != nil {
		ch <- prometheus.MustNewConstMetric(c.eventLoop, prometheus.GaugeValue, *status.Metrics.Process.EventLoopDelay/1000.0)
	}

	// Uptime
	if status.Metrics.Process.Uptime != nil {
		ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.GaugeValue, *status.Metrics.Process.Uptime/1000.0)
	}

	// Request metrics
	if status.Metrics.Requests != nil {
		reqs := status.Metrics.Requests
		if reqs.Total != nil {
			ch <- prometheus.MustNewConstMetric(c.requestsTotal, prometheus.CounterValue, float64(*reqs.Total), "total")
		}
		if reqs.Disconnects != nil {
			ch <- prometheus.MustNewConstMetric(c.requestsTotal, prometheus.CounterValue, float64(*reqs.Disconnects), "disconnects")
		}
		if reqs.StatusCodes != nil {
			for code, count := range reqs.StatusCodes {
				ch <- prometheus.MustNewConstMetric(c.requestsTotal, prometheus.CounterValue, float64(count), code)
			}
		}
	}

	// Concurrent connections
	if status.Metrics.ConcurrentConnections != nil {
		ch <- prometheus.MustNewConstMetric(c.concurrentConn, prometheus.GaugeValue, float64(*status.Metrics.ConcurrentConnections))
	}

	// Response time
	if status.Metrics.ResponseTimes != nil {
		rt := status.Metrics.ResponseTimes
		if rt.Avg != nil {
			ch <- prometheus.MustNewConstMetric(c.responseTime, prometheus.GaugeValue, *rt.Avg/1000.0, "avg")
		}
		if rt.Max != nil {
			ch <- prometheus.MustNewConstMetric(c.responseTime, prometheus.GaugeValue, *rt.Max/1000.0, "max")
		}
	}

	// OS metrics
	if status.Metrics.OS != nil {
		os := status.Metrics.OS
		if os.CPU != nil && os.CPU.ControlGroup != nil && os.CPU.ControlGroup.CPUPercent != nil {
			ch <- prometheus.MustNewConstMetric(c.osCPUPercent, prometheus.GaugeValue, *os.CPU.ControlGroup.CPUPercent)
		}
		if os.Load != nil {
			if os.Load.Load1m != nil {
				ch <- prometheus.MustNewConstMetric(c.osLoadAvg1m, prometheus.GaugeValue, *os.Load.Load1m)
			}
			if os.Load.Load5m != nil {
				ch <- prometheus.MustNewConstMetric(c.osLoadAvg5m, prometheus.GaugeValue, *os.Load.Load5m)
			}
			if os.Load.Load15m != nil {
				ch <- prometheus.MustNewConstMetric(c.osLoadAvg15m, prometheus.GaugeValue, *os.Load.Load15m)
			}
		}
		if os.Memory != nil {
			if os.Memory.TotalBytes != nil {
				ch <- prometheus.MustNewConstMetric(c.osMemTotal, prometheus.GaugeValue, float64(*os.Memory.TotalBytes))
			}
			if os.Memory.FreeBytes != nil {
				ch <- prometheus.MustNewConstMetric(c.osMemFree, prometheus.GaugeValue, float64(*os.Memory.FreeBytes))
			}
			if os.Memory.UsedBytes != nil {
				ch <- prometheus.MustNewConstMetric(c.osMemUsed, prometheus.GaugeValue, float64(*os.Memory.UsedBytes))
			}
		}
	}
}
