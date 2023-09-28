package metrics

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tektoncd/experimental/metrics-operator/pkg/metrics/recorder"
	pipelinev1beta1 "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func (m *MetricManager) RecordTaskRunDone(ctx context.Context, taskRun *pipelinev1beta1.TaskRun) error {
	if !taskRun.IsDone() {
		return fmt.Errorf("record task run done called with a running TaskRun")
	}
	m.rw.Lock()
	defer m.rw.Unlock()
	key := fmt.Sprintf("%s/%s/%s", taskRun.GetNamespace(), taskRun.GetName(), taskRun.GetUID())
	_, exists := m.runs[key]
	if !exists {
		m.runs[key] = &sync.Once{}
	}

	run := recorder.TaskRunDimensions(taskRun)

	m.runs[key].Do(func() {
		m.GetIndex().Record(ctx, run, "histogram")
		m.GetIndex().Record(ctx, run, "counter")
		m.GetIndex().Record(ctx, run, "gauge")
	})
	time.AfterFunc(1*time.Hour, func() {
		m.clean(ctx, run)
	})
	return nil
}

func (m *MetricManager) RecordTaskRunRunning(ctx context.Context, taskRun *pipelinev1beta1.TaskRun) error {
	if taskRun.IsDone() {
		return fmt.Errorf("record task run running called with a done TaskRun")
	}
	run := recorder.TaskRunDimensions(taskRun)
	m.GetIndex().Record(ctx, run, "gauge")
	return nil
}
