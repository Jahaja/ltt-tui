package tui

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/Jahaja/ltt"
)

type LTTClient struct {
	URI string
}

func (c *LTTClient) Start() error {
	resp, err := http.Get(fmt.Sprintf("%s/start", c.URI))
	if err != nil {
		return fmt.Errorf("failed to start ttl: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status code returned: %d", resp.StatusCode)
	}

	return nil
}

func (c *LTTClient) SetNumUsers(numUsers int) error {
	resp, err := http.Get(fmt.Sprintf("%s/set-num-users?num-users=%d", c.URI, numUsers))
	if err != nil {
		return fmt.Errorf("failed to set user num: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status code returned: %d", resp.StatusCode)
	}

	return nil
}

func (c *LTTClient) Stop() error {
	resp, err := http.Get(fmt.Sprintf("%s/stop", c.URI))
	if err != nil {
		return fmt.Errorf("failed to stop ttl: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status code returned: %d", resp.StatusCode)
	}

	return nil
}

func (c *LTTClient) Reset() error {
	resp, err := http.Get(fmt.Sprintf("%s/reset", c.URI))
	if err != nil {
		return fmt.Errorf("failed to reset ttl: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status code returned: %d", resp.StatusCode)
	}

	return nil
}

func (c *LTTClient) GetLoadTestInfo() (*ltt.LoadTest, error) {
	resp, err := http.Get(c.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch load test info: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	lt := &ltt.LoadTest{
		Stats: ltt.NewStatistics(),
	}

	err = json.Unmarshal(body, lt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse load test info from ttl: %w", err)
	}

	return lt, nil
}

func NewLTTClient(uri string) *LTTClient {
	return &LTTClient{URI: strings.TrimRight(uri, "/")}
}
