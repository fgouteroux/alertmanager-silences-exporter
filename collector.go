package main

import (
	"os"
	"regexp"
	"time"

	"github.com/prometheus/common/log"

	"github.com/prometheus/alertmanager/api/v2/models"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	invalidMetricChars = regexp.MustCompile("[^a-zA-Z0-9:_]")
)

type Silence struct {
	Gettable *models.GettableSilence
	Labels   map[string]string
	Status   string
}

func (s *Silence) Decorate() {
	s.Labels = map[string]string{}
	s.Labels["id"] = *s.Gettable.ID
	s.Labels["comment"] = *s.Gettable.Comment
	s.Labels["createdBy"] = *s.Gettable.CreatedBy
	s.Labels["status"] = *s.Gettable.Status.State

	for _, m := range s.Gettable.Matchers {
		s.Labels["matcher_"+*m.Name] = *m.Value
	}

	s.Status = *s.Gettable.Status.State
}

// AlertmanagerSilencesCollector collects Alertmanager Silence metrics
type AlertmanagerSilencesCollector struct {
	Config             *Config
	AlertmanagerClient AlertmanagerAPI
}

// NewAlertmanagerSilencesCollector returns the collector
func NewAlertmanagerSilencesCollector(conf *Config, client AlertmanagerAPI) *AlertmanagerSilencesCollector {
	return &AlertmanagerSilencesCollector{Config: conf, AlertmanagerClient: client}
}

// Describe to satisfy the collector interface.
func (c *AlertmanagerSilencesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("AlertmanagerSilencesCollector", "dummy", nil, nil)
}

// Collect metrics from Alertmanager
func (c *AlertmanagerSilencesCollector) Collect(ch chan<- prometheus.Metric) {
	silences, err := c.AlertmanagerClient.ListSilences()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for _, s := range silences {
		silence := &Silence{Gettable: s}

		silence.Decorate()
		c.extractMetric(ch, silence)
	}
}

func (c *AlertmanagerSilencesCollector) extractMetric(ch chan<- prometheus.Metric, silence *Silence) {
	startTime, err := time.Parse(time.RFC3339, silence.Gettable.StartsAt.String())
	if err != nil {
		log.Printf("cannot parse start time of silence with ID '%s'\n", silence.Labels["id"])
		return
	}

	endTime, err := time.Parse(time.RFC3339, silence.Gettable.EndsAt.String())
	if err != nil {
		log.Printf("cannot parse end time of silence with ID '%s'\n", silence.Labels["id"])
		return
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_info", "Alertmanager silence info metric", nil, silence.Labels),
		prometheus.GaugeValue,
		1,
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_start_seconds", "Alertmanager silence start time, elapsed seconds since epoch", nil, map[string]string{"id": silence.Labels["id"]}),
		prometheus.GaugeValue,
		float64(startTime.Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_end_seconds", "Alertmanager silence end time, elapsed seconds since epoch", nil, map[string]string{"id": silence.Labels["id"]}),
		prometheus.GaugeValue,
		float64(endTime.Unix()),
	)
}
