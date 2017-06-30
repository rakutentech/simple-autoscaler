package main

import (
	"math"

	"github.com/pkg/errors"
)

type Rule struct {
	App          string `json:"app"`
	Space        string `json:"space"`
	Org          string `json:"org"`
	MinInstances int    `json:"min_instances"`
	MaxInstances int    `json:"max_instances"`
	MinCpu       int    `json:"scale_in_cpu"`
	MaxCpu       int    `json:"scale_out_cpu"`
	MinMem       int    `json:"scale_in_mem"`
	MaxMem       int    `json:"scale_out_mem"`
}

func validateRules(rules []Rule) error {
	for idx, rule := range rules {
		if rule, err := validateRule(rule); err != nil {
			return errors.Wrapf(err, "rule %d", idx)
		} else {
			rules[idx] = rule
		}
	}
	return nil
}

func validateRule(rule Rule) (Rule, error) {
	switch {
	case rule.App == "":
		return rule, errors.New("no app specified")
	case rule.Space == "":
		return rule, errors.New("no space specified")
	case rule.Org == "":
		return rule, errors.New("no org specified")
	case rule.MinInstances < MinInstancesLimit:
		return rule, errors.Errorf("minimum instances should be >= %d", MinInstancesLimit)
	case rule.MaxInstances <= rule.MinInstances:
		return rule, errors.New("maximum instances should be more than minimum instances")
	case rule.MaxCpu < 0 || rule.MaxCpu > 100:
		return rule, errors.New("max cpu threshold should be in the range 0<=t<=100")
	case rule.MinCpu < 0 || rule.MinCpu > 100:
		return rule, errors.New("min cpu threshold should be in the range 0<=t<=100")
	case rule.MinCpu >= rule.MaxCpu && !(rule.MinCpu == 0 && rule.MaxCpu == 0):
		return rule, errors.New("min cpu threshold should be less than max cpu threshold")
	case rule.MaxMem < 0 || rule.MaxMem > 100:
		return rule, errors.New("max mem threshold should be in the range 0<=t<=100")
	case rule.MinMem < 0 || rule.MinMem > 100:
		return rule, errors.New("min mem threshold should be in the range 0<=t<=100")
	case rule.MinMem >= rule.MaxMem && !(rule.MinMem == 0 && rule.MaxMem == 0):
		return rule, errors.New("min mem threshold should be less than max mem threshold")
	case rule.MinMem == 0 && rule.MaxMem == 0 && rule.MinCpu == 0 && rule.MaxCpu == 0:
		return rule, errors.New("no cpu/mem thresholds defined")
	}

	switch {
	case rule.MinMem == 0 && rule.MaxMem == 0:
		// disable the memory thresholds
		rule.MinMem, rule.MaxMem = math.MaxInt32, math.MaxInt32
	case rule.MinCpu == 0 && rule.MaxCpu == 0:
		// disable the cpu thresholds
		rule.MinCpu, rule.MaxCpu = math.MaxInt32, math.MaxInt32
	}

	return rule, nil
}

func ruleFor(rules []Rule, app, space, org string) (Rule, bool) {
	for _, rule := range rules {
		if rule.App == app && rule.Space == space && rule.Org == org {
			return rule, true
		}
	}
	return Rule{}, false
}
