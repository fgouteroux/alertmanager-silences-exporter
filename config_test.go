package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestLoadConfig_No_Env_No_Config(t *testing.T) {
	os.Setenv("ALERTMANAGER_URL", "")
	_, err := loadConfig("no-config.yml")
	if err == nil {
		t.Error(err)
	}
}

func TestLoadConfig_Config_OK(t *testing.T) {
	confStr := `
---
alertmanager_url: "http://localhost:9093/"
`
	err := ioutil.WriteFile("test-config.yml", []byte(confStr), 0755)
	if err != nil {
		t.Error(err)
	}

	want := "http://localhost:9093/"
	conf, err := loadConfig("sample-config.yml")
	if err != nil {
		t.Error(err)
	}

	if conf.AlertmanagerURL != want {
		t.Errorf("want '%s' got '%s'", want, conf.AlertmanagerURL)
	}

	err = os.Remove("test-config.yml")
	if err != nil {
		t.Error(err)
	}
}

func TestNewAlertmanagerSilencesCollector_Env_OK(t *testing.T) {
	err := os.Setenv("ALERTMANAGER_URL", "http://localhost:9093/")
	if err != nil {
		t.Error(err)
	}

	conf, err := loadConfig("")
	if err != nil {
		t.Error(err)
	}

	asc := NewAlertmanagerSilencesCollector(conf, &AlertmanagerClient{})
	got := asc.Config.AlertmanagerURL
	want := "http://localhost:9093/"

	if got != want {
		t.Errorf("got '%s' want '%s'", got, want)
	}
}
