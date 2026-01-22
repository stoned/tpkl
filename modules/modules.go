// Package modules embeds Pkl modules in tpkl
package modules

import (
	"embed"
	"log"
)

//go:embed *.pkl
var embeddedModules embed.FS

// Modules "export" the map of embedded Pkl modules.
func Modules() map[string]string {
	var data []byte

	var err error

	files := []string{"tpkl"}
	mods := map[string]string{}

	for idx := range files {
		data, err = embeddedModules.ReadFile(files[idx] + ".pkl")
		if err != nil {
			log.Fatalf("Error reading embedded module `%s`: %s\n", files[idx], err)
		}

		mods[files[idx]] = string(data)
	}

	return mods
}
