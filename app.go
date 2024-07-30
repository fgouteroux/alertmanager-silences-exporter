package main

import (
	"net/http"
	"os"
	"time"

	"github.com/go-kit/log/level"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
	"github.com/prometheus/exporter-toolkit/web"
	webflag "github.com/prometheus/exporter-toolkit/web/kingpinflag"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/alecthomas/kingpin/v2"
)

// App is responsible for managing each instance of config and clients.
type App struct {
	config *Config
	client AlertmanagerAPI
}

func main() {
	configFile := kingpin.Flag("config.file", "Path to config file.").Short('c').Default("config/config.yml").String()
	metricsPath := kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	webConfig := webflag.AddFlags(kingpin.CommandLine, ":9666")

	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("alertmanager-silences-exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	appConf, err := loadConfig(*configFile)
	if err != nil {
		level.Error(logger).Log("Failed to load config: %s\n", err.Error()) // #nosec G104
		os.Exit(1)
	}

	application := App{
		config: appConf,
		client: NewAlertManagerClient(
			appConf.AlertmanagerURL,
			appConf.AlertmanagerUsername,
			appConf.AlertmanagerPassword,
			"",
		),
	}

	collector := NewAlertmanagerSilencesCollector(application.config, application.client, logger)
	prometheus.MustRegister(collector)

	level.Info(logger).Log("msg", "alertmanager-silences-exporter", "version", version.Info()) // #nosec G104
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())    // #nosec G104

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Alertmanager Silences Exporter</title></head>
			<body>
			<h1>Alertmanager Silences Exporter</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	server := &http.Server{
		ReadTimeout:       120 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
	}

	if err := web.ListenAndServe(server, webConfig, logger); err != nil {
		level.Error(logger).Log("err", err) // #nosec G104
		os.Exit(1)
	}
}
