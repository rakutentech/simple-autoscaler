package main

import (
	"bytes"
	"fmt"
	"io/ioutil"

	cfclient "github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
)

type Client interface {
	GetApps() (Apps, error)
	Scale(app App, desired int) error
}

type App struct {
	Guid             string
	App              string
	Space            string
	Org              string
	Instances        int
	InstancesRunning int
	CpuAvg           int
	MemAvg           int
}

type Apps map[string]App

type ApiClient struct {
	Client *cfclient.Client
}

func (c *ApiClient) GetApps() (Apps, error) {
	apps, err := c.Client.ListApps()
	if err != nil {
		return nil, errors.Wrap(err, "get app list")
	}

	r := make(Apps, len(apps))

	for _, app := range apps {
		space, err := app.Space()
		if err != nil {
			return nil, errors.Wrapf(err, "get app %s space", app.Guid)
		}

		org, err := space.Org()
		if err != nil {
			return nil, errors.Wrapf(err, "get app %s org", app.Guid)
		}

		started := app.State == "STARTED"

		var instances map[string]cfclient.AppStats
		if started {
			instances, err = c.Client.GetAppStats(app.Guid)
			if err != nil {
				return nil, errors.Wrapf(err, "get app %s stats", app.Guid)
			}
		}

		r[app.Guid] = processApp(app.Guid, app.Name, space.Name, org.Name, started, app.Instances, instances)
	}

	return r, nil
}

func processApp(guid, app, space, org string, started bool, desired int, instances map[string]cfclient.AppStats) App {
	a := App{Guid: guid, App: app, Space: space, Org: org}

	if started {
		var cpu, mem float64

		for _, instance := range instances {
			if instance.State != "RUNNING" || instance.Stats.Usage.CPU < 0 || instance.Stats.Usage.CPU > 1 || instance.Stats.Usage.Mem < 0 || instance.Stats.Usage.Mem > instance.Stats.MemQuota {
				// if anything seems suspicious, we skip this instance; autoscaleApp
				// will refuse to scale the app if instances are missing
				continue
			}
			a.InstancesRunning += 1
			cpu += instance.Stats.Usage.CPU
			mem += float64(instance.Stats.Usage.Mem) / float64(instance.Stats.MemQuota)
		}

		if a.InstancesRunning > 0 {
			// FIXME: golang fail: there's no round function so do it manually by adding +0.5
			// this should be fine because we only treat non-negative, normal numbers
			a.CpuAvg = int(cpu/float64(a.InstancesRunning)*100.0 + 0.5)
			a.MemAvg = int(mem/float64(a.InstancesRunning)*100.0 + 0.5)
		}
		a.Instances = desired
	}

	return a
}

func (c *ApiClient) Scale(app App, desired int) error {
	requestURL := fmt.Sprintf("/v2/apps/%s?async=true", app.Guid)
	body := bytes.NewBufferString(fmt.Sprintf(`{"instances":%d}`, desired))
	req := c.Client.NewRequestWithBody("PUT", requestURL, body)
	resp, err := c.Client.DoRequest(req)
	if err != nil {
		return errors.Wrap(err, "sending scaling request")
	}
	defer resp.Body.Close()

	msg, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "reading scaling response")
	}
	if resp.StatusCode != 201 {
		return errors.Errorf("scaling request rejected: %d %s", resp.StatusCode, string(msg))
	}

	return nil
}
