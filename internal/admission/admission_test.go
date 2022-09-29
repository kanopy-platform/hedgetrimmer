package admission

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/kanopy-platform/hedgetrimmer/internal/limitrange"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	corev1Listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var cfg *rest.Config

type MockLimitRanger struct {
	lrc *limitrange.Config
	err error
}

func (mlr *MockLimitRanger) SetConfig(lrc *limitrange.Config) {
	mlr.lrc = lrc
}

func (mlr *MockLimitRanger) SetErr(err error) {
	mlr.err = err
}

func (mlr *MockLimitRanger) List(selector labels.Selector) (ret []*corev1.LimitRange, err error) {
	return
}
func (mlr *MockLimitRanger) LimitRanges(namespace string) corev1Listers.LimitRangeNamespaceLister {
	return nil
}

func (mlr *MockLimitRanger) LimitRangeConfig(namespace string) (*limitrange.Config, error) {
	return mlr.lrc, mlr.err
}

func TestMain(m *testing.M) {
	flag.Parse()
	testenv := &envtest.Environment{}
	if !testing.Short() {
		testenv := &envtest.Environment{}
		var err error
		cfg, err = testenv.Start()
		if err != nil {
			panic(err)
		}
	}

	res := m.Run()

	if !testing.Short() {
		if err := testenv.Stop(); err != nil {
			panic(err)
		}
	}

	os.Exit(res)
}

func TestNewAdmissionRouter(t *testing.T) {
	t.Parallel()
	mlr := &MockLimitRanger{}
	r, err := NewRouter(mlr)
	assert.NoError(t, err)
	assert.NotNil(t, r)

}

func TestIntegrationSetupWithManager(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip()
	}
	mlr := &MockLimitRanger{}
	r, err := NewRouter(mlr)
	assert.NoError(t, err)
	m, err := manager.New(cfg, manager.Options{
		Scheme: &runtime.Scheme{},
		Port:   8084,
		Host:   "127.0.0.1",
	})

	assert.NoError(t, err)
	r.SetupWithManager(m)
}

func TestWithAdmissonHandlers_AddHandler(t *testing.T) {
	t.Parallel()
	mlr := &MockLimitRanger{}
	r, err := NewRouter(mlr, WithAdmissionHandlers(&MockDeploymentHandler{}))
	assert.NoError(t, err)
	assert.Len(t, r.handlers, 1)
}

func TestWithAdmissionHandlers_AddDuplciateHandlerFails(t *testing.T) {
	t.Parallel()
	mlr := &MockLimitRanger{}
	r, err := NewRouter(mlr, WithAdmissionHandlers(&MockDeploymentHandler{}, &MockDeploymentHandler{}))
	assert.Error(t, err)
	assert.Nil(t, r)
}

func TestAllowObjects(t *testing.T) {
	t.Parallel()
	scheme := runtime.NewScheme()
	decoder, err := admission.NewDecoder(scheme)
	assert.NoError(t, err)
	mlr := &MockLimitRanger{
		lrc: &limitrange.Config{},
	}
	r, err := NewRouter(mlr, WithAdmissionHandlers(&MockDeploymentHandler{}, &MockReplicaSetHandler{}))
	assert.NoError(t, err)
	assert.NoError(t, r.InjectDecoder(decoder))

	tests := []struct {
		object      runtime.Object
		wantAllowed bool
	}{
		{
			object: &appsv1.Deployment{
				TypeMeta: metav1.TypeMeta{
					Kind:       "Deployment",
					APIVersion: "apps/v1",
				},
			},
			wantAllowed: true,
		},
		{
			object: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ReplicaSet",
					APIVersion: "apps/v1",
				},
			},
			wantAllowed: true,
		},
		{
			object: &appsv1.ReplicaSet{
				TypeMeta: metav1.TypeMeta{
					Kind: "Unknown",
				},
			},
		},
	}

	for _, test := range tests {
		b, err := json.Marshal(test.object)
		assert.NoError(t, err)
		response := r.Handle(context.TODO(), admission.Request{AdmissionRequest: v1.AdmissionRequest{
			RequestKind: &metav1.GroupVersionKind{
				Kind: test.object.GetObjectKind().GroupVersionKind().Kind,
			},
			Object: runtime.RawExtension{
				Raw: b,
			},
		}})

		assert.Equal(t, test.wantAllowed, response.Allowed, fmt.Sprintf("allow test on %s", test.object.GetObjectKind().GroupVersionKind().Kind))

		patchLengthWant := 0
		if test.wantAllowed {
			patchLengthWant = 1
		}

		assert.Len(t, response.Patches, patchLengthWant, fmt.Sprintf("patches assert on %s", test.object.GetObjectKind().GroupVersionKind().Kind))

	}
}

type MockHandler struct {
	decoder *admission.Decoder
}

func (m *MockHandler) InjectDecoder(dec *admission.Decoder) error {
	m.decoder = dec
	return nil
}

func (m *MockHandler) PatchResponse(raw []byte, v interface{}) admission.Response {
	pjson, err := json.Marshal(v)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return admission.PatchResponseFromRaw(raw, pjson)
}

type MockDeploymentHandler struct {
	MockHandler
}

func (d *MockDeploymentHandler) Kind() string {
	return "Deployment"
}

func (d *MockDeploymentHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	deployment := &appsv1.Deployment{}
	if err := d.decoder.Decode(req, deployment); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return d.PatchResponse(req.Object.Raw, d.mutate(deployment))
}

func (d *MockDeploymentHandler) mutate(dp *appsv1.Deployment) *appsv1.Deployment {
	patched := dp.DeepCopy()
	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}
	patched.Annotations["mutated"] = "deployment"
	return patched
}

type MockReplicaSetHandler struct {
	MockHandler
}

func (r *MockReplicaSetHandler) Kind() string {
	return "ReplicaSet"
}

func (r *MockReplicaSetHandler) Handle(ctx context.Context, req admission.Request) admission.Response {

	rs := &appsv1.ReplicaSet{}

	if err := r.decoder.Decode(req, rs); err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	return r.PatchResponse(req.Object.Raw, r.mutate(rs))
}

func (r *MockReplicaSetHandler) mutate(rs *appsv1.ReplicaSet) *appsv1.ReplicaSet {
	patched := rs.DeepCopy()

	if patched.Annotations == nil {
		patched.Annotations = map[string]string{}
	}
	patched.Annotations["mutated"] = "replicaset"

	return patched
}
