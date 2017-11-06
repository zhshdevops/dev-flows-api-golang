package client

import (
	"testing"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

// TestNewK8sClientSet test the apiServerClient is valid or not
func TestNewK8sClientSet(t *testing.T) {
	clientset, err := NewK8sClientSet("test", "https", "192.168.1.93:6443", "c0d7rQicMtZJkeFllBaCZSMjfaCbASDV", "v1")
	if err != nil {
		t.Errorf("should create api server client, error: %v", err)
		return
	}

	_, err = clientset.Namespaces().Get("kube-systemdd")
	if err != nil {
		t.Logf("error:%v\n", err)
		if statusErr, ok := err.(*k8serrors.StatusError); ok {
			t.Logf("error:%v, status:%v\n", statusErr, statusErr.Status())
			if statusErr.Status().Code != 404 {
				t.Errorf("error code should be 404")
			}
		} else {
			t.Errorf("unexpected error: %v\n", err)
		}
	}
}
