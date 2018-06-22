package cmds

import (
	utilerrors "github.com/appscode/go/util/errors"
	v "github.com/appscode/go/version"
	"github.com/golang/glog"
	"github.com/kubevault/unsealer/pkg/worker"
	"github.com/spf13/cobra"
)

func NewCmdRun() *cobra.Command {
	opts := worker.NewWorkerOptions()

	cmd := &cobra.Command{
		Use:               "run",
		Short:             "Launch Vault unsealer",
		DisableAutoGenTag: true,
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
