package bcmr

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/golang/glog"
)

func Test_parseSignatureFromText(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "asdf",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}

}

func getRegistry(url string) (*Registry, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		glog.Errorf("Error creating a new request for %v: %v", url, err)
		return nil, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Invalid response status: " + string(resp.Status))
	}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data Registry
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		glog.Errorf("Error unmarshalling response from %s: %v", url, err)
		return nil, err
	}

	return &data, nil
}

func Test_download(t *testing.T) {
	registry, _ := getRegistry("https://bcmr.paytaca.com/api/registries/cade35f821c314c4f16de1f99484deb47e0320a688e2557a7f0a7d865371d695:1/")
	// println(err.Error())
	for _, identity := range *registry.Identities {
		for _, revision := range identity {
			for token_id, token := range revision.Token.Nfts.Parse.Types {
				println(token_id, " == ", token.Name)
			}
		}
	}
}
