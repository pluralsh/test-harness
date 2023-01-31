package controllers

import (
	"time"

	"github.com/pluralsh/gqlclient"
	"github.com/pluralsh/gqlclient/pkg/utils"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
)

func suiteCompleted(suite *testv1alpha1.TestSuite) bool {
	if suite.Status.CompletionTime != nil {
		return true
	}

	return suite.Status.Status == plural.StatusSucceeded || suite.Status.Status == plural.StatusFailed
}

func suiteExpired(suite *testv1alpha1.TestSuite) bool {
	if suite.Status.CompletionTime == nil {
		return true
	}

	return suite.Status.CompletionTime.Time.Add(suiteExpiry).Before(time.Now())
}

func suiteToPluralTest(suite *testv1alpha1.TestSuite) (test gqlclient.TestAttributes) {
	status := gqlclient.TestStatus(suite.Status.Status)
	test.Name = &suite.Name
	test.Status = &status
	test.PromoteTag = &suite.Spec.PromoteTag
	test.Tags = utils.ToStringArrayPtr(suite.Spec.Tags)
	test.Steps = make([]*gqlclient.TestStepAttributes, 0)

	statuses := stepStatuses(suite)
	for _, step := range suite.Spec.Steps {
		status, ok := statuses[step.Name]
		stepStatus := plural.StatusQueued
		if ok {
			stepStatus = status.Status
		}

		tsa := &gqlclient.TestStepAttributes{
			Name:        &step.Name,
			Description: &step.Description,
		}
		if status.PluralId != "" {
			tsa.ID = &status.PluralId
		}
		tsaStatus := gqlclient.TestStatus(stepStatus)
		tsa.Status = &tsaStatus
		test.Steps = append(test.Steps, tsa)
	}

	return
}

func stepStatuses(suite *testv1alpha1.TestSuite) map[string]*testv1alpha1.StepStatus {
	res := map[string]*testv1alpha1.StepStatus{}
	for _, step := range suite.Status.Steps {
		res[step.Name] = step
	}
	return res
}

func toPluralStatus(argo string) plural.Status {
	switch argo {
	case "Pending":
		return plural.StatusQueued
	case "Running":
		return plural.StatusRunning
	case "Succeeded":
		return plural.StatusSucceeded
	case "Failed":
		return plural.StatusFailed
	case "Error":
		return plural.StatusFailed
	}

	return plural.StatusQueued
}
