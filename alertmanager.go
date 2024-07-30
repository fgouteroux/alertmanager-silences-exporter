package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/prometheus/alertmanager/api/v2/models"
)

const (
	apiVersion = "/api/v2/"
)

type AlertmanagerAPI interface {
	ListSilences() (models.GettableSilences, error)
}

// AlertmanagerClient is the concrete implementation of the client object for methods calling the Alertmanager API
type AlertmanagerClient struct {
	URL      string
	Username string
	Password string
	TenantID string
}

// NewAlertManagerClient creates a client to work with
func NewAlertManagerClient(baseURL, username, password, tenantID string) *AlertmanagerClient {
	u := baseURL + "/" + apiVersion
	return &AlertmanagerClient{URL: u, Username: username, Password: password, TenantID: tenantID}
}

func (ac *AlertmanagerClient) constructURL(pairs ...string) (string, error) {
	u, err := url.Parse(ac.URL)
	if err != nil {
		return "", err
	}
	p := path.Join(pairs...)
	u.Path = path.Join(u.Path, p)

	return u.String(), nil
}

func (ac *AlertmanagerClient) doRequest(method, url string, requestBody io.Reader) ([]byte, error) {
	var client = &http.Client{}
	req, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return nil, fmt.Errorf("unable to create HTTP request: %s", err.Error())
	}

	if ac.Username != "" && ac.Password != "" {
		req.SetBasicAuth(ac.Username, ac.Password)
	}

	if ac.TenantID != "" {
		req.Header.Set("X-Scope-OrgID", ac.TenantID)
	}

	req.Header.Set("User-Agent", "alertmanager-silences-exporter")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("unable to get response: %s", err.Error())
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Alertmanager returned an HTTP error code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %s", err.Error())
	}
	return body, nil
}

func (ac *AlertmanagerClient) ListSilences() (models.GettableSilences, error) {
	var silences models.GettableSilences

	url, err := ac.constructURL("silences")
	if err != nil {
		return silences, err
	}

	body, err := ac.doRequest("GET", url, nil)
	if err != nil {
		return silences, fmt.Errorf("unable to create HTTP request: %s", err.Error())
	}

	err = json.Unmarshal(body, &silences)
	if err != nil {
		return silences, fmt.Errorf("unable to unmarshal body: %s", err.Error())
	}
	return silences, nil
}
