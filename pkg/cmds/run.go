package cmds

import (
	utilerrors "github.com/appscode/go/util/errors"
	v "github.com/appscode/go/version"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"kmodules.xyz/client-go/tools/cli"
	"kubevault.dev/unsealer/pkg/worker"
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
			glog.Infof("Starting operator version %s+%s ...", v.Version.Version, v.Version.CommitHash)

			if errs := opts.Validate(); errs != nil {
				return utilerrors.NewAggregate(errs)
			}
			return opts.Run()
		},
	}

	opts.AddFlags(cmd.Flags())

	return cmd
}
