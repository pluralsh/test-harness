package controllers

import (
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
)

func suiteToPluralTest(suite *testv1alpha1.TestSuite) (test *plural.Test) {
	test.Id = suite.Status.PluralId
	test.PromoteTag = suite.Spec.PromoteTag
	test.Status = suite.Status.Status
	test.Steps = make([]*plural.TestStep, 0)

	statuses := stepStatuses(suite)
	for _, step := range suite.Spec.Steps {
		if status, ok := statuses[step.Name]; ok {
			test.Steps = append(test.Steps, &plural.TestStep{
				Id:          status.PluralId,
				Name:        step.Name,
				Description: step.Description,
				Status:      status.Status,
			})
		}
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
