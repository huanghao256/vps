package checks

import "github.com/vps-inspector/vps-inspector/internal/agent"

// DefaultRegistry returns the checks included in the standard VPS inspection run.
func DefaultRegistry() []agent.Check {
	return []agent.Check{
		NewSystemCheck(),
		NewNetworkOverviewCheck(),
		NewRouteProfileCheck(DefaultRouteTargets()),
		NewLatencyCheck(DefaultLatencyTargets()),
		NewBandwidthCheck(),
		NewStabilityCheck(DefaultLatencyTargets()),
		NewRiskCheck(),
	}
}
