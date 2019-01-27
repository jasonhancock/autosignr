package autosignr

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
)

// PuppetDBNode - Returning from the API call to puppetDB
type PuppetDBNode struct {
	Certname string `json:"certname"`
}

// FindInactiveNodes Queries PuppetDB and returns the list of nodes found to be inactive, where
// inactive is the last report timestamp is greater than $hours old
func FindInactiveNodes(hours int, host string, protocol string, uri string, ignoreCertErrors bool, includeFilters []string) ([]string, error) {
	t := time.Now().Add(time.Hour * time.Duration(hours*-1)).Format(time.RFC3339)

	url := fmt.Sprintf("%s://%s%s", protocol, host, uri)

	data := fmt.Sprintf("{ \"query\": \"nodes[certname]{ report_timestamp < \\\"%s\\\"", t)
	for _, val := range includeFilters {
		data = data + " " + val
	}
	data = data + " }\"}"

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignoreCertErrors},
	}
	client := &http.Client{Transport: tr}

	var list []string
	resp, err := client.Post(url, "application/json", strings.NewReader(data))
	if err != nil {
		return list, errors.Wrap(err, "Post Error: ")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return list, errors.New(fmt.Sprintf("Unable to download: %d", resp.StatusCode))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var l []PuppetDBNode
	if err := json.Unmarshal(body, &l); err != nil {
		return list, errors.Wrap(err, "Unmarshal Error: ")
	}

	for _, val := range l {
		list = append(list, val.Certname)
	}

	return list, nil
}
