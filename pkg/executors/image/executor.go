/*
Package image defines an executor that produces random image names.
*/
package image

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/ovh/venom"
)

// Name defines the executor name.
const Name = "random-image-name"

// New creates a new instance of the executor.
func New() venom.Executor {
	return &Executor{}
}

// Executor struct for pack.
type Executor struct{}

// Result defines the output of the pack executor.
type Result struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func randSeq() string {
	randomStringLength := 10
	b := make([]rune, randomStringLength)
	for i := range b {
		// #nosec: G404
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Run executes the executor and sets an appropriate output result.
func (Executor) Run(ctx context.Context, step venom.TestStep) (interface{}, error) {
	// transform step to Executor Instance
	res := Result{
		Name: fmt.Sprintf("local.venom.buildpacks.io/pack-test-%s", randSeq()),
	}
	return res, nil
}
