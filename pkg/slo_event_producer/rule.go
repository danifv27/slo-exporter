//revive:disable:var-naming
package slo_event_producer

//revive:enable:var-naming

import (
	"github.com/prometheus/client_golang/prometheus"
	"gitlab.seznam.net/sklik-devops/slo-exporter/pkg/producer"
	"strconv"
)

type eventMetadata map[string]string

var (
	unclassifiedEventsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "slo_exporter",
		Subsystem: "slo_event_producer",
		Name:      "unclassified_events_total",
		Help:      "Total number of dropped events without classification.",
	})
)

func init() {
	prometheus.MustRegister(unclassifiedEventsTotal)
}

func (e *eventMetadata) matches(otherMetadata eventMetadata) bool {
	for k, v := range *e {
		otherV, ok := otherMetadata[k]
		if !ok {
			return false
		}
		if otherV != v {
			return false
		}
	}
	return true
}

func mergeMetadata(a, b map[string]string) map[string]string {
	newMetadata := map[string]string{}
	for k, v := range a {
		newMetadata[k] = v
	}
	for k, v := range b {
		newMetadata[k] = v
	}
	return newMetadata
}

func newEvaluationRule(opts ruleOptions) (*evaluationRule, error) {
	var failureCriteria []criterium
	for _, criteriumOpts := range opts.FailureCriteriaOptions {
		criterium, err := newCriterium(criteriumOpts)
		if err != nil {
			return nil, err
		}
		failureCriteria = append(failureCriteria, criterium)
	}
	return &evaluationRule{
		matcher:            opts.Matcher,
		failureCriteria:    failureCriteria,
		additionalMetadata: opts.AdditionalMetadata,
	}, nil
}

type evaluationRule struct {
	matcher            eventMetadata
	failureCriteria    []criterium
	additionalMetadata eventMetadata
}

func (er *evaluationRule) evaluateEvent(event *producer.RequestEvent) (*SloEvent, bool) {
	eventMetadata := event.GetSloMetadata()
	if !event.IsClassified() || eventMetadata == nil {
		unclassifiedEventsTotal.Inc()
		log.Warnf("dropping event %v with no classification", event)
		return nil, false
	}
	// Check if rule matches the event
	if er.matcher != nil {
		if !er.matcher.matches(*eventMetadata) {
			return nil, false
		}
	}
	// Evaluate all criteria and if matches any, mark it as failed.
	failed := false
	for _, criterium := range er.failureCriteria {
		log.Tracef("evaluating criterium %v", criterium)
		if criterium.Evaluate(event) {
			failed = true
			break
		}
	}
	finalMetadata := map[string]string{}
	if er.additionalMetadata != nil {
		finalMetadata = mergeMetadata(er.additionalMetadata, *eventMetadata)
	} else {
		finalMetadata = *eventMetadata
	}
	// Add label to metadata to indicate result of the event.
	finalMetadata["failed"] = strconv.FormatBool(failed)
	log.Tracef("event extended metadata: %v", finalMetadata)
	return &SloEvent{TimeOccurred: event.GetTimeOccurred(), SloMetadata: finalMetadata}, true

}