package main

import (
	"log"
	"net/http"
	"time"

	"github.com/cloudfoundry-community/go-cfclient"
	"github.com/pkg/errors"
)

const (
	// autoscaler can not operate safely on apps with less than this many instances
	MinInstancesLimit = 3
	// autoscaler interval
	Interval = 30 * time.Second
)

type autoscaler struct {
	rules  []Rule
	log    *log.Logger
	client Client
}

type Config struct {
	ApiUrl            string
	ApiUsername       string
	ApiPassword       string
	SkipSslValidation bool
	Rules             []Rule
	Logger            *log.Logger
}

func run(cfg Config) {
	cfg.Logger.Print("creating cf client")
	client, err := cfclient.NewClient(&cfclient.Config{
		ApiAddress:        cfg.ApiUrl,
		Username:          cfg.ApiUsername,
		Password:          cfg.ApiPassword,
		SkipSslValidation: cfg.SkipSslValidation,
		HttpClient:        &http.Client{Timeout: 10 * time.Second},
	})
	if err != nil {
		cfg.Logger.Fatal(errors.Wrap(err, "create cf client"))
	}

	cfg.Logger.Printf("validating autoscaler rules: %+v", cfg.Rules)
	err = validateRules(cfg.Rules)
	if err != nil {
		cfg.Logger.Fatal(errors.Wrap(err, "validate autoscaler rules"))
	}

	as := &autoscaler{client: &ApiClient{client}, rules: cfg.Rules, log: cfg.Logger}

	cfg.Logger.Print("starting autoscaler loop")
	for range time.Tick(Interval) {
		cfg.Logger.Print("starting autoscaler iteration")
		as.autoscaleApps()
	}
}

func (as *autoscaler) autoscaleApps() error {
	apps, err := as.client.GetApps()
	if err != nil {
		return errors.Wrap(err, "get app list")
	}

	for _, app := range apps {
		err := as.autoscaleApp(app)
		if err != nil {
			as.log.Print(errors.Wrapf(err, "autoscale app %v", app))
		}
	}

	return nil
}

func (as *autoscaler) autoscaleApp(app App) error {
	desired, err := as.analyzeApp(app)
	if err != nil {
		return errors.Wrap(err, "analyze app")
	}

	as.log.Printf("autoscale app %v: target %d instances", app, desired)

	if desired != app.Instances {
		if desired < MinInstancesLimit {
			// this should never happen
			return errors.Errorf("illegal to scale below %d instances", MinInstancesLimit)
		}
		err = as.client.Scale(app, desired)
		if err != nil {
			return errors.Wrap(err, "scale app")
		}
	}

	return nil
}

func (as *autoscaler) analyzeApp(app App) (desired int, err error) {
	rule, found := ruleFor(as.rules, app.App, app.Space, app.Org)
	if !found {
		err = errors.New("no applicable rule")
		return
	}

	switch {
	case app.Instances < rule.MinInstances || app.Instances > rule.MaxInstances:
		err = errors.New("number of instances outside of min/max bounds")
	case app.Instances != app.InstancesRunning:
		err = errors.Errorf("number of running instances differs from desired: %d/%d", app.InstancesRunning, app.Instances)
	case app.Instances < rule.MaxInstances && (rule.MaxCpu <= app.CpuAvg || rule.MaxMem <= app.MemAvg):
		desired = app.Instances + 1
	case app.Instances > rule.MinInstances && (rule.MinCpu >= app.CpuAvg && rule.MinMem >= app.MemAvg):
		desired = app.Instances - 1
	default:
		desired = app.Instances
	}
	return
}
