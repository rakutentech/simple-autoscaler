package main

import (
	"encoding/json"
	"fmt"
	"testing"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
)

type MockClient struct {
	Apps      Apps
	AppsError error

	ScaleApp     *App
	ScaleDesired *int
	ScaleError   error
}

func (c *MockClient) GetApps() (Apps, error) {
	return c.Apps, c.AppsError
}

func (c *MockClient) Scale(app App, desired int) error {
	c.ScaleApp = &app
	c.ScaleDesired = &desired
	return c.ScaleError
}

func IS(cpuPct, memPct float64) (a cfclient.AppStats) {
	memQuota := 1024 * 1024 * 1024
	s := fmt.Sprintf(`{"state":"RUNNING","stats":{"usage":{"cpu":%f,"mem":%d},"mem_quota":%d}}`, cpuPct, int(memPct*float64(memQuota)), memQuota)
	json.Unmarshal([]byte(s), &a)
	return a
}

func TestProcessApps(t *testing.T) {
	a := processApp(guid, "a", "s", "o", true, 1, map[string]cfclient.AppStats{"0": IS(0.5, 0.75)})
	if a.CpuAvg != 50 || a.MemAvg != 75 || a.Instances != 1 || a.InstancesRunning != 1 {
		t.Fatalf("processApp fail: %+v", a)
	}

	a = processApp(guid, "a", "s", "o", true, 2, map[string]cfclient.AppStats{"0": IS(0.5, 0.8), "1": IS(0.3, 0.6)})
	if a.CpuAvg != 40 || a.MemAvg != 70 || a.Instances != 2 || a.InstancesRunning != 2 {
		t.Fatalf("processApp fail: %+v", a)
	}

	a = processApp(guid, "a", "s", "o", true, 2, map[string]cfclient.AppStats{"0": IS(0.5, 0.8), "1": {}})
	if a.CpuAvg != 50 || a.MemAvg != 80 || a.Instances != 2 || a.InstancesRunning != 1 {
		t.Fatalf("processApp fail: %+v", a)
	}
}
