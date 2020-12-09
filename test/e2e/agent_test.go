// +build e2emc

package e2e

import (
	"testing"

	"github.com/stretchr/testify/suite"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type AgentSuite struct {
	suite.Suite
}

func (s *AgentSuite) TestAgent() {

	config := &rest.Config{
		Host: "http://127.0.0.1:2468",
		// VnRDaElZVzBsYjJnUDFDOGZDNVE4bGFBZjdoZ1BCQzQ=
		BearerToken: "VnRDaElZVzBsYjJnUDFDOGZDNVE4bGFBZjdoZ1BCQzQ=",
	}

	clientset, err := kubernetes.NewForConfig(config)
	s.Assert().NoError(err)

	pods := clientset.CoreV1().Pods("argo")

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "test-pod-",
			Labels:       map[string]string{"test": "true"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "main", Image: "argoproj/argosay:v2"},
			},
			RestartPolicy: corev1.RestartPolicyNever,
		},
	}

	var name string

	listOptions := metav1.ListOptions{LabelSelector: "test"}

	s.Run("DeleteCollection", func() {
		err := pods.DeleteCollection(&metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "test"})
		s.Assert().NoError(err)
	})
	if s.T().Failed() {
		s.T().FailNow()
	}
	s.Run("Create", func() {
		pod, err := pods.Create(pod)
		if s.Assert().NoError(err) {
			s.Assert().NotNil(pod)
		}
	})
	s.Run("List", func() {
		podList, err := pods.List(listOptions)
		if s.Assert().NoError(err) {
			s.Assert().Len(podList.Items, 1)
			name = podList.Items[0].Name
			s.Assert().NotEmpty(name)
		}
	})
	s.Run("Get", func() {
		pod, err = pods.Get(name, metav1.GetOptions{})
		if s.Assert().NoError(err) && s.Assert().NotNil(pod) {
			s.Assert().Equal(name, pod.Name)
		}
	})
	s.Run("Patch", func() {
		pod, err := pods.Patch(pod.Name, types.MergePatchType, []byte(`{"metadata": {"annotations": {"patched": "true"}}}`))
		if s.Assert().NoError(err) && s.Assert().NotNil(pod) {
			s.Assert().NotEmpty(pod.Annotations["patched"])
		}
	})
	s.Run("Watch", func() {
		w, err := pods.Watch(listOptions)
		if s.Assert().NoError(err) && s.Assert().NotNil(w) {
			defer w.Stop()
		loop:
			for event := range w.ResultChan() {
				println(event.Type)
				switch event.Type {
				case watch.Modified:
					break loop
				default:
					if !s.Assert().NotEqual(watch.Error, event.Type) {
						break loop
					}
				}
			}
			println("done")
		}
	})
	s.Run("Update", func() {
		pod, err = pods.Get(name, metav1.GetOptions{})
		s.Assert().NoError(err)
		pod, err := pods.Update(pod)
		if s.Assert().NoError(err) && s.Assert().NotNil(pod) {
			s.Assert().Equal(name, pod.Name)
		}
	})
	s.Run("Delete", func() {
		err := pods.Delete(name, &metav1.DeleteOptions{})
		s.Assert().NoError(err)
	})
}

func TestAgentSuite(t *testing.T) {
	suite.Run(t, new(AgentSuite))
}
