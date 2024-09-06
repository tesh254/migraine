package utils

import "reflect"

// FeatureFlags represents the available feature flags
type FeatureFlags struct {
	VerifyDeps bool `json:"verify_deps"`
}

var ActiveFeatures FeatureFlags

// InitFeatures initializes the feature flags with default values
func InitFeatures() {
	ActiveFeatures = FeatureFlags{
		VerifyDeps: false,
	}
}

func IsFeatureEnabled(feature string) bool {
	v := reflect.ValueOf(ActiveFeatures)
	f := v.FieldByName(feature)
	if !f.IsValid() {
		return false
	}
	return f.Bool()
}

func SetFeature(feature string, enabled bool) {
	v := reflect.ValueOf(&ActiveFeatures).Elem()
	f := v.FieldByName(feature)
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
		f.SetBool(enabled)
	}
}

func RemoveFeature(feature string) {
	v := reflect.ValueOf(&ActiveFeatures).Elem()
	f := v.FieldByName(feature)
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.Bool {
		f.SetBool(false)
	}
}
