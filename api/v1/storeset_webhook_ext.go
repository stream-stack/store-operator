package v1

import (
	"k8s.io/apimachinery/pkg/util/validation/field"
	"strconv"
	"strings"
)

//+kubebuilder:object:generate:=false
//+kubebuilder:object:root:=false
type ClusterValidator interface {
	ValidateCreate(c *StoreSet) field.ErrorList
	ValidateUpdate(now *StoreSet, old *StoreSet) field.ErrorList
}

//+kubebuilder:object:generate:=false
//+kubebuilder:object:root:=false
type ClusterDefaulter interface {
	Default(c *StoreSet)
}

var VersionedValidators = map[string][]ClusterValidator{}
var VersionedDefaulters = map[string][]ClusterDefaulter{}

func RegisterVersionedValidators(version string, o ClusterValidator) {
	setMaxVersion(version)
	value, ok := VersionedValidators[version]
	if !ok {
		value = make([]ClusterValidator, 0)
	}
	value = append(value, o)
	VersionedValidators[version] = value
}

const VersionSeparator = "-"

func setMaxVersion(version string) {
	if maxVersion == "" {
		maxVersion = version
		return
	}
	split := strings.Split(version, VersionSeparator)
	if len(split) != 2 {
		return
	}
	versionNowStr := split[1]
	now, err := strconv.Atoi(strings.ReplaceAll(versionNowStr, ".", ""))
	if err != nil {
		return
	}
	versionOldStr := strings.Split(maxVersion, VersionSeparator)[1]
	old, err := strconv.Atoi(strings.ReplaceAll(versionOldStr, ".", ""))
	if err != nil {
		return
	}
	if now > old {
		maxVersion = version
	}
}

func RegisterVersionedDefaulters(version string, o ClusterDefaulter) {
	value, ok := VersionedDefaulters[version]
	if !ok {
		value = make([]ClusterDefaulter, 0)
	}
	value = append(value, o)
	VersionedDefaulters[version] = value
}

var maxVersion = ""

func getVersion(version string) string {
	if version != "" {
		return version
	}
	return maxVersion
}
