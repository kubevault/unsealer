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
	"os"
	"strings"
	"time"

	"kubevault.dev/unsealer/pkg/vault/guard"

	"github.com/pkg/errors"
	core "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	markerKeyActiveDC  = "activeDC"
	markerKeyRenewTime = "renewTime"

	guardInterval = time.Second
	// markerStaleAfter: a marker whose renewTime is older is stale (fail closed).
	markerStaleAfter = 30 * time.Second
	// markerDamperAfter: a marker naming another data center whose renewTime is
	// older than this (but not stale) is left alone (the damper).
	markerDamperAfter = 12 * time.Second

	guardTokenRenewPeriod = time.Hour
	guardTokenRetryPeriod = time.Minute
)

// Guard keeps the local Vault node from holding the HA lock outside the data
// center named by the dr-controlplane marker. It never seals Vault and never
// touches another node.
type Guard struct {
	Env       *Env
	TokenFile string
	KC        client.Client
	VC        *guard.Client

	nextRenew time.Time
}

type guardDecision struct {
	stepDown bool
	reason   string
}

// decide implements the guard's rules. It is called only for an active node.
// members, when non-empty, are the data centers that run this VaultServer.
func decide(ownDC string, members map[string]bool, cm *core.ConfigMap, readErr error, now time.Time) guardDecision {
	if readErr != nil {
		return guardDecision{true, "marker unreadable: " + readErr.Error()}
	}
	if cm == nil {
		return guardDecision{true, "marker missing"}
	}
	activeDC := strings.TrimSpace(cm.Data[markerKeyActiveDC])
	rt := strings.TrimSpace(cm.Data[markerKeyRenewTime])
	if activeDC == "" || rt == "" {
		return guardDecision{true, "marker is missing activeDC or renewTime"}
	}
	renewTime, err := time.Parse(time.RFC3339, rt)
	if err != nil {
		return guardDecision{true, "marker renewTime is unparsable: " + err.Error()}
	}
	age := now.Sub(renewTime)
	if age > markerStaleAfter {
		return guardDecision{true, "marker is stale, renewTime age " + age.String()}
	}
	if activeDC == ownDC {
		return guardDecision{false, "marker names this data center"}
	}
	// The Lease Vault follows can be held by a data center that runs no Vault
	// pods (the global Lease, shared with other workloads). No node could ever
	// satisfy it, and stepping down would leave Vault without an active node, so
	// any data center's node may keep the lock.
	if len(members) > 0 && !members[activeDC] {
		return guardDecision{false, "marker names data center " + activeDC + ", which runs no Vault pods"}
	}
	if age > markerDamperAfter {
		return guardDecision{false, "marker names data center " + activeDC + " but renewTime age is " + age.String() + " (damper)"}
	}
	return guardDecision{true, "marker names data center " + activeDC}
}

func (g *Guard) readToken() (string, error) {
	data, err := os.ReadFile(g.TokenFile)
	if err != nil {
		return "", errors.Wrapf(err, "failed to read guard token file %s", g.TokenFile)
	}
	token := strings.TrimSpace(string(data))
	if token == "" {
		return "", errors.Errorf("guard token file %s is empty", g.TokenFile)
	}
	return token, nil
}

func (g *Guard) readMarker(ctx context.Context) (*core.ConfigMap, error) {
	if g.Env.MarkerNamespace == "" || g.Env.MarkerName == "" {
		return nil, errors.Errorf("env %s and %s must be set in distributed mode", EnvDCDRNamespace, EnvDCDRMarker)
	}
	cm := &core.ConfigMap{}
	if err := g.KC.Get(ctx, client.ObjectKey{Namespace: g.Env.MarkerNamespace, Name: g.Env.MarkerName}, cm); err != nil {
		return nil, err
	}
	return cm, nil
}

func (g *Guard) Run() {
	klog.Infof("starting dc-dr guard for data center %q, marker %s/%s", g.Env.DCName, g.Env.MarkerNamespace, g.Env.MarkerName)

	ticker := time.NewTicker(guardInterval)
	defer ticker.Stop()
	for range ticker.C {
		g.check()
	}
}

func (g *Guard) check() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	active, unsealed, err := g.VC.Health(ctx)
	if err != nil {
		klog.V(4).Infof("dc-dr guard: %s", err.Error())
		return
	}

	// renew after the step-down decision, so a slow renewal never delays it
	if unsealed {
		defer g.renewToken()
	}

	if !active {
		return
	}

	// Only an active node needs the marker, so standbys do not poll the API
	// server every second.
	cm, markerErr := g.readMarker(ctx)
	d := decide(g.Env.DCName, g.Env.Members, cm, markerErr, time.Now())
	if !d.stepDown {
		klog.V(4).Infof("dc-dr guard: active node, nothing to do: %s", d.reason)
		return
	}

	klog.Infof("dc-dr guard: stepping down the active node: %s", d.reason)
	token, err := g.readToken()
	if err != nil {
		klog.Errorf("dc-dr guard: cannot step down: %s", err.Error())
		return
	}
	steppedDown, err := g.VC.StepDown(ctx, token)
	if err != nil {
		klog.Errorf("dc-dr guard: %s", err.Error())
		return
	}
	if !steppedDown {
		klog.Infoln("dc-dr guard: node is no longer active, nothing to do")
		return
	}
	klog.Infoln("dc-dr guard: the active node stepped down")
}

func (g *Guard) renewToken() {
	if time.Now().Before(g.nextRenew) {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := g.readToken()
	if err == nil {
		err = g.VC.RenewSelf(ctx, token)
	}
	if err != nil {
		klog.Errorf("dc-dr guard: failed to renew the guard token: %s", err.Error())
		g.nextRenew = time.Now().Add(guardTokenRetryPeriod)
		return
	}
	klog.Infoln("dc-dr guard: renewed the guard token")
	g.nextRenew = time.Now().Add(guardTokenRenewPeriod)
}
