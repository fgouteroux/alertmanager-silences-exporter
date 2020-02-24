package main

import (
	"log"
	"os"
	"regexp"

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

func (s *Silence) Decorate() error {
	s.Labels = map[string]string{}
	s.Labels["id"] = *s.Gettable.ID
	s.Labels["comment"] = *s.Gettable.Comment
	s.Labels["createdBy"] = *s.Gettable.CreatedBy
	s.Labels["startsAt"] = s.Gettable.StartsAt.String()
	s.Labels["endsAt"] = s.Gettable.EndsAt.String()
	s.Labels["status"] = *s.Gettable.Status.State

	for _, m := range s.Gettable.Matchers {
		s.Labels[*m.Name] = *m.Value
	}

	s.Status = *s.Gettable.Status.State
	return nil
}

// AlertmanagerSilencesCollector collects Alertmanager Silence metrics
type AlertmanagerSilencesCollector struct {
	Config *Config
}

// NewAlertmanagerSilencesCollector returns the collector
func NewAlertmanagerSilencesCollector(conf *Config) *AlertmanagerSilencesCollector {
	return &AlertmanagerSilencesCollector{Config: conf}
}

// Describe to satisfy the collector interface.
func (c *AlertmanagerSilencesCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- prometheus.NewDesc("AlertmanagerSilencesCollector", "dummy", nil, nil)
}

// Collect metrics from Alertmanager
func (c *AlertmanagerSilencesCollector) Collect(ch chan<- prometheus.Metric) {
	ac := NewAlertManagerClient(c.Config.AlertmanagerURL)

	silences, err := ac.ListSilences()
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}

	for _, s := range silences {
		silence := &Silence{Gettable: s}

		err = silence.Decorate()
		if err != nil {
			log.Printf("Error exporting silence %v: %s\n", silence, err)
			continue
		}

		if silence.Status == "active" {
			c.extractMetric(ch, silence)
		}
	}
}

func (c *AlertmanagerSilencesCollector) extractMetric(ch chan<- prometheus.Metric, silence *Silence) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence", "Alertmanager silence extract", nil, silence.Labels),
		prometheus.GaugeValue,
		1,
	)
}
