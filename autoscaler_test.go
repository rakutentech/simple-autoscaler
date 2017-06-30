package main

import (
	"bytes"
	"errors"
	"log"
	"testing"
)

type appTest struct {
	App          *App
	AppsError    error
	ScaleApp     *App
	ScaleDesired *int
}

type test struct {
	rules []Rule
	apps  []appTest
}

var guid = "01234567-89ab-cdef-0123-456789abcdef"
var someError = errors.New("error")

func TestAutoscaler(t *testing.T) {
	tests := []test{
		{
			rules: []Rule{
				Rule{App: "a", Space: "s", Org: "o", MinInstances: 5, MaxInstances: 10, MinCpu: 40, MaxCpu: 60, MinMem: 50, MaxMem: 70},
			},
			apps: []appTest{
				// no rule
				appTest{App: &App{App: "x", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}},
				// GetApps error
				appTest{AppsError: someError},
				appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}, AppsError: someError},
				// low load
				asoNoScale(11, 0, 0, nil),
				asoScale(10, 0, 0, 9),
				asoScale(9, 0, 0, 8),
				asoScale(6, 0, 0, 5),
				asoNoScale(5, 0, 0, nil),
				asoNoScale(4, 0, 0, nil),
				// hi load
				asoNoScale(11, 100, 100, nil),
				asoNoScale(10, 100, 100, nil),
				asoScale(9, 100, 100, 10),
				asoScale(6, 100, 100, 7),
				asoScale(5, 100, 100, 6),
				asoNoScale(4, 100, 100, nil),
				// cpu hi load, mem low load
				asoNoScale(11, 100, 0, nil),
				asoNoScale(10, 100, 0, nil),
				asoScale(9, 100, 0, 10),
				asoScale(6, 100, 0, 7),
				asoScale(5, 100, 0, 6),
				asoNoScale(4, 100, 0, nil),
				// cpu low load, mem hi load
				asoNoScale(11, 0, 100, nil),
				asoNoScale(10, 0, 100, nil),
				asoScale(9, 0, 100, 10),
				asoScale(6, 0, 100, 7),
				asoScale(5, 0, 100, 6),
				asoNoScale(4, 0, 100, nil),
				// cpu ok load, mem low load
				asoNoScale(11, 50, 0, nil),
				asoNoScale(10, 50, 0, nil),
				asoNoScale(9, 50, 0, nil),
				asoNoScale(6, 50, 0, nil),
				asoNoScale(5, 50, 0, nil),
				asoNoScale(4, 50, 0, nil),
				// cpu low load, mem ok load
				asoNoScale(11, 0, 60, nil),
				asoNoScale(10, 0, 60, nil),
				asoNoScale(9, 0, 60, nil),
				asoNoScale(6, 0, 60, nil),
				asoNoScale(5, 0, 60, nil),
				asoNoScale(4, 0, 60, nil),
				// cpu ok load, mem ok load
				asoNoScale(11, 50, 60, nil),
				asoNoScale(10, 50, 60, nil),
				asoNoScale(9, 50, 60, nil),
				asoNoScale(6, 50, 60, nil),
				asoNoScale(5, 50, 60, nil),
				asoNoScale(4, 50, 60, nil),
				// cpu ok load, mem hi loadCpuAvg: 50,
				asoNoScale(11, 50, 100, nil),
				asoNoScale(10, 50, 100, nil),
				asoScale(9, 50, 100, 10),
				asoScale(6, 50, 100, 7),
				asoScale(5, 50, 100, 6),
				asoNoScale(4, 50, 100, nil),
				// cpu hi load, mem ok load
				asoNoScale(11, 100, 60, nil),
				asoNoScale(10, 100, 60, nil),
				asoScale(9, 100, 60, 10),
				asoScale(6, 100, 60, 7),
				asoScale(5, 100, 60, 6),
				asoNoScale(4, 100, 60, nil),
			},
		},
		{
			rules: []Rule{
				Rule{App: "a", Space: "s", Org: "o", MinInstances: 5, MaxInstances: 10, MinCpu: 40, MaxCpu: 60},
			},
			apps: []appTest{
				// no rule
				appTest{App: &App{App: "x", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}},
				// GetApps error
				appTest{AppsError: someError},
				appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}, AppsError: someError},
				// low load
				asoNoScale(11, 0, 0, nil),
				asoScale(10, 0, 0, 9),
				asoScale(9, 0, 0, 8),
				asoScale(6, 0, 0, 5),
				asoNoScale(5, 0, 0, nil),
				asoNoScale(4, 0, 0, nil),
				// hi load
				asoNoScale(11, 100, 100, nil),
				asoNoScale(10, 100, 100, nil),
				asoScale(9, 100, 100, 10),
				asoScale(6, 100, 100, 7),
				asoScale(5, 100, 100, 6),
				asoNoScale(4, 100, 100, nil),
				// cpu hi load, mem low load
				asoNoScale(11, 100, 0, nil),
				asoNoScale(10, 100, 0, nil),
				asoScale(9, 100, 0, 10),
				asoScale(6, 100, 0, 7),
				asoScale(5, 100, 0, 6),
				asoNoScale(4, 100, 0, nil),
				// cpu low load, mem hi load
				asoNoScale(11, 0, 100, nil),
				asoScale(10, 0, 100, 9),
				asoScale(9, 0, 100, 8),
				asoScale(6, 0, 100, 5),
				asoNoScale(5, 0, 100, nil),
				asoNoScale(4, 0, 100, nil),
				// cpu ok load, mem low load
				asoNoScale(11, 50, 0, nil),
				asoNoScale(10, 50, 0, nil),
				asoNoScale(9, 50, 0, nil),
				asoNoScale(6, 50, 0, nil),
				asoNoScale(5, 50, 0, nil),
				asoNoScale(4, 50, 0, nil),
				// cpu low load, mem ok load
				asoNoScale(11, 0, 60, nil),
				asoScale(10, 0, 60, 9),
				asoScale(9, 0, 60, 8),
				asoScale(6, 0, 60, 5),
				asoNoScale(5, 0, 60, nil),
				asoNoScale(4, 0, 60, nil),
				// cpu ok load, mem ok load
				asoNoScale(11, 50, 60, nil),
				asoNoScale(10, 50, 60, nil),
				asoNoScale(9, 50, 60, nil),
				asoNoScale(6, 50, 60, nil),
				asoNoScale(5, 50, 60, nil),
				asoNoScale(4, 50, 60, nil),
				// cpu ok load, mem hi load
				asoNoScale(11, 50, 100, nil),
				asoNoScale(10, 50, 100, nil),
				asoNoScale(9, 50, 100, nil),
				asoNoScale(6, 50, 100, nil),
				asoNoScale(5, 50, 100, nil),
				asoNoScale(4, 50, 100, nil),
				// cpu hi load, mem ok load
				asoNoScale(11, 100, 60, nil),
				asoNoScale(10, 100, 60, nil),
				asoScale(9, 100, 60, 10),
				asoScale(6, 100, 60, 7),
				asoScale(5, 100, 60, 6),
				asoNoScale(4, 100, 60, nil),
			},
		}, {
			rules: []Rule{
				Rule{App: "a", Space: "s", Org: "o", MinInstances: 5, MaxInstances: 10, MinMem: 50, MaxMem: 70},
			},
			apps: []appTest{
				// no rule
				appTest{App: &App{App: "x", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}},
				// GetApps error
				appTest{AppsError: someError},
				appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}, AppsError: someError},
				// low load
				asoNoScale(11, 0, 0, nil),
				asoScale(10, 0, 0, 9),
				asoScale(9, 0, 0, 8),
				asoScale(6, 0, 0, 5),
				asoNoScale(5, 0, 0, nil),
				asoNoScale(4, 0, 0, nil),
				// hi load
				asoNoScale(11, 100, 100, nil),
				asoNoScale(10, 100, 100, nil),
				asoScale(9, 100, 100, 10),
				asoScale(6, 100, 100, 7),
				asoScale(5, 100, 100, 6),
				asoNoScale(4, 100, 100, nil),
				// cpu hi load, mem low load
				asoNoScale(11, 100, 0, nil),
				asoScale(10, 100, 0, 9),
				asoScale(9, 100, 0, 8),
				asoScale(6, 100, 0, 5),
				asoNoScale(5, 100, 0, nil),
				asoNoScale(4, 100, 0, nil),
				// cpu low load, mem hi load
				asoNoScale(11, 0, 100, nil),
				asoNoScale(10, 0, 100, nil),
				asoScale(9, 0, 100, 10),
				asoScale(6, 0, 100, 7),
				asoScale(5, 0, 100, 6),
				asoNoScale(4, 0, 100, nil),
				// cpu ok load, mem low load
				asoNoScale(11, 50, 0, nil),
				asoScale(10, 50, 0, 9),
				asoScale(9, 50, 0, 8),
				asoScale(6, 50, 0, 5),
				asoNoScale(5, 50, 0, nil),
				asoNoScale(4, 50, 0, nil),
				// cpu low load, mem ok load
				asoNoScale(11, 0, 60, nil),
				asoNoScale(10, 0, 60, nil),
				asoNoScale(9, 0, 60, nil),
				asoNoScale(6, 0, 60, nil),
				asoNoScale(5, 0, 60, nil),
				asoNoScale(4, 0, 60, nil),
				// cpu ok load, mem ok load
				asoNoScale(11, 50, 60, nil),
				asoNoScale(10, 50, 60, nil),
				asoNoScale(9, 50, 60, nil),
				asoNoScale(6, 50, 60, nil),
				asoNoScale(5, 50, 60, nil),
				asoNoScale(4, 50, 60, nil),
				// cpu ok load, mem hi load
				asoNoScale(11, 50, 100, nil),
				asoNoScale(10, 50, 100, nil),
				asoScale(9, 50, 100, 10),
				asoScale(6, 50, 100, 7),
				asoScale(5, 50, 100, 6),
				asoNoScale(4, 50, 100, nil),
				// cpu hi load, mem ok load
				asoNoScale(11, 100, 60, nil),
				asoNoScale(10, 100, 60, nil),
				asoNoScale(9, 100, 60, nil),
				asoNoScale(6, 100, 60, nil),
				asoNoScale(5, 100, 60, nil),
				asoNoScale(4, 100, 60, nil),
			},
		},
		{
			rules: []Rule{
				Rule{App: "a", Space: "s", Org: "o", MinInstances: 5, MaxInstances: 10, MinCpu: 50, MaxCpu: 70},
				Rule{App: "b", Space: "s", Org: "o", MinInstances: 15, MaxInstances: 20, MinCpu: 30, MaxCpu: 50},
				Rule{App: "c", Space: "s", Org: "o", MinInstances: 25, MaxInstances: 30, MinCpu: 10, MaxCpu: 30},
			},
			apps: []appTest{
				// no rule
				appTest{App: &App{App: "x", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7}},

				// different apps
				appTest{
					App:          &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7},
					ScaleApp:     &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7},
					ScaleDesired: func(x int) *int { return &x }(6),
				},
				appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 17, InstancesRunning: 17}},
				appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: 7, InstancesRunning: 7, CpuAvg: 60}},

				appTest{
					App:          &App{App: "b", Space: "s", Org: "o", Guid: guid, Instances: 17, InstancesRunning: 17},
					ScaleApp:     &App{App: "b", Space: "s", Org: "o", Guid: guid, Instances: 17, InstancesRunning: 17},
					ScaleDesired: func(x int) *int { return &x }(16),
				},
				appTest{App: &App{App: "b", Space: "s", Org: "o", Guid: guid, Instances: 27, InstancesRunning: 27}},
				appTest{App: &App{App: "b", Space: "s", Org: "o", Guid: guid, Instances: 17, InstancesRunning: 17, CpuAvg: 40}},

				appTest{
					App:          &App{App: "c", Space: "s", Org: "o", Guid: guid, Instances: 27, InstancesRunning: 27},
					ScaleApp:     &App{App: "c", Space: "s", Org: "o", Guid: guid, Instances: 27, InstancesRunning: 27},
					ScaleDesired: func(x int) *int { return &x }(26),
				},
				appTest{App: &App{App: "c", Space: "s", Org: "o", Guid: guid, Instances: 37, InstancesRunning: 37}},
				appTest{App: &App{App: "c", Space: "s", Org: "o", Guid: guid, Instances: 27, InstancesRunning: 27, CpuAvg: 20}},
			},
		},
	}

	for i, test := range tests {
		if err := validateRules(test.rules); err != nil {
			t.Fatalf("%d: validateRules: %s", i, err)
		}

		for j, app := range test.apps {
			apps := make(Apps, 1)
			if app.App != nil {
				apps[app.App.Guid] = *app.App
			}
			mock := &MockClient{Apps: apps, AppsError: app.AppsError}

			buf := &bytes.Buffer{}
			logger := log.New(buf, "", log.Lshortfile)

			as := &autoscaler{client: mock, rules: test.rules, log: logger}
			as.autoscaleApps()

			if app.ScaleApp == nil {
				if mock.ScaleApp != nil {
					t.Fatalf("%d/%d: Scale called: %v %d\n%s", i, j, *mock.ScaleApp, *mock.ScaleDesired, buf.String())
				}
			} else {
				if mock.ScaleApp == nil {
					t.Fatalf("%d/%d: Scale not called\n%s", i, j, buf.String())
				} else if *mock.ScaleApp != *app.ScaleApp || *mock.ScaleDesired != *app.ScaleDesired {
					t.Fatalf("%d/%d: Scale called with wrong args: %v %d\n%s", i, j, *mock.ScaleApp, *mock.ScaleDesired, buf.String())
				}
			}
		}
	}
}

func asoNoScale(i, c, m int, err error) appTest {
	return appTest{App: &App{App: "a", Space: "s", Org: "o", Guid: guid, Instances: i, InstancesRunning: i, CpuAvg: c, MemAvg: m}, AppsError: err}
}

func asoScale(i, c, m, d int) appTest {
	t := asoNoScale(i, c, m, nil)
	t.ScaleApp, t.ScaleDesired = &*t.App, &d
	return t
}
