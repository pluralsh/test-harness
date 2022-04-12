package logs

import (
	"bufio"
	"context"
	"fmt"
	testv1alpha1 "github.com/pluralsh/test-harness/api/v1alpha1"
	"github.com/pluralsh/test-harness/pkg/utils"
	"github.com/sethvargo/go-retry"
	"io"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"os"
	"sync"
	"time"
)

type LogWatcher struct {
	Pod       *corev1.Pod
	Step      *testv1alpha1.StepStatus
	Publisher *LogPublisher
}

const (
	sinceSeconds int64 = 60 * 60 * 24
)

func (w *LogWatcher) Tail(ctx context.Context) error {
	config, err := utils.KubeConfig()
	if err != nil {
		return err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	f, err := ioutil.TempFile("", w.Pod.Name)
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())

	wg := &sync.WaitGroup{}
	functionList := []func(){}
	for _, container := range w.Pod.Spec.Containers {
		podLogOpts := &corev1.PodLogOptions{
			Follow:       true,
			SinceSeconds: utils.Int64(sinceSeconds),
			Container:    container.Name,
		}

		var podLogs io.ReadCloser
		backoff := retry.NewExponential(1 * time.Second)
		backoff = retry.WithMaxRetries(10, backoff)
		backoff = retry.WithJitterPercent(5, backoff)
		if err := retry.Do(ctx, backoff, func(ctx context.Context) error {
			logs, err := clientset.CoreV1().Pods(w.Pod.Namespace).GetLogs(w.Pod.Name, podLogOpts).Stream(ctx)
			if err != nil {
				fmt.Println("Failed to tail pod logs", err)
				return retry.RetryableError(err)
			}
			podLogs = logs
			return nil
		}); err != nil {
			return err
		}
		defer podLogs.Close()

		functionList = append(functionList, func() {
			defer wg.Done()
			reader := bufio.NewScanner(podLogs)
			for reader.Scan() {
				select {
				case <-ctx.Done():
					return
				default:
					line := reader.Text()
					f.WriteString(line + "\n")
					if err := w.Publisher.Publish(line, w.Step); err != nil {
						fmt.Println("failed to publish line", err)
					}
				}
			}
		})
	}

	w.Publisher.Wait.Add(1)
	defer w.Publisher.Wait.Done()
	wg.Add(len(functionList))
	for _, f := range functionList {
		go f()
	}
	wg.Wait()
	fmt.Println("uploading logfile to plural")
	return w.uploadFile(f)
}

func (w *LogWatcher) uploadFile(f *os.File) error {
	stepId := w.Step.PluralId
	if err := w.Publisher.Client.UpdateStep(stepId, f.Name()); err != nil {
		fmt.Println("failed to upload logs", err)
		return err
	}

	return nil
}
