package cmd

import (
	"github.com/spf13/cobra"
)

func addEnvFlag(cmd *cobra.Command, variable *[]string) {
	cmd.Flags().StringArrayVarP(variable, "env-var", "e", nil, "Set environment variable `name[=value]` (repeatable)")
}

func addModuleFlag(cmd *cobra.Command, variable *string) {
	cmd.Flags().StringVarP(variable, "module", "m", "", "Set Pkl module path or URI")
}

func addPropertyFlag(cmd *cobra.Command, variable *[]string) {
	cmd.Flags().StringArrayVarP(variable, "property", "p", nil, "Set external property `name[=value]` (repeatable)")
}

func addVerboseFlag(cmd *cobra.Command, variable *int) {
	cmd.Flags().CountVarP(variable, "v", "v", "Set log `level[=+1]` verbosity (repeatable)")
}
