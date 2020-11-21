// Copyright (c) 2020 rookie-ninja
//
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.
package rk_prom

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
)

const (
	maxKeyLength     = 128
	separator        = "::"
	namespaceDefault = "rk"
	subSystemDefault = "service"
)

var SummaryObjectives = map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001, 0.999: 0.0001}

type MetricsSet struct {
	namespace  string
	subSystem  string
	keys       map[string]bool
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	summaries  map[string]*prometheus.SummaryVec
	histograms map[string]*prometheus.HistogramVec
	lock       sync.Mutex
}

func NewMetricsSet(namespace, subSystem string) *MetricsSet {
	if len(namespace) < 1 {
		namespace = namespaceDefault
	}

	if len(subSystem) < 1 {
		subSystem = subSystemDefault
	}

	metrics := MetricsSet{
		namespace:  namespace,
		subSystem:  subSystem,
		keys:       make(map[string]bool),
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		lock:       sync.Mutex{},
	}

	return &metrics
}

// Get namespace
func (set *MetricsSet) GetNamespace() string {
	return set.namespace
}

// Get subsystem
func (set *MetricsSet) GetSubSystem() string {
	return set.subSystem
}

// Thread safe
//
// Register a counter with namespace and subsystem in MetricsSet
// If not no namespace and subsystem was provided, then default one would be applied
func (set *MetricsSet) RegisterCounter(name string, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate metrics:%s", key))
	}

	// Create a new one with default options
	opts := prometheus.CounterOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Help:      fmt.Sprintf("Counter for name:%s and labels:%s", name, labelKeys),
	}

	// It will panic if labels are not matching
	counterVec := prometheus.NewCounterVec(opts, labelKeys)

	err = prometheus.Register(counterVec)

	if err == nil {
		set.counters[key] = counterVec
		set.keys[key] = true
	}

	return err
}

// Thread safe
//
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterCounter(name string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		counterVec := set.counters[key]
		prometheus.Unregister(counterVec)

		delete(set.counters, key)
		delete(set.keys, key)
	}

	return nil
}

func (set *MetricsSet) GetCounterVec(name string) *prometheus.CounterVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return nil
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return set.counters[key]
	}

	return nil
}

func (set *MetricsSet) GetGaugeVec(name string) *prometheus.GaugeVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return nil
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return set.gauges[key]
	}

	return nil
}

func (set *MetricsSet) GetHistogramVec(name string) *prometheus.HistogramVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return nil
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return set.histograms[key]
	}

	return nil
}

func (set *MetricsSet) GetSummaryVec(name string) *prometheus.SummaryVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return nil
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return set.summaries[key]
	}

	return nil
}

func (set *MetricsSet) ListCounters() []*prometheus.CounterVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.CounterVec, 0)
	for _, v := range set.counters {
		res = append(res, v)
	}
	return res
}

func (set *MetricsSet) ListGauge() []*prometheus.GaugeVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.GaugeVec, 0)
	for _, v := range set.gauges {
		res = append(res, v)
	}
	return res
}

func (set *MetricsSet) ListHistogram() []*prometheus.HistogramVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.HistogramVec, 0)
	for _, v := range set.histograms {
		res = append(res, v)
	}
	return res
}

func (set *MetricsSet) ListSummary() []*prometheus.SummaryVec {
	set.lock.Lock()
	defer set.lock.Unlock()

	res := make([]*prometheus.SummaryVec, 0)
	for _, v := range set.summaries {
		res = append(res, v)
	}
	return res
}

// Thread safe
//
// Get counter with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetCounterWithValues(name string, values ...string) prometheus.Counter {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		counterVec := set.counters[key]
		// ignore err
		counter, _ := counterVec.GetMetricWithLabelValues(values...)
		return counter
	} else {
		return nil
	}
}

// Thread safe
//
// Get counter with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetCounterWithLabels(name string, labels prometheus.Labels) prometheus.Counter {
	set.lock.Lock()
	defer set.lock.Unlock()

	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		counterVec := set.counters[key]
		// ignore error
		counter, _ := counterVec.GetMetricWith(labels)

		return counter
	} else {
		return nil
	}
}

// Thread safe
//
// Register a gauge with namespace and subsystem in MetricsSet
// If not no namespace and subsystem was provided, then default one would be applied
func (set *MetricsSet) RegisterGauge(name string, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate metrics:%s", key))
	}

	// Create a new one with default options
	opts := prometheus.GaugeOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Help:      fmt.Sprintf("Gauge for name:%s and labels:%s", name, labelKeys),
	}

	// It will panic if labels are not matching
	gaugeVec := prometheus.NewGaugeVec(opts, labelKeys)
	err = prometheus.Register(gaugeVec)

	if err == nil {
		set.gauges[key] = gaugeVec
		set.keys[key] = true
	}

	return err
}

// Thread safe
//
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterGauge(name string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		gaugeVec := set.gauges[key]
		prometheus.Unregister(gaugeVec)

		delete(set.gauges, key)
		delete(set.keys, key)
	}

	return nil
}

// Thread safe
//
// Get gauge with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetGaugeWithValues(name string, values ...string) prometheus.Gauge {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		gaugeVec := set.gauges[key]
		// ignore error
		gauge, _ := gaugeVec.GetMetricWithLabelValues(values...)

		return gauge
	} else {
		return nil
	}
}

