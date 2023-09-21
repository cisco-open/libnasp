// Copyright (c) 2022 Cisco and/or its affiliates. All rights reserved.
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//       https://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package proxywasm

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"emperror.dev/errors"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/cisco-open/libnasp/pkg/proxywasm/api"
)

var ErrInvalidMetricType = errors.New("invalid metric type")

type MetricType int32

const (
	MetricTypeCounter   MetricType = 0
	MetricTypeGauge     MetricType = 1
	MetricTypeHistogram MetricType = 2
	MetricTypeMax       MetricType = 2
)

type prometheusMetricHandler struct {
	metricsByID   map[int32]*metricWithLabels
	metricsByName map[string]*metricWithLabels
	latestID      int32

	mux    sync.RWMutex
	logger logr.Logger

	registry *prometheus.Registry
}

type metricWithLabels struct {
	ID     int32
	Name   string
	Metric interface{}
	Labels map[string]string
}

func NewPrometheusMetricHandler(registry *prometheus.Registry, logger logr.Logger) api.MetricHandler {
	return &prometheusMetricHandler{
		metricsByID:   make(map[int32]*metricWithLabels),
		metricsByName: make(map[string]*metricWithLabels),
		logger:        logger,

		registry: registry,
	}
}

func (h *prometheusMetricHandler) HTTPHandler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		h.registry, promhttp.HandlerFor(h.registry, promhttp.HandlerOpts{}),
	)
}

func (h *prometheusMetricHandler) DefineMetric(metricType int32, name string) int32 {
	parsedName := ParseEnvoyStatTag(name)

	parsedName.Name = strings.ReplaceAll(parsedName.Name, ".", "_")

	tags := make([]string, 0)
	for k := range parsedName.Values {
		tags = append(tags, k)
	}

	var retval int32

	defer func() {
		h.logger.V(2).Info("define metric", "type", metricType, "name", parsedName.Name, "id", retval, "values", parsedName.Values)
	}()

	switch MetricType(metricType) {
	case MetricTypeCounter:
		retval = h.Add(parsedName.Name, prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: parsedName.Name,
		}, tags), parsedName.Values)
	case MetricTypeGauge:
		retval = h.Add(parsedName.Name, prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: parsedName.Name,
		}, tags), parsedName.Values)
	case MetricTypeHistogram:
		retval = h.Add(parsedName.Name, prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: parsedName.Name,
		}, tags), parsedName.Values)
	}

	return retval
}

func (h *prometheusMetricHandler) RecordMetric(metricID int32, value int64) error {
	h.logger.V(2).Info("record metric", "id", metricID, "value", value)

	if m, ok := h.Get(metricID); ok {
		switch metric := m.Metric.(type) {
		case *prometheus.CounterVec:
			metric.With(prometheus.Labels(m.Labels)).Add(float64(value))

			return nil
		case *prometheus.GaugeVec:
			metric.With(prometheus.Labels(m.Labels)).Add(float64(value))

			return nil
		case *prometheus.HistogramVec:
			metric.With(prometheus.Labels(m.Labels)).Observe(float64(value))

			return nil
		}
	}

	return ErrInvalidMetricType
}

func (h *prometheusMetricHandler) IncrementMetric(id int32, offset int64) error {
	h.logger.V(2).Info("increment metric", "id", id, "offset", offset)

	if m, ok := h.Get(id); ok {
		switch metric := m.Metric.(type) {
		case *prometheus.CounterVec:
			metric.With(prometheus.Labels(m.Labels)).Add(float64(offset))

			return nil
		case *prometheus.GaugeVec:
			metric.With(prometheus.Labels(m.Labels)).Add(float64(offset))

			return nil
		case *prometheus.HistogramVec:
			metric.With(prometheus.Labels(m.Labels)).Observe(float64(offset))

			return nil
		}
	}

	return ErrInvalidMetricType
}

func (h *prometheusMetricHandler) Add(name string, metric interface{}, labels map[string]string) int32 {
	h.mux.Lock()
	defer h.mux.Unlock()

	addMetric := func(name string, metric interface{}, labels map[string]string) *metricWithLabels {
		h.latestID++

		m := &metricWithLabels{
			ID:     h.latestID,
			Name:   name,
			Metric: metric,
			Labels: labels,
		}

		h.metricsByName[name] = m
		h.metricsByID[h.latestID] = m

		return m
	}

	if m, ok := metric.(prometheus.Collector); ok {
		if existingMetric, ok := h.metricsByName[name]; ok && reflect.DeepEqual(existingMetric.Labels, labels) {
			h.logger.V(2).Info("metric already exists", "id", existingMetric.ID, "name", name, "labels", labels)

			return existingMetric.ID
		} else if ok {
			addedMetric := addMetric(existingMetric.Name, existingMetric.Metric, labels)
			h.logger.V(2).Info("add metric", "id", addedMetric.ID, "name", name, "labels", labels)

			return addedMetric.ID
		} else if err := h.registry.Register(m); err != nil {
			h.logger.Error(err, "could not register metric")

			return -1
		}

		addedMetric := addMetric(name, metric, labels)
		h.logger.V(2).Info("add metric", "id", addedMetric.ID, "name", name, "labels", labels)

		return addedMetric.ID
	}

	return -1
}

func (h *prometheusMetricHandler) Remove(id int32) {
	h.mux.Lock()
	defer h.mux.Unlock()

	if metric, ok := h.metricsByID[id]; ok {
		h.logger.V(2).Info("remove metric", "id", id)
		delete(h.metricsByID, id)
		delete(h.metricsByName, metric.Name)
	}
}

func (h *prometheusMetricHandler) Get(id int32) (*metricWithLabels, bool) {
	h.mux.Lock()
	defer h.mux.Unlock()

	h.logger.V(2).Info("get metric", "id", id)

	if m, ok := h.metricsByID[id]; ok {
		return m, ok
	}

	return nil, false
}
