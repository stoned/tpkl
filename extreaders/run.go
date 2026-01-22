package extreaders

import (
	"log"

	"github.com/apple/pkl-go/pkl"
)

// Run implements tpkl external readers.
func Run() {
	valsReader, err := NewValsResourceReader()
	if err != nil {
		log.Fatalln(err)
	}

	client, err := pkl.NewExternalReaderClient(func(opts *pkl.ExternalReaderClientOptions) {
		opts.ModuleReaders = append(opts.ModuleReaders, ModuleReader{})
		opts.ResourceReaders = append(opts.ResourceReaders, valsReader)
	})
	if err != nil {
		log.Fatalln(err)
	}

	err = client.Run()
	if err != nil {
		log.Fatalln(err)
	}
}
