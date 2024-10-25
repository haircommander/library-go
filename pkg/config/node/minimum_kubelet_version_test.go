package node

import (
	"testing"

	"github.com/stretchr/testify/assert"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	corev1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
)

func TestValidateConfigNodeForMinimumKubeletVersion(t *testing.T) {
	testCases := []struct {
		name         string
		version      string
		shouldReject bool
		nodes        []*v1.Node
		nodeListErr  error
		errType      field.ErrorType
		errMsg       string
	}{
		// no rejections
		{
			name:         "should not reject when minimum kubelet version is empty",
			version:      "",
			shouldReject: false,
		},
		{
			name:         "should reject when min kubelet version bogus",
			version:      "bogus",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.30.0",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "Failed to parse submitted version bogus No Major.Minor.Patch elements found",
		},
		{
			name:         "should reject when kubelet version is bogus",
			version:      "1.30.0",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "bogus",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "failed to parse node version bogus: No Major.Minor.Patch elements found",
		},
		{
			name:         "should reject when kubelet version is too old",
			version:      "1.30.0",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.29.0",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "kubelet version of node node is 1.29.0, which is lower than minimumKubeletVersion of 1.30.0",
		},
		{
			name:         "should reject when one kubelet version is too old",
			version:      "1.30.0",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.30.0",
						},
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node2",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.29.0",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "kubelet version of node node2 is 1.29.0, which is lower than minimumKubeletVersion of 1.30.0",
		},
		{
			name:         "should not reject when kubelet version is equal",
			version:      "1.30.0",
			shouldReject: false,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.30.0",
						},
					},
				},
			},
		},
		{
			name:         "should reject when min version incomplete",
			version:      "1.30",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.30.0",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "Failed to parse submitted version 1.30 No Major.Minor.Patch elements found",
		},
		{
			name:         "should reject when kubelet version incomplete",
			version:      "1.30.0",
			shouldReject: true,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.30",
						},
					},
				},
			},
			errType: field.ErrorTypeInvalid,
			errMsg:  "failed to parse node version 1.30: No Major.Minor.Patch elements found",
		},
		{
			name:         "should not reject when kubelet version is new enough",
			version:      "1.30.0",
			shouldReject: false,
			nodes: []*v1.Node{
				&v1.Node{
					ObjectMeta: metav1.ObjectMeta{
						Name: "node",
					},
					Status: v1.NodeStatus{
						NodeInfo: v1.NodeSystemInfo{
							KubeletVersion: "1.31.0",
						},
					},
				},
			},
		},
	}

	for _, testCase := range testCases {
		shouldStr := "should not be"
		if testCase.shouldReject {
			shouldStr = "should be"
		}
		t.Run(testCase.name, func(t *testing.T) {
			fieldErr := ValidateMinimumKubeletVersion(fakeNodeLister(testCase.nodes), testCase.version)
			assert.Equal(t, testCase.shouldReject, fieldErr != nil, "minimum kubelet version %q %s rejected", testCase.version, shouldStr)

			if testCase.shouldReject {
				assert.Equal(t, "spec.minimumKubeletVersion", fieldErr.Field, "field name during for mininumKubeletVersion should be spec.mininumKubeletVersion")
				assert.Equal(t, fieldErr.Type, testCase.errType, "error type should be %q", testCase.errType)
				assert.Contains(t, fieldErr.Detail, testCase.errMsg, "error message should contain %q", testCase.errMsg)
			}
		})
	}
}

func fakeNodeLister(nodes []*corev1.Node) corev1listers.NodeLister {
	indexer := cache.NewIndexer(cache.MetaNamespaceKeyFunc, cache.Indexers{})
	for _, node := range nodes {
		_ = indexer.Add(node)
	}
	return corev1listers.NewNodeLister(indexer)
}
