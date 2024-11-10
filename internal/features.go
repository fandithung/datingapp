package internal

import (
	"context"
)

type activeFeaturesKey struct{}

const (
	FeatureDailyResponses = "daily_responses"
)

func HasFeature(ctx context.Context, featureName string) bool {
	features, ok := ctx.Value("active_features").([]*UserFeature)
	if !ok {
		return false
	}

	for _, f := range features {
		if f.FeatureName == featureName {
			return true
		}
	}
	return false
}

func SetActiveFeatures(ctx context.Context, features []*UserFeature) context.Context {
	return context.WithValue(ctx, activeFeaturesKey{}, features)
}

func GetActiveFeatures(ctx context.Context) ([]*UserFeature, bool) {
	features, ok := ctx.Value(activeFeaturesKey{}).([]*UserFeature)
	return features, ok
}
