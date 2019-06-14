package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/minus5/svckit/dcy"
	"github.com/minus5/svckit/log"
)

const (
	// BaseURL is base url of cloud atlas
	BaseURL = "https://cloud.mongodb.com/api/atlas/v1.0"
	// GroupID is id of project
	GroupID = "5bd6fc3a014b76fe61ca7291"
	// DiskUtilizationAlertID is id of disk utilization alert
	DiskUtilizationAlertID = "5bd6fc3a014b76fe61ca7296"
	// ScannedObjectsAlertID is id of query targeting: scanned objects alert
	ScannedObjectsAlertID = "5bd6fc3a014b76fe61ca7297"
)

// ToggleMaintenanceAlerts turns all alerts which can give false positives during maintenence window on or off
func ToggleMaintenanceAlerts(enabled bool) {
	if err := toggleAlert(enabled, DiskUtilizationAlertID); err != nil {
		log.Error(err)
	}

	if err := toggleAlert(enabled, ScannedObjectsAlertID); err != nil {
		log.Error(err)
	}

}

func toggleAlert(enabled bool, alertID string) error {

	p := ToggleAlert{Enabled: enabled}
	b, _ := json.Marshal(p)

	_, err := apiRequest("PATCH", fmt.Sprintf("groups/%s/alertConfigs/%s", GroupID, alertID), b)
	return err
}

func apiRequest(method, uri string, body []byte) ([]byte, error) {
	kvs, err := dcy.KVs("mongo/api")
	if err != nil {
		return nil, err
	}

	r := NewDigestRequest(kvs["username"], kvs["key"])

	req, _ := http.NewRequest(method, fmt.Sprintf("%s/%s", BaseURL, uri), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := r.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)

	return b, err
}

// ToggleAlert is struct used for turning alert off or on
type ToggleAlert struct {
	Enabled bool `json:"enabled"`
}
