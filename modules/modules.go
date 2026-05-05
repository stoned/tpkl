// Package modules embeds Pkl modules in tpkl
package modules

import (
	"embed"
	"io/fs"
	"log"
	"strings"
)

//go:embed *.pkl
var embeddedModules embed.FS

// Modules "export" the map of embedded Pkl modules.
func Modules() map[string]string {
	mods := map[string]string{}

	err := fs.WalkDir(embeddedModules, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() {
			data, err := fs.ReadFile(embeddedModules, path)
			if err != nil {
				return err //nolint:wrapcheck
			}

			mods[strings.TrimSuffix(path, ".pkl")] = string(data)
		}

		return nil
	})
	if err != nil {
		log.Fatalf("Error reading embedded modules: %s\n", err)
	}

	return mods
}
