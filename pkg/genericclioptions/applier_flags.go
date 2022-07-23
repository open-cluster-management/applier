// Copyright Contributors to the Open Cluster Management project
package genericclioptions

import (
	"github.com/spf13/pflag"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
)

type ApplierFlags struct {
	KubectlFactory cmdutil.Factory
	//if set the resources will be sent to stdout instead of being applied
	DryRun  bool
	Timeout int
}

// NewApplierFlags returns ApplierFlags with default values set
func NewApplierFlags(f cmdutil.Factory) *ApplierFlags {
	return &ApplierFlags{
		KubectlFactory: f,
	}
}

func (f *ApplierFlags) AddFlags(flags *pflag.FlagSet) {
}
