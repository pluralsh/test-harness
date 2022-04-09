/*
Copyright 2022.

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

package v1alpha1

import (
	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TestStep struct {
	// the name for this step
	Name string `json:"name"`

	// a description for what this step is doing (for visualization)
	Description string `json:"description"`

	// the argo template to use for this step
	Template *argov1alpha1.Template `json:"template"`
}

// TestSuiteSpec defines the desired state of TestSuite
type TestSuiteSpec struct {
	// the tag you'll promote to on test success
	PromoteTag string `json:"promoteTag,omitempty"`

	// the repository this test is run in
	Repository string `json:"repository,omitempty"`

	// test steps to run
	Steps []*TestStep `json:"steps,omitempty"`
}

type StepStatus struct {
	// the id for this test step
	PluralId string `json:"pluralId"`

	// name of this step
	Name string `json:"name"`

	// the status of this test step
	Status plural.Status `json:"status"`
}

// TestSuiteStatus defines the observed state of TestSuite
type TestSuiteStatus struct {
	// the id for this test suite
	PluralId string `json:"pluralId"`

	// the status of the entire test
	Status plural.Status `json:"testStatus"`

	// the status for each individual step
	Steps []*StepStatus `json:"stepStatus"`

	// the name of the associated argo workflow
	WorkflowName string `json:"workflowName"`

	// time when the suite was completed
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TestSuite is the Schema for the testsuites API
type TestSuite struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TestSuiteSpec   `json:"spec,omitempty"`
	Status TestSuiteStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TestSuiteList contains a list of TestSuite
type TestSuiteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TestSuite `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TestSuite{}, &TestSuiteList{})
}