// Thread safe
//
// Get gauge with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetGaugeWithLabels(name string, labels prometheus.Labels) prometheus.Gauge {
	set.lock.Lock()
	defer set.lock.Unlock()

	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		gaugeVec := set.gauges[key]
		// ignore error
		gauge, _ := gaugeVec.GetMetricWith(labels)

		return gauge
	} else {
		return nil
	}
}

// Thread safe
//
// Register a summary with namespace, subsystem and objectives in MetricsSet
// If not no namespace and subsystem was provided, then default one would be applied
// If objectives is nil, then default SummaryObjectives would be applied
func (set *MetricsSet) RegisterSummary(name string, objectives map[float64]float64, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate metrics:%s", key))
	}

	if objectives == nil {
		objectives = SummaryObjectives
	}

	// Create a new one with default options
	// Create a new one with default options
	opts := prometheus.SummaryOpts{
		Namespace:  set.namespace,
		Subsystem:  set.subSystem,
		Name:       name,
		Objectives: objectives,
		Help:       fmt.Sprintf("Summary for name:%s and labels:%s", name, labelKeys),
	}

	// It will panic if labels are not matching
	summaryVec := prometheus.NewSummaryVec(opts, labelKeys)

	err = prometheus.Register(summaryVec)

	if err == nil {
		set.summaries[key] = summaryVec
		set.keys[key] = true
	}

	return err
}

// Thread safe
//
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterSummary(name string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		summaryVec := set.summaries[key]
		prometheus.Unregister(*summaryVec)

		delete(set.summaries, key)
		delete(set.keys, key)
	}

	return nil
}

// Thread safe
//
// Get summary with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetSummaryWithValues(name string, values ...string) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		summaryVec := set.summaries[key]
		// ignore error
		observer, _ := summaryVec.GetMetricWithLabelValues(values...)

		return observer
	} else {
		return nil
	}
}

// Thread safe
//
// Get summary with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetSummaryWithLabels(name string, labels prometheus.Labels) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		summaryVec := set.summaries[key]
		// ignore error
		observer, _ := summaryVec.GetMetricWith(labels)

		return observer
	} else {
		return nil
	}
}

// Thread safe
//
// Register a histogram with namespace, subsystem and objectives in MetricsSet
// If not no namespace and subsystem was provided, then default one would be applied
// If bucket is nil, then empty bucket would be applied
func (set *MetricsSet) RegisterHistogram(name string, bucket []float64, labelKeys ...string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		return errors.New(fmt.Sprintf("duplicate metrics:%s", key))
	}

	if bucket == nil {
		bucket = make([]float64, 0)
	}

	// Create a new one with default options
	// Create a new one with default options
	opts := prometheus.HistogramOpts{
		Namespace: set.namespace,
		Subsystem: set.subSystem,
		Name:      name,
		Buckets:   bucket,
		Help:      fmt.Sprintf("Histogram for name:%s and labels:%s", name, labelKeys),
	}

	// It will panic if labels are not matching
	hisVec := prometheus.NewHistogramVec(opts, labelKeys)

	err = prometheus.Register(hisVec)

	if err == nil {
		set.histograms[key] = hisVec
		set.keys[key] = true
	}

	return err
}

// Thread safe
//
// Unregister metrics, error would be thrown only when invalid name was provided
func (set *MetricsSet) UnRegisterHistogram(name string) error {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	err := validateRawName(name)
	if err != nil {
		return err
	}

	// Construct full key
	key := set.getKey(name)

	// Check existence with maps contains all keys
	if set.containsKey(key) {
		hisVec := set.histograms[key]
		prometheus.Unregister(*hisVec)

		delete(set.histograms, key)
		delete(set.keys, key)
	}

	return nil
}

// Thread safe
//
// Get histogram with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetHistogramWithValues(name string, values ...string) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	// Trim the input string of name
	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		hisVec := set.histograms[key]
		// ignore error
		observer, _ := hisVec.GetMetricWithLabelValues(values...)

		return observer
	} else {
		return nil
	}
}

// Thread safe
//
// Get histogram with values matched with labels
// Users should always be sure about the number of labels.
// If any unmatched case happens, then WARNING would be logged and you would get nil from function
func (set *MetricsSet) GetHistogramWithLabels(name string, labels prometheus.Labels) prometheus.Observer {
	set.lock.Lock()
	defer set.lock.Unlock()

	name = strings.TrimSpace(name)
	if validateRawName(name) != nil {
		return nil
	}

	key := set.getKey(name)

	if set.containsKey(key) {
		hisVec := set.histograms[key]
		// ignore error
		observer, _ := hisVec.GetMetricWith(labels)

		return observer
	} else {
		return nil
	}
}

func (set *MetricsSet) getKey(name string) string {
	key := strings.Join([]string{
		set.namespace,
		set.subSystem,
		name}, separator)

	return key
}

func (set *MetricsSet) containsKey(key string) bool {
	_, contains := set.keys[key]

	return contains
}

func validateRawName(name string) error {
	if len(name) < 1 {
		errMsg := "empty counter name"
		return errors.New(errMsg)
	}

	if len(name) > maxKeyLength {
		errMsg := fmt.Sprintf("exceed max name length:%d", maxKeyLength)
		return errors.New(errMsg)
	}

	return nil
}
