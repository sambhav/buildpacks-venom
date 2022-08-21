// Buildpacks Venom CLI.
package main

import (
	"github.com/ovh/venom/cmd/venom/root"
	_ "github.com/samj1912/buildpacks-venom/pkg/executors"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := root.New().Execute(); err != nil {
		log.Fatalf("Err:%s", err)
	}
}
