/*
Copyright AppsCode Inc. and Contributors

Licensed under the AppsCode Community License 1.0.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    https://github.com/appscode/licenses/raw/1.0.0/AppsCode-Community-1.0.0.md

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package dcdr

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	coordination "k8s.io/api/coordination/v1"
	core "k8s.io/api/core/v1"
	kerr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// initLeaseDuration is how long an unsealer that took the init Lease keeps it
// without renewing; init itself takes seconds.
const initLeaseDuration = 60 * time.Second

// MayInitialize reports whether this unsealer may run sys/init in distributed
// mode. Pods start in every data center at once on one shared storage, and Vault
// has no lock around init, so two concurrent inits would leave keys that do not
// match the barrier. Only a pod in the initializing data center may initialize
// (one data center), and within it only the pod holding the init Lease (one pod).
//
// The initializing data center is the one the marker names, or, when that data
// center runs no Vault pods, the first Member. Each data center reads its own
// marker, so two could disagree while the Lease moves; requiring a marker renewed
// within markerDamperAfter, and reading it again after taking the Lease, narrows
// that window to a Lease move during the init itself.
func (r *Registrar) MayInitialize(ctx context.Context) error {
	if err := r.initDCIsMine(ctx); err != nil {
		return err
	}
	if err := r.acquireInitLease(ctx); err != nil {
		return err
	}
	return r.initDCIsMine(ctx)
}

func (r *Registrar) initDCIsMine(ctx context.Context) error {
	cm := &core.ConfigMap{}
	if err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.MarkerNamespace, Name: r.Env.MarkerName}, cm); err != nil {
		return errors.Wrap(err, "cannot read the active-DC marker")
	}
	activeDC := strings.TrimSpace(cm.Data[markerKeyActiveDC])
	rt, err := time.Parse(time.RFC3339, strings.TrimSpace(cm.Data[markerKeyRenewTime]))
	if err != nil || time.Since(rt) > markerDamperAfter {
		return errors.New("the active-DC marker is not recently renewed")
	}
	initDC := activeDC
	if len(r.Env.Members) > 0 && !r.Env.Members[activeDC] && len(r.Env.MemberList) > 0 {
		initDC = r.Env.MemberList[0]
	}
	if initDC != r.Env.DCName {
		return errors.Errorf("data center %s initializes Vault, not %s", initDC, r.Env.DCName)
	}
	return nil
}

// ReleaseInitLease gives the init Lease up after a failed init, so a pod whose
// init keeps failing does not keep it renewed and block every other pod.
func (r *Registrar) ReleaseInitLease(ctx context.Context) {
	lease := &coordination.Lease{}
	if err := r.KC.Get(ctx, client.ObjectKey{Namespace: r.Env.PodNamespace, Name: r.Env.InitLease}, lease); err != nil {
		return
	}
	if ptr.Deref(lease.Spec.HolderIdentity, "") != r.Env.PodName {
		return
	}
	lease.Spec.HolderIdentity = nil
	lease.Spec.RenewTime = nil
	_ = r.KC.Update(ctx, lease)
}

func (r *Registrar) acquireInitLease(ctx context.Context) error {
	if r.Env.InitLease == "" {
		return errors.Errorf("env %s must be set in distributed mode", EnvDCDRInitLease)
	}
	now := metav1.NewMicroTime(time.Now())
	key := client.ObjectKey{Namespace: r.Env.PodNamespace, Name: r.Env.InitLease}

	lease := &coordination.Lease{}
	err := r.KC.Get(ctx, key, lease)
	if kerr.IsNotFound(err) {
		lease = &coordination.Lease{
			ObjectMeta: metav1.ObjectMeta{Namespace: key.Namespace, Name: key.Name},
			Spec: coordination.LeaseSpec{
				HolderIdentity:       ptr.To(r.Env.PodName),
				LeaseDurationSeconds: ptr.To(int32(initLeaseDuration / time.Second)),
				AcquireTime:          &now,
				RenewTime:            &now,
			},
		}
		return errors.Wrap(r.KC.Create(ctx, lease), "failed to take the init Lease")
	}
	if err != nil {
		return errors.Wrap(err, "failed to read the init Lease")
	}

	holder := ptr.Deref(lease.Spec.HolderIdentity, "")
	expired := lease.Spec.RenewTime == nil ||
		time.Since(lease.Spec.RenewTime.Time) > time.Duration(ptr.Deref(lease.Spec.LeaseDurationSeconds, 0))*time.Second
	if holder != r.Env.PodName && !expired {
		return errors.Errorf("pod %s holds the init Lease", holder)
	}
	if holder != r.Env.PodName {
		lease.Spec.AcquireTime = &now
	}
	lease.Spec.HolderIdentity = ptr.To(r.Env.PodName)
	lease.Spec.LeaseDurationSeconds = ptr.To(int32(initLeaseDuration / time.Second))
	lease.Spec.RenewTime = &now
	// Update carries the resourceVersion read above, so of two pods taking an
	// expired Lease at once only one succeeds.
	return errors.Wrap(r.KC.Update(ctx, lease), "failed to take the init Lease")
}
