/*
Package executors contains the custom executors for buildpacks-venom CLI.
*/
package executors

import (
	"github.com/ovh/venom/executors"
	"github.com/samj1912/buildpacks-venom/pkg/executors/image"
	"github.com/samj1912/buildpacks-venom/pkg/executors/pack"
)

func init() {
	executors.Registry[pack.Name] = pack.New
	executors.Registry[image.Name] = image.New
}
