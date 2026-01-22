// Package extreaders implements external readers for Pkl
package extreaders

// ExternalReadersRegistrationArgs returns arguments to register tpkl's
// external readers when running the `pkl` command.
func ExternalReadersRegistrationArgs(tpkl string) []string {
	return []string{
		"--external-module-reader=" + ModuleScheme + "=" + tpkl + " readers",
		"--external-resource-reader=" + ValsScheme + "=" + tpkl + " readers",
	}
}
