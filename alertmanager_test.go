package main

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/models"
)

func TestConstructURL_OK(t *testing.T) {
	ac := NewAlertManagerClient("http://localhost:9093/")
	var cases = []struct {
		paths []string
		want  string
	}{
		{[]string{"silences"}, "http://localhost:9093/api/v2/silences"},
		{[]string{"silences"}, "http://localhost:9093/api/v2/silences"},
		{[]string{"silences"}, "http://localhost:9093/api/v2/silences"},

		{[]string{"silence", "1234"}, "http://localhost:9093/api/v2/silence/1234"},
		{[]string{"silence", "1234"}, "http://localhost:9093/api/v2/silence/1234"},
	}

	for _, c := range cases {
		got, err := ac.constructURL(c.paths...)
		if err != nil {
			t.Errorf("unable to construct Alertmanager URL: '%s'", err.Error())
		}

		if got != c.want {
			t.Errorf("unexpected Alertmanager URL\ngot: '%s'\nwant: '%s'", got, c.want)
		}
	}
}

func TestAlertmanagerClient_listSilences(t *testing.T) {
	silence := `
[{
  "id": "7d8eb77e-00f9-4e0e-9f20-047695569296",
  "status": {
    "state": "pending"
  },
  "updatedAt": "2020-02-21T13:12:21.232Z",
  "comment": "Silence",
  "createdBy": "api",
  "endsAt": "2020-02-29T23:11:44.603Z",
  "matchers": [
    {
      "isRegex": false,
      "name": "job",
      "value": "FakeApp"
    }
  ],
  "startsAt": "2020-02-20T22:12:33.533Z"
}]
`
	resourcesHandler := func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(silence))
	}
	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(resourcesHandler))
	defer ts.Close()

	id := "7d8eb77e-00f9-4e0e-9f20-047695569296"
	state := "pending"
	updatedAt, _ := strfmt.ParseDateTime("2020-02-21T13:12:21.232Z")
	comment := "Silence"
	createdBy := "api"
	endsAt, _ := strfmt.ParseDateTime("2020-02-29T23:11:44.603Z")
	startsAt, _ := strfmt.ParseDateTime("2020-02-20T22:12:33.533Z")
	name := "job"
	value := "FakeApp"
	isRegex := false

	want := models.GettableSilences{{
		ID:        &id,
		Status:    &models.SilenceStatus{State: &state},
		UpdatedAt: &updatedAt,
		Silence: models.Silence{
			Comment:   &comment,
			CreatedBy: &createdBy,
			EndsAt:    &endsAt,
			Matchers: models.Matchers{
				&models.Matcher{Name: &name, Value: &value, IsRegex: &isRegex},
			},
			StartsAt: &startsAt,
		},
	}}

	ac := NewAlertManagerClient(ts.URL)
	got, err := ac.ListSilences()
	if err != nil {
		t.Errorf("unexpected error received: '%s'", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("ListSilences() didn't return expected results\ngot: '%v'\nwant: '%v'\n", got, want)
	}
}
