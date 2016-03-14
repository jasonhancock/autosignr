package autosignr

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type PuppetDBNode struct {
	Certname string
}

// Queries PuppetDB and returns the list of nodes found to be inactive, where
// inactive is the last report timestamp is greater than $hours old
func FindInactiveNodes(hours int, host string, protocol string, uri string, ignore_cert_errors bool) ([]string, error) {
	var list []string

	t := time.Now().Add(time.Hour * time.Duration(hours*-1)).Format(time.RFC3339)

	url := fmt.Sprintf("%s://%s%s?query=[\"<\",\"report_timestamp\",\"%s\"]",
		protocol,
		host,
		uri,
		t)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: ignore_cert_errors},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Printf("Error : %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return list, errors.New(fmt.Sprintf("Unable to download: %d", resp.StatusCode))
	}

	body, _ := ioutil.ReadAll(resp.Body)

	var l []PuppetDBNode

	if err := json.Unmarshal(body, &l); err != nil {
		return list, err
	}

	for _, val := range l {
		list = append(list, val.Certname)
	}

	return list, nil
}
