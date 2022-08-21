/*
Package pack defines the pack venom executor to build container images to the local docker daemon.
*/
package pack

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"path"
	"strconv"

	"github.com/mitchellh/mapstructure"
	"github.com/ovh/venom"
)

// Name defines the executor name.
const Name = "pack"

// New creates a new instance of the pack executor.
func New() venom.Executor {
	return &Executor{}
}

// Executor struct for pack.
type Executor struct {
	ImageName     string            `json:"image-name,omitempty" yaml:"image-name,omitempty"`
	Builder       string            `json:"builder,omitempty" yaml:"builder,omitempty"`
	Buildpacks    []string          `json:"buildpacks,omitempty" yaml:"buildpacks,omitempty"`
	ClearCache    bool              `json:"clear-cache,omitempty" yaml:"clear-cache,omitempty"`
	Env           map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	ExtraArgs     []string          `json:"extra-args,omitempty" yaml:"extra-args,omitempty"`
	GID           *int              `json:"gid,omitempty" yaml:"gid,omitempty"`
	Network       string            `json:"network,omitempty" yaml:"network,omitempty"`
	NoColor       bool              `json:"no-color,omitempty" yaml:"no-color,omitempty"`
	NoPull        bool              `json:"no-pull,omitempty" yaml:"no-pull,omitempty"`
	PackBinary    string            `json:"pack-binary,omitempty" yaml:"pack-binary,omitempty"`
	Path          string            `json:"path,omitempty" yaml:"path,omitempty"`
	PullPolicy    string            `json:"pull-policy,omitempty" yaml:"pull-policy,omitempty"`
	SBOMOutputDir string            `json:"sbom-output-dir,omitempty" yaml:"sbom-output-dir,omitempty"`
	TrustBuilder  bool              `json:"trust-builder,omitempty" yaml:"trust-builder,omitempty"`
	Verbose       bool              `json:"verbose,omitempty" yaml:"verbose,omitempty"`
	Volumes       []string          `json:"volumes,omitempty" yaml:"volumes,omitempty"`
}

// Result defines the output of the pack executor.
type Result struct {
	Code      int                    `json:"code,omitempty" yaml:"code,omitempty"`
	Command   string                 `json:"command,omitempty" yaml:"command,omitempty"`
	Systemout string                 `json:"systemout,omitempty" yaml:"systemout,omitempty"`
	Systemerr string                 `json:"systemerr,omitempty" yaml:"systemerr,omitempty"`
	ImageInfo map[string]interface{} `json:"image-info,omitempty" yaml:"image-info,omitempty"`
}

func (e Executor) generateArgs() []string {
	args := []string{"build", e.ImageName}
	if e.Builder != "" {
		args = append(args, "--builder", e.Builder)
	}
	for _, bp := range e.Buildpacks {
		args = append(args, "-b", bp)
	}
	if e.ClearCache {
		args = append(args, "--clear-cache")
	}
	for name, value := range e.Env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", name, value))
	}
	args = append(args, e.ExtraArgs...)
	if e.GID != nil {
		args = append(args, "--gid", strconv.Itoa(*e.GID))
	}
	if e.Network != "" {
		args = append(args, "--network", e.Network)
	}
	if e.NoColor {
		args = append(args, "--no-color")
	}
	if e.NoPull {
		args = append(args, "--no-pull")
	}
	if e.Path != "" {
		args = append(args, "-p", e.Path)
	}
	if e.PullPolicy != "" {
		args = append(args, "--pull-policy", e.PullPolicy)
	}
	if e.SBOMOutputDir != "" {
		args = append(args, "--sbom-output-dir", e.SBOMOutputDir)
	}
	if e.TrustBuilder {
		args = append(args, "--trust-builder")
	}
	if e.Verbose {
		args = append(args, "--verbose")
	}
	for _, vol := range e.Volumes {
		args = append(args, "--volume", vol)
	}
	return args
}

// GenerateCommand generates the appropriate pack build command based on the input.
func (e Executor) GenerateCommand(ctx context.Context) (cmd *exec.Cmd, err error) {
	packBinary := e.PackBinary
	if packBinary == "" {
		packBinary = "pack"
	}
	packBinary = path.Clean(packBinary)
	if packBinary, err = exec.LookPath(packBinary); err != nil {
		return nil, fmt.Errorf("unable to find pack the pack binary: %w", err)
	}
	if e.ImageName == "" {
		return nil, fmt.Errorf("image-name must be defined")
	}
	// #nosec: G204
	cmd = exec.Command(packBinary, e.generateArgs()...)
	cmd.Stdout = bytes.NewBuffer(nil)
	cmd.Stderr = bytes.NewBuffer(nil)
	cmd.Dir = venom.StringVarFromCtx(ctx, "venom.testsuite.workdir")
	return cmd, nil
}

// GenerateImageInfo runs pack inspect on the output image and returns the output JSON.
func (e Executor) GenerateImageInfo() (info map[string]interface{}, err error) {
	packBinary := e.PackBinary
	if packBinary == "" {
		packBinary = "pack"
	}
	packBinary = path.Clean(packBinary)
	if packBinary, err = exec.LookPath(packBinary); err != nil {
		return nil, fmt.Errorf("unable to find pack the pack binary: %w", err)
	}
	if e.ImageName == "" {
		return nil, fmt.Errorf("image-name must be defined")
	}
	// #nosec: G204
	cmd := exec.Command(packBinary, "inspect", e.ImageName, "--output", "json", "-q")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	v := struct {
		Info map[string]interface{} `json:"local_info"`
	}{}
	if err := venom.JSONUnmarshal(output, &v); err != nil {
		return nil, fmt.Errorf("unable to extract local_info from pack inspect output: %w", err)
	}
	return v.Info, nil
}

// Run executes the executor and sets an appropriate output result.
func (Executor) Run(ctx context.Context, step venom.TestStep) (interface{}, error) {
	// transform step to Executor Instance
	var e Executor
	if err := mapstructure.Decode(step, &e); err != nil {
		return nil, err
	}
	command, err := e.GenerateCommand(ctx)
	stdoutBytes := bytes.NewBuffer(nil)
	stderrBytes := bytes.NewBuffer(nil)
	command.Stderr = stderrBytes
	command.Stdout = stdoutBytes
	res := Result{}
	if err != nil {
		return res, fmt.Errorf("unable to execute pack: %w", err)
	}
	venom.Info(ctx, "running pack command: %s", command.String())
	if e := command.Run(); e != nil {
		if exitError, ok := e.(*exec.ExitError); ok {
			res.Code = exitError.ExitCode()
			venom.Error(ctx, "pack build failed: %s", e.Error())
		}
	}
	res.Systemout = stdoutBytes.String()
	res.Systemerr = stderrBytes.String()
	res.Command = command.String()
	if res.Code == 0 {
		if res.ImageInfo, err = e.GenerateImageInfo(); err != nil {
			return Result{}, fmt.Errorf("unable to inspect image: %w", err)
		}
	}
	return res, nil
}

// GetDefaultAssertions return default assertions for type pack.
func (Executor) GetDefaultAssertions() *venom.StepAssertions {
	return &venom.StepAssertions{Assertions: []venom.Assertion{"result.code ShouldEqual 0", "result.systemerr ShouldEqual ''"}}
}
