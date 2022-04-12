package logs

import (
	"context"
	"fmt"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/plural"
	corev1 "k8s.io/api/core/v1"
)

type SuiteManager struct {
	Test      *testv1alpha1.TestSuite
	Pods      map[string]*LogWatcher
	Ctx       context.Context
	Publisher *LogPublisher
	Cancel    context.CancelFunc
}

type LogManager struct {
	Socket *plural.Socket
	Config *plural.Config
	Suites map[string]*SuiteManager
}

func NewManager(config *plural.Config) *LogManager {
	socket := plural.WebSocket(config)
	return &LogManager{
		Socket: &socket,
		Config: config,
		Suites: make(map[string]*SuiteManager),
	}
}

func (mgr *LogManager) SuiteManager(test *testv1alpha1.TestSuite) (smgr *SuiteManager, err error, found bool) {
	name := mgr.name(test)
	if ssmgr, ok := mgr.Suites[name]; ok {
		smgr = ssmgr
		found = true
		return
	}

	smgr = &SuiteManager{Test: test, Pods: make(map[string]*LogWatcher)}
	smgr.Ctx, smgr.Cancel = context.WithCancel(context.Background())
	smgr.Publisher = NewPublisher(mgr, test)
	smgr.Test = test
	mgr.Suites[name] = smgr
	return
}

func (mgr *LogManager) Cancel(test *testv1alpha1.TestSuite) error {
	name := mgr.name(test)
	smgr, ok := mgr.Suites[name]
	if !ok {
		return fmt.Errorf("No manager found for %s", name)
	}

	smgr.Cancel()
	delete(mgr.Suites, name)
	smgr.Publisher.Wait.Wait()
	return smgr.Publisher.Close()
}

func (mgr *LogManager) name(test *testv1alpha1.TestSuite) string {
	return fmt.Sprintf("%s:%s", test.Namespace, test.Name)
}

func (mgr *SuiteManager) AddWatcher(pod *corev1.Pod, step *testv1alpha1.StepStatus) {
	if _, ok := mgr.Pods[pod.Name]; ok {
		return
	}

	watcher := &LogWatcher{Pod: pod, Step: step, Publisher: mgr.Publisher}
	mgr.Pods[pod.Name] = watcher
	go watcher.Tail(mgr.Ctx)
}
