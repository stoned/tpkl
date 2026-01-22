// main
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

func main() {
	var doc any

	decoder := json.NewDecoder(os.Stdin)

	if err := decoder.Decode(&doc); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
