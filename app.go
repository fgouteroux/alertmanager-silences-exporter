package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	configFile   = kingpin.Flag("config.file", "Path to config file.").Short('c').Required().String()
	genericError = 1

	router *mux.Router
)

// App is responsible for managing each instance of config and clients.
type App struct {
	config *Config
	client AlertmanagerAPI
}

func (a *App) listMetrics(w http.ResponseWriter, r *http.Request) {
	registry := prometheus.NewRegistry()
	collector := NewAlertmanagerSilencesCollector(a.config, a.client)
	registry.MustRegister(collector)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	h.ServeHTTP(w, r)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`<html>
            <head>
            <title>Alertmanager Silences Exporter</title>
            </head>
            <body>
            <h1>Alertmanager Silences Exporter</h1>
						<p><a href="/metrics">Metrics</a></p>
            </body>
            </html>`))
}

func main() {
	kingpin.Parse()

	appConf, err := loadConfig(*configFile)
	if err != nil {
		log.Printf("error loading config: %s\n", err.Error())
		os.Exit(genericError)
	}

	application := App{
		config: appConf,
		client: NewAlertManagerClient(appConf.AlertmanagerURL),
	}

	router = mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/metrics", application.listMetrics).Methods("GET").Name("listMetrics")
	router.HandleFunc("/", indexHandler).Name("indexHandler")
	http.Handle("/", router)

	log.Printf("alertmanager-silences-exporter listening on port %d", 9666)
	if err := http.ListenAndServe("0:9666", nil); err != nil {
		log.Fatalf("Error starting HTTP server: %v", err)
	}
}
