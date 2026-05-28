package awareness

import "flint/core/internal/config"

// Feature is a capability that may be gated by awareness level.
type Feature int

const (
	// Available at all levels
	FeatureManualTools Feature = iota
	FeatureCLI

	// Flame and above
	FeatureDaemon
	FeatureObservations
	FeaturePreCommit
	FeatureErrorLog
	FeatureAutoFix
	FeatureTripwires
	FeatureSessionTracking

	// Forge only
	FeatureForgeWatchAll
	FeatureHumanLayer // also requires config.HumanLayer = true
)

var flameFeatures = map[Feature]bool{
	FeatureManualTools:     true,
	FeatureCLI:             true,
	FeatureDaemon:          true,
	FeatureObservations:    true,
	FeaturePreCommit:       true,
	FeatureErrorLog:        true,
	FeatureAutoFix:         true,
	FeatureTripwires:       true,
	FeatureSessionTracking: true,
}

var forgeFeatures = map[Feature]bool{
	FeatureForgeWatchAll: true,
	FeatureHumanLayer:    true,
}

// Allowed reports whether feature f is available at the given awareness level.
// For FeatureHumanLayer, the caller must also check config.HumanLayer.
func Allowed(level config.AwarenessLevel, f Feature) bool {
	switch level {
	case config.Spark:
		return f == FeatureManualTools || f == FeatureCLI
	case config.Flame:
		return flameFeatures[f]
	case config.Forge:
		return flameFeatures[f] || forgeFeatures[f]
	}
	return false
}

// RequireDaemon returns true when the awareness level supports daemon features.
func RequireDaemon(level config.AwarenessLevel) bool {
	return level == config.Flame || level == config.Forge
}
