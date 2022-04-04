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

package controllers

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/pluralsh/test-harness/pkg/plural"
	"github.com/pluralsh/test-harness/pkg/utils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	argov1alpha1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
)

// TestSuiteReconciler reconciles a TestSuite object
type TestSuiteReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Plural *plural.Client
}

const (
	ownedAnnotation = "test.plural.sh/owned-by"
	entrypointName  = "__entrypoint__"
)

//+kubebuilder:rbac:groups=test.plural.sh,resources=testsuites,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=argoproj.io,resources=workflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=test.plural.sh,resources=testsuites/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=test.plural.sh,resources=testsuites/finalizers,verbs=update

func (r *TestSuiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("testsuite", req.NamespacedName)

	var suite testv1alpha1.TestSuite
	if err := r.Get(ctx, req.NamespacedName, &suite); err != nil {
		log.Error(err, "Failed to fetch testsuite resource")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if suite.Status.WorkflowName == "" {
		// suite hasn't been set up yet so set it up
		wf := suiteToWorkflow(&suite)
		plrl := suiteToPluralTest(&suite)
		tst, err := r.Plural.CreateTest(suite.Spec.Repository, plrl)
		if err != nil {
			log.Error(err, "failed to create plural test")
			return ctrl.Result{}, err
		}

		suite.Status.PluralId = tst.Id
		statuses := stepStatuses(&suite)
		for _, step := range tst.Steps {
			if status, ok := statuses[step.Name]; ok {
				status.PluralId = step.Id
			}
		}

		if err := controllerutil.SetControllerReference(&suite, wf, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.Create(ctx, wf); err != nil {
			log.Error(err, "failed to create workflow")
			return ctrl.Result{}, err
		}

		if err := r.Status().Update(ctx, &suite); err != nil {
			log.Error(err, "failed to update suite status")
			return ctrl.Result{}, err
		}

		return ctrl.Result{}, nil
	}

	var wf argov1alpha1.Workflow
	if err := r.Get(ctx, types.NamespacedName{Namespace: suite.Namespace, Name: suite.Status.WorkflowName}, &wf); err != nil {
		log.Error(err, "could not find associated workflow")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	syncWorkflowStatus(&wf, &suite)
	plrl := suiteToPluralTest(&suite)
	if _, err := r.Plural.UpdateTest(plrl); err != nil {
		log.Error(err, "failed to update plural test")
		return ctrl.Result{}, nil
	}

	if err := r.Status().Update(ctx, &suite); err != nil {
		log.Error(err, "failed to update suite status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func syncWorkflowStatus(wf *argov1alpha1.Workflow, suite *testv1alpha1.TestSuite) {
	suite.Status.Status = toPluralStatus(string(wf.Status.Phase))
	statuses := stepStatuses(suite)
	for _, nodeStatus := range wf.Status.Nodes {
		if status, ok := statuses[nodeStatus.TemplateName]; ok {
			status.Status = toPluralStatus(string(nodeStatus.Phase))
		}
	}
}

func suiteToWorkflow(suite *testv1alpha1.TestSuite) (workflow *argov1alpha1.Workflow) {
	name := fmt.Sprintf("%s-%s", suite.Name, utils.RandomStr(8))
	workflow.Name = name
	workflow.Namespace = suite.Namespace
	workflow.Annotations = map[string]string{}
	workflow.Annotations[ownedAnnotation] = suite.Name

	workflow.Spec.Entrypoint = entrypointName
	workflow.Spec.Templates = make([]argov1alpha1.Template, 0)
	for _, step := range suite.Spec.Steps {
		step.Template.Name = step.Name
		workflow.Spec.Templates = append(workflow.Spec.Templates, *step.Template)
	}

	curr := suite.Spec.Steps[0].Name
	dag := &argov1alpha1.DAGTemplate{
		Tasks: []argov1alpha1.DAGTask{{Name: curr, Template: curr}},
	}
	for _, step := range suite.Spec.Steps[1:] {
		dag.Tasks = append(dag.Tasks, argov1alpha1.DAGTask{
			Name:         step.Name,
			Template:     step.Name,
			Dependencies: []string{curr},
		})
		curr = step.Name
	}
	workflow.Spec.Templates = append(workflow.Spec.Templates, argov1alpha1.Template{DAG: dag, Name: entrypointName})

	// wire in workflow details to the base suite resource
	suite.Status.WorkflowName = name
	suite.Status.Status = plural.StatusQueued
	suite.Status.Steps = make([]*testv1alpha1.StepStatus, 0)
	for _, step := range suite.Spec.Steps {
		suite.Status.Steps = append(suite.Status.Steps, &testv1alpha1.StepStatus{
			Name:   step.Name,
			Status: plural.StatusQueued,
		})
	}

	return
}

// SetupWithManager sets up the controller with the Manager.
func (r *TestSuiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&testv1alpha1.TestSuite{}).
		Owns(&argov1alpha1.Workflow{}).
		Complete(r)
}
