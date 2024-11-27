package main

import (
	"slices"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

type StateMetric struct {
	AllStates []string
	Name      string
	Gauge     *prometheus.GaugeVec
}

// Define metrics
var (
	nodepoolPowerState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nodepool_power_state",
			Help: "The power status of a nodepool",
		},
		[]string{"nodepool", "state"},
	)

	nodepoolProvisioningState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nodepool_provisioning_state",
			Help: "The provisioning status of a nodepool",
		},
		[]string{"nodepool", "state"},
	)

	nodepoolScalingState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nodepool_scaling_state",
			Help: "The health state of nodepool autoscaling",
		},
		[]string{"nodepool", "state"},
	)

	clusterScalingState = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cluster_scaling_state",
			Help: "The health state of cluster autoscaling",
		},
		[]string{"cluster", "state"},
	)
)

func NewStateMetric(states []string, name string, gauge *prometheus.GaugeVec) *StateMetric {
	return &StateMetric{
		AllStates: states,
		Name:      name,
		Gauge:     gauge,
	}
}

// Add new states to AllStates field
func (metric *StateMetric) checkForNewState(state string) {
	if !slices.Contains(metric.AllStates, state) {
		log.Infof("Adding new %s state: %s", metric.Name, state)
		metric.AllStates = append(metric.AllStates, state)
	}
}

func (metric *StateMetric) setState(label string, state string) {
	log.WithField("state", state).WithField("label", label).WithField("type", metric.Name).Debug("state update")
	metric.checkForNewState(state)
	for _, s := range metric.AllStates {
		if s != state {
			// Set all states not matching the current state to 0
			metric.Gauge.WithLabelValues(label, s).Set(0)
			continue
		}

		// Update the current state to 1
		metric.Gauge.WithLabelValues(label, state).Set(1)
	}
}

func RegisterMetrics(enableScalingStatus bool) {
	// Register prometheus metrics
	prometheus.MustRegister(nodepoolPowerState)
	prometheus.MustRegister(nodepoolProvisioningState)

	if enableScalingStatus {
		prometheus.MustRegister(nodepoolScalingState)
		prometheus.MustRegister(clusterScalingState)
	}
}
