package tekton

import (
	"context"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// TaskRunLog represents logs for a single TaskRun.
type TaskRunLog struct {
	TaskRunName string
	Steps       []StepLog
}

// StepLog represents logs for a single step in a TaskRun.
type StepLog struct {
	StepName  string
	Container string
	Logs      string
}

// GetTaskRunLogs retrieves logs for a TaskRun.
func (c *Client) GetTaskRunLogs(ctx context.Context, namespace, name string) (*TaskRunLog, error) {
	tr, err := c.GetTaskRun(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	podName := tr.Status.PodName
	if podName == "" {
		return nil, fmt.Errorf("taskrun %s has no pod assigned yet", name)
	}

	logs := &TaskRunLog{
		TaskRunName: name,
	}

	for _, step := range tr.Status.Steps {
		logContent, err := c.getContainerLogs(ctx, namespace, podName, step.Container)
		if err != nil {
			logContent = fmt.Sprintf("failed to get logs: %v", err)
		}

		logs.Steps = append(logs.Steps, StepLog{
			StepName:  step.Name,
			Container: step.Container,
			Logs:      logContent,
		})
	}

	return logs, nil
}

// GetPipelineRunLogs retrieves logs for all TaskRuns in a PipelineRun.
func (c *Client) GetPipelineRunLogs(ctx context.Context, namespace, name string) ([]*TaskRunLog, error) {
	trs, err := c.ListTaskRunsForPipelineRun(ctx, namespace, name)
	if err != nil {
		return nil, err
	}

	var allLogs []*TaskRunLog
	for _, tr := range trs.Items {
		logs, err := c.GetTaskRunLogs(ctx, namespace, tr.Name)
		if err != nil {
			// Log error but continue
			allLogs = append(allLogs, &TaskRunLog{
				TaskRunName: tr.Name,
				Steps: []StepLog{{
					Logs: fmt.Sprintf("failed to get logs: %v", err),
				}},
			})
			continue
		}
		allLogs = append(allLogs, logs)
	}

	return allLogs, nil
}

func (c *Client) getContainerLogs(ctx context.Context, namespace, podName, containerName string) (string, error) {
	req := c.K8s.CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
	})

	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(strings.Builder)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
