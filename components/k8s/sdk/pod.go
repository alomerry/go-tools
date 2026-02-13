package sdk

import (
	"context"
	"io"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// ListPods lists pods in a namespace
func (c *Client) ListPods(ctx context.Context, namespace string, opts metav1.ListOptions) (*corev1.PodList, error) {
	return c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
}

// GetPod gets a pod by name
func (c *Client) GetPod(ctx context.Context, namespace, name string, opts metav1.GetOptions) (*corev1.Pod, error) {
	return c.clientset.CoreV1().Pods(namespace).Get(ctx, name, opts)
}

// DeletePod deletes a pod by name
func (c *Client) DeletePod(ctx context.Context, namespace, name string, opts metav1.DeleteOptions) error {
	return c.clientset.CoreV1().Pods(namespace).Delete(ctx, name, opts)
}

// RestartPod restarts a pod by deleting it (relying on controller to recreate)
func (c *Client) RestartPod(ctx context.Context, namespace, name string) error {
	// Standard way to "restart" a pod is to delete it
	return c.DeletePod(ctx, namespace, name, metav1.DeleteOptions{})
}

// GetPodLogs returns the logs of a pod
func (c *Client) GetPodLogs(ctx context.Context, namespace, name string, opts *corev1.PodLogOptions) (io.ReadCloser, error) {
	req := c.clientset.CoreV1().Pods(namespace).GetLogs(name, opts)
	return req.Stream(ctx)
}

// ExecOptions holds parameters for Exec command
type ExecOptions struct {
	Namespace string
	PodName   string
	Container string
	Command   []string
	Stdin     io.Reader
	Stdout    io.Writer
	Stderr    io.Writer
	TTY       bool
}

// Exec executes a command in a pod
func (c *Client) Exec(ctx context.Context, opts ExecOptions) error {
	req := c.clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(opts.PodName).
		Namespace(opts.Namespace).
		SubResource("exec")

	option := &corev1.PodExecOptions{
		Container: opts.Container,
		Command:   opts.Command,
		Stdin:     opts.Stdin != nil,
		Stdout:    opts.Stdout != nil,
		Stderr:    opts.Stderr != nil,
		TTY:       opts.TTY,
	}

	req.VersionedParams(option, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(c.config, "POST", req.URL())
	if err != nil {
		return err
	}

	return exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  opts.Stdin,
		Stdout: opts.Stdout,
		Stderr: opts.Stderr,
		Tty:    opts.TTY,
	})
}
