//revive:disable:var-naming
package slo_event_producer

//revive:enable:var-naming

import (
	multierror "github.com/hashicorp/go-multierror"
	"gitlab.seznam.net/sklik-devops/slo-exporter/pkg/producer"
)

type EventEvaluator interface {
	Evaluate(event *producer.RequestEvent, outChan chan<- *SloEvent)
	AddEvaluationRule(*evaluationRule)
}

func NewEventEvaluatorFromConfigFile(path string) (EventEvaluator, error) {
	var config rulesConfig
	if _, err := config.loadFromFile(path); err != nil {
		return nil, err
	}
	evaluator, err := NewEventEvaluatorFromConfig(&config)
	if err != nil {
		return nil, err
	}
	return evaluator, nil
}

func NewEventEvaluatorFromConfig(config *rulesConfig) (EventEvaluator, error) {
	var configurationErrors error
	evaluator := requestEventEvaluator{}
	for _, ruleOpts := range config.Rules {
		rule, err := newEvaluationRule(ruleOpts)
		if err != nil {
			log.Errorf("invalid rule configuration: %v", err)
			configurationErrors = multierror.Append(configurationErrors, err)
			continue
		}
		evaluator.AddEvaluationRule(rule)
	}
	return &evaluator, configurationErrors
}
