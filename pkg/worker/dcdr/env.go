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
	"fmt"
	"os"
	"strings"

	kubesliceapi "github.com/kubeslice/worker-operator/api/v1beta1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"kmodules.xyz/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Environment variables set by the operator on the unsealer container in
// cross data center (distributed) mode.
const (
	EnvDCName        = "DC_NAME"
	EnvDCDRNamespace = "DC_DR_NAMESPACE"
	EnvDCDRMarker    = "DC_DR_MARKER"
	// EnvDCDRMembers lists the Member data centers that run this VaultServer's
	// pods, comma separated.
	EnvDCDRMembers = "DC_DR_MEMBERS"
	// EnvDCDRInitLease names the Lease, in the pod's namespace, an unsealer must
	// hold to run sys/init.
	EnvDCDRInitLease = "DC_DR_INIT_LEASE"
	EnvPodName       = "POD_NAME"
	EnvPodUID        = "POD_UID"
	EnvPodNamespace  = "POD_NAMESPACE"

	// PodNameLabelKey selects a single Vault pod for its per-pod Service and ServiceExport.
	PodNameLabelKey = "vaultserver.kubevault.com/pod"

	DCDRGuardTokenFileDefault = "/var/run/secrets/kubevault.com/dc-dr-guard/token"

	vaultAPIPortName     = "api"
	vaultAPIPort         = 8200
	vaultClusterPortName = "cluster"
	vaultClusterPort     = 8201
)

// Env holds the downward API and data center inputs of distributed mode.
type Env struct {
	DCName          string
	MarkerNamespace string
	MarkerName      string
	Members         map[string]bool
	MemberList      []string
	InitLease       string
	PodName         string
	PodUID          string
	PodNamespace    string
}

func EnvFromOS() *Env {
	return &Env{
		DCName:          strings.TrimSpace(os.Getenv(EnvDCName)),
		MarkerNamespace: strings.TrimSpace(os.Getenv(EnvDCDRNamespace)),
		MarkerName:      strings.TrimSpace(os.Getenv(EnvDCDRMarker)),
		Members:         parseMembers(os.Getenv(EnvDCDRMembers)),
		MemberList:      memberList(os.Getenv(EnvDCDRMembers)),
		InitLease:       strings.TrimSpace(os.Getenv(EnvDCDRInitLease)),
		PodName:         strings.TrimSpace(os.Getenv(EnvPodName)),
		PodUID:          strings.TrimSpace(os.Getenv(EnvPodUID)),
		PodNamespace:    strings.TrimSpace(os.Getenv(EnvPodNamespace)),
	}
}

// Distributed reports whether cross data center mode is on.
func (e *Env) Distributed() bool {
	return e.DCName != ""
}

// PerPodName is the name of the per-pod Service and ServiceExport, <pod>-<dc>.
func (e *Env) PerPodName() string {
	return fmt.Sprintf("%s-%s", e.PodName, e.DCName)
}

func NewKubeClient() (client.Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create in cluster config")
	}
	clientcmd.Fix(config)

	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kubesliceapi.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(config, client.Options{Scheme: scheme})
}
