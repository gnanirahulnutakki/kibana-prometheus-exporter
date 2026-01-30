package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gnanirahulnutakki/kibana-prometheus-exporter/internal/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	version   = "dev"
	buildTime = "unknown"
	gitCommit = "unknown"
)

func main() {
	// Command line flags
	listenAddr := flag.String("listen-address", ":9684", "Address to listen on for metrics")
	metricsPath := flag.String("metrics-path", "/metrics", "Path under which to expose metrics")
	kibanaURL := flag.String("kibana-url", "http://localhost:5601", "Kibana URL to scrape")
	kibanaUsername := flag.String("kibana-username", "", "Username for Kibana basic auth (optional)")
	kibanaPassword := flag.String("kibana-password", "", "Password for Kibana basic auth (optional)")
	timeout := flag.Duration("timeout", 10*time.Second, "Timeout for Kibana API requests")
	insecureSkipVerify := flag.Bool("insecure-skip-verify", false, "Skip TLS certificate verification")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	logFormat := flag.String("log-format", "text", "Log format (text, json)")
	showVersion := flag.Bool("version", false, "Show version information")

	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("kibana-prometheus-exporter %s\n", version)
		fmt.Printf("  Build time: %s\n", buildTime)
		fmt.Printf("  Git commit: %s\n", gitCommit)
		os.Exit(0)
	}

	// Configure logging
	configureLogging(*logLevel, *logFormat)

	log.WithFields(log.Fields{
		"version":    version,
		"build_time": buildTime,
		"git_commit": gitCommit,
	}).Info("Starting Kibana Prometheus Exporter")

	// Override from environment variables if set
	if envURL := os.Getenv("KIBANA_URL"); envURL != "" {
		*kibanaURL = envURL
	}
	if envUser := os.Getenv("KIBANA_USERNAME"); envUser != "" {
		*kibanaUsername = envUser
	}
	if envPass := os.Getenv("KIBANA_PASSWORD"); envPass != "" {
		*kibanaPassword = envPass
	}

	log.WithField("kibana_url", *kibanaURL).Info("Configured Kibana endpoint")

	// Create collector
	kibanaCollector := collector.NewKibanaCollector(collector.Config{
		KibanaURL:          *kibanaURL,
		Username:           *kibanaUsername,
		Password:           *kibanaPassword,
		Timeout:            *timeout,
		InsecureSkipVerify: *insecureSkipVerify,
	})

	// Register collector
	prometheus.MustRegister(kibanaCollector)

	// HTTP handlers
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Kibana Prometheus Exporter</title></head>
			<body>
			<h1>Kibana Prometheus Exporter</h1>
			<p>Version: ` + version + `</p>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		// Check if we can reach Kibana
		if err := kibanaCollector.CheckHealth(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(fmt.Sprintf("NOT READY: %v", err)))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("READY"))
	})

	log.WithFields(log.Fields{
		"address":      *listenAddr,
		"metrics_path": *metricsPath,
	}).Info("Starting HTTP server")

	if err := http.ListenAndServe(*listenAddr, nil); err != nil {
		log.WithError(err).Fatal("Failed to start HTTP server")
	}
}

func configureLogging(level, format string) {
	// Set log level
	switch level {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	// Set log format
	if format == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}
}
