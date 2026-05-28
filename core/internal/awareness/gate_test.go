package awareness

import (
	"testing"

	"flint/core/internal/config"
)

func TestAllowed_Spark(t *testing.T) {
	allowed := []Feature{FeatureManualTools, FeatureCLI}
	blocked := []Feature{
		FeatureDaemon, FeatureObservations, FeaturePreCommit,
		FeatureErrorLog, FeatureAutoFix, FeatureTripwires,
		FeatureSessionTracking, FeatureForgeWatchAll, FeatureHumanLayer,
	}

	for _, f := range allowed {
		if !Allowed(config.Spark, f) {
			t.Errorf("spark: feature %d should be allowed", f)
		}
	}
	for _, f := range blocked {
		if Allowed(config.Spark, f) {
			t.Errorf("spark: feature %d should be blocked", f)
		}
	}
}

func TestAllowed_Flame(t *testing.T) {
	allowed := []Feature{
		FeatureManualTools, FeatureCLI, FeatureDaemon,
		FeatureObservations, FeaturePreCommit, FeatureErrorLog,
		FeatureAutoFix, FeatureTripwires, FeatureSessionTracking,
	}
	blocked := []Feature{FeatureForgeWatchAll, FeatureHumanLayer}

	for _, f := range allowed {
		if !Allowed(config.Flame, f) {
			t.Errorf("flame: feature %d should be allowed", f)
		}
	}
	for _, f := range blocked {
		if Allowed(config.Flame, f) {
			t.Errorf("flame: feature %d should be blocked", f)
		}
	}
}

func TestAllowed_Forge(t *testing.T) {
	// Forge gets everything flame has plus forge-only features
	all := []Feature{
		FeatureManualTools, FeatureCLI, FeatureDaemon,
		FeatureObservations, FeaturePreCommit, FeatureErrorLog,
		FeatureAutoFix, FeatureTripwires, FeatureSessionTracking,
		FeatureForgeWatchAll, FeatureHumanLayer,
	}
	for _, f := range all {
		if !Allowed(config.Forge, f) {
			t.Errorf("forge: feature %d should be allowed", f)
		}
	}
}

func TestAllowed_UnknownLevel(t *testing.T) {
	// An unrecognised awareness level should block everything
	unknown := config.AwarenessLevel("unknown")
	for _, f := range []Feature{FeatureManualTools, FeatureCLI, FeatureDaemon} {
		if Allowed(unknown, f) {
			t.Errorf("unknown level: feature %d should be blocked", f)
		}
	}
}

func TestRequireDaemon(t *testing.T) {
	if RequireDaemon(config.Spark) {
		t.Error("spark should not require daemon")
	}
	if !RequireDaemon(config.Flame) {
		t.Error("flame should require daemon")
	}
	if !RequireDaemon(config.Forge) {
		t.Error("forge should require daemon")
	}
}
