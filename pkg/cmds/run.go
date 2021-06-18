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

package cmds

import (
	"kubevault.dev/unsealer/pkg/worker"

	"github.com/spf13/cobra"
	utilerrors "gomodules.xyz/errors"
	v "gomodules.xyz/x/version"
	"k8s.io/klog/v2"
	"kmodules.xyz/client-go/tools/cli"
)

func NewCmdRun() *cobra.Command {
	opts := worker.NewWorkerOptions()

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch Vault unsealer",
		DisableAutoGenTag: true,
		PreRun: func(c *cobra.Command, args []string) {
			cli.SendPeriodicAnalytics(c, v.Version.Version)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			klog.Infof("Starting operator version %s+%s ...", v.Version.Version, v.Version.CommitHash)

			if errs := opts.Validate(); errs != nil {
				return utilerrors.NewAggregate(errs)
			}
			return opts.Run()
		},
	}

	opts.AddFlags(cmd.Flags())
	opts.AddEnvVars()
	return cmd
}
