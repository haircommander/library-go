package node

import (
	"fmt"
	"strings"

	"github.com/blang/semver/v4"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation/field"
	corev1listers "k8s.io/client-go/listers/core/v1"
)

func ValidateMinimumKubeletVersion(nodeLister corev1listers.NodeLister, minimumKubeletVersion string) *field.Error {
	// unset, no error
	if minimumKubeletVersion == "" {
		return nil
	}

	fieldPath := field.NewPath("spec", "minimumKubeletVersion")
	nodes, err := nodeLister.List(labels.Everything())
	if err != nil {
		return field.Forbidden(fieldPath, fmt.Sprintf("Getting nodes to compare minimum version %v", err.Error()))
	}

	version, err := semver.Parse(minimumKubeletVersion)
	if err != nil {
		return field.Invalid(fieldPath, minimumKubeletVersion, fmt.Sprintf("Failed to parse submitted version %s %v", minimumKubeletVersion, err.Error()))
	}

	for _, node := range nodes {
		_, errStr := IsKubeletVersionTooOld(node, &version)
		if errStr != "" {
			return field.Invalid(fieldPath, minimumKubeletVersion, errStr)
		}
	}
	return nil
}

func IsKubeletVersionTooOld(node *corev1.Node, minVersion *semver.Version) (bool, string) {
	version, err := semver.Parse(strings.TrimPrefix(node.Status.NodeInfo.KubeletVersion, "v"))
	if err != nil {
		return false, fmt.Sprintf("failed to parse node version %s: %v", node.Status.NodeInfo.KubeletVersion, err)
	}

	version.Pre = nil
	version.Build = nil

	name := node.ObjectMeta.Name
	if minVersion.GT(version) {
		return true, fmt.Sprintf("kubelet version of node %s is %v, which is lower than minimumKubeletVersion of %v", name, version, *minVersion)
	}
	return false, ""
}
