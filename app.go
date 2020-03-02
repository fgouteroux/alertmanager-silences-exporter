package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile    = kingpin.Flag("config.file", "Path to config file.").Short('c').Required().String()
	listenAddress = kingpin.Flag("web.listen-address", "The address to listen on for HTTP requests.").Default(":9666").String()
	metricsPath   = kingpin.Flag("web.telemetry-path", "Path under which to expose metrics.").Default("/metrics").String()
	genericError  = 1
	amErrorDesc   = prometheus.NewDesc("alertmanager_error", "Error collecting metrics", nil, nil)

	router *mux.Router
)

// App is responsible for managing each instance of config and clients.
type App struct {
	config *Config
	client AlertmanagerAPI
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
            <head>
            <title>Alertmanager Silences Exporter</title>
            </head>
            <body>
            <h1>Alertmanager Silences Exporter</h1>
						<p><a href="` + *metricsPath + `">Metrics</a></p>
            </body>
            </html>`))
}

func main() {
	kingpin.Parse()

	appConf, err := loadConfig(*configFile)
	if err != nil {
		log.Fatalf("error loading config: %s\n", err.Error())
		os.Exit(genericError)
	}

	application := App{
		config: appConf,
		client: NewAlertManagerClient(appConf.AlertmanagerURL),
	}

	router = mux.NewRouter().StrictSlash(true)

	collector := NewAlertmanagerSilencesCollector(application.config, application.client)
	prometheus.MustRegister(collector)

	router.Handle(*metricsPath, promhttp.Handler())
	router.HandleFunc("/", indexHandler).Name("indexHandler")
	http.Handle("/", router)

	log.Infof("alertmanager-silences-exporter listening on port %d", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
