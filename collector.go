package main

import (
	"time"

	"github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
)

type Silence struct {
	Gettable *models.GettableSilence
	Labels   map[string]string
	Status   string
}

func (s *Silence) Decorate(tenant string) {
	s.Labels = map[string]string{}
	s.Labels["id"] = *s.Gettable.ID
	s.Labels["comment"] = *s.Gettable.Comment
	s.Labels["createdBy"] = *s.Gettable.CreatedBy
	s.Labels["status"] = *s.Gettable.Status.State

	if tenant != "" {
		s.Labels["tenant"] = tenant
	}

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
	if len(c.Config.Tenants) == 0 {
		silences, err := c.AlertmanagerClient.ListSilences()
		if err != nil {
			log.Errorf("unable to list silences: %s", err.Error())
			ch <- prometheus.NewInvalidMetric(amErrorDesc, err)
			return
		}

		for _, s := range silences {
			silence := &Silence{Gettable: s}
			silence.Decorate("")

			if !c.Config.ExpiredSilences {
				if silence.Status != "active" {
					continue
				}
			}

			c.extractMetric(ch, silence, "")
		}
	} else {
		for _, tenant := range c.Config.Tenants {

			client := NewAlertManagerClient(
				c.Config.AlertmanagerURL,
				c.Config.AlertmanagerUsername,
				c.Config.AlertmanagerPassword,
				tenant,
			)
			silences, err := client.ListSilences()
			if err != nil {
				log.Errorf("unable to list silences: %s", err.Error())
				ch <- prometheus.NewInvalidMetric(amErrorDesc, err)
				return
			}

			for _, s := range silences {
				silence := &Silence{Gettable: s}
				silence.Decorate(tenant)

				if !c.Config.ExpiredSilences {
					if silence.Status != "active" {
						continue
					}
				}

				c.extractMetric(ch, silence, tenant)
			}
		}
	}
}

func (c *AlertmanagerSilencesCollector) extractMetric(ch chan<- prometheus.Metric, silence *Silence, tenant string) {
	startTime, err := time.Parse(time.RFC3339, silence.Gettable.StartsAt.String())
	if err != nil {
		log.Errorf("cannot parse start time of silence with ID '%s'\n", silence.Labels["id"])
		return
	}

	endTime, err := time.Parse(time.RFC3339, silence.Gettable.EndsAt.String())
	if err != nil {
		log.Errorf("cannot parse end time of silence with ID '%s'\n", silence.Labels["id"])
		return
	}

	state := 0
	if silence.Status == "active" {
		state = 1
	}

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_info", "Alertmanager silence info metric", nil, silence.Labels),
		prometheus.GaugeValue,
		float64(state),
	)

	labels := map[string]string{"id": silence.Labels["id"]}
	if tenant != "" {
		labels["tenant"] = tenant
	}
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_start_seconds", "Alertmanager silence start time, elapsed seconds since epoch", nil, labels),
		prometheus.GaugeValue,
		float64(startTime.Unix()),
	)

	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc("alertmanager_silence_end_seconds", "Alertmanager silence end time, elapsed seconds since epoch", nil, labels),
		prometheus.GaugeValue,
		float64(endTime.Unix()),
	)
}
