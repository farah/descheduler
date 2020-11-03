/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package strategies

import (
	"fmt"
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/descheduler/test"
)

func BuildTestPod2(name string, cpu int64, memory int64, nodeName string, apply func(*v1.Pod)) *v1.Pod {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      name,
			SelfLink:  fmt.Sprintf("/api/v1/namespaces/default/pods/%s", name),
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{},
						Limits:   v1.ResourceList{},
					},
				},
				{
					Resources: v1.ResourceRequirements{
						Requests: v1.ResourceList{},
						Limits:   v1.ResourceList{},
					},
				},
			},
			NodeName: nodeName,
		},
	}
	if cpu >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceCPU] = *resource.NewMilliQuantity(cpu, resource.DecimalSI)
	}
	if memory >= 0 {
		pod.Spec.Containers[0].Resources.Requests[v1.ResourceMemory] = *resource.NewQuantity(memory, resource.DecimalSI)
	}
	if apply != nil {
		apply(pod)
	}
	return pod
}
func TestGetUniqueOwnerKeyOfPod(t *testing.T) {
	node := test.BuildTestNode("n1", 2000, 3000, 2, nil)
	p1 := BuildTestPod2("p1", 100, 0, node.Name, nil)
	p1.Namespace = "different-images"

	p2 := BuildTestPod2("p2", 100, 0, node.Name, nil)
	p2.Namespace = "different-images"

	ownerRef1 := test.GetReplicaSetOwnerRefList()

	p1.Spec.Containers[0].Image = "C1"
	p1.Spec.Containers[1].Image = "C2"
	p1.ObjectMeta.OwnerReferences = ownerRef1

	p2.Spec.Containers[0].Image = "C3"
	p2.Spec.Containers[1].Image = "C4"
	p2.ObjectMeta.OwnerReferences = ownerRef1

	testCases := []struct {
		description string
		pod         *v1.Pod
		expected    string
	}{
		{
			pod:      p1,
			expected: "different-images/ReplicaSet/replicaset-1/C1",
		},
		{
			pod:      p2,
			expected: "different-images/ReplicaSet/replicaset-1/C3",
		},
	}

	for _, testCase := range testCases {
		uniqueKey := GetUniqueOwnerKeyOfPod(testCase.pod)
		if uniqueKey != testCase.expected {
			t.Errorf("Expected unique key %v, got %v", testCase.expected, uniqueKey)
		}
	}
}
