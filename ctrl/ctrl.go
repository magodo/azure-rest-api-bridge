package ctrl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver"
	"github.com/zclconf/go-cty/cty"
)

type Option struct {
	ConfigFile   string
	ServerOption mockserver.Option
}

type Ctrl struct {
	ExecSpec   Config
	MockServer mockserver.Server
}

func NewCtrl(opt Option) (*Ctrl, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(opt.ConfigFile)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing %s: %v", opt.ConfigFile, diags.Error())
	}

	homedir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("getting user home dir: %v", err)
	}
	var execSpec Config
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"home":        cty.StringVal(homedir),
			"server_addr": cty.StringVal(fmt.Sprintf("%s:%d", opt.ServerOption.Addr, opt.ServerOption.Port)),
		},
	}
	if diags := gohcl.DecodeBody(f.Body, ctx, &execSpec); diags.HasErrors() {
		return nil, fmt.Errorf("decoding %s: %v", opt.ConfigFile, diags.Error())
	}

	srv, err := mockserver.New(opt.ServerOption)
	if err != nil {
		return nil, fmt.Errorf("creating mock server: %v", err)
	}

	return &Ctrl{
		ExecSpec:   execSpec,
		MockServer: *srv,
	}, nil
}

func (ctrl *Ctrl) Run(ctx context.Context) error {
	// Start mock server
	log.Info("Starting the mock server")
	if err := ctrl.MockServer.Start(); err != nil {
		return err
	}

	// Launch each execution
	for _, execution := range ctrl.ExecSpec.Executions {
		overrides := append([]Override{}, execution.Overrides...)
		overrides = append(overrides, ctrl.ExecSpec.Overrides...)

		env := os.Environ()
		for k, v := range execution.Env {
			env = append(env, k+"="+v)
		}

		var stdout bytes.Buffer
		var stderr bytes.Buffer

		cmd := exec.Cmd{
			Path:   execution.Path,
			Args:   append([]string{filepath.Base(execution.Path)}, execution.Args...),
			Env:    env,
			Dir:    execution.Dir,
			Stdout: &stdout,
			Stderr: &stderr,
		}

		log.Info(fmt.Sprintf("Executing %s", execution.Name))

		log.Debug("execution detail", "path", execution.Path, "args", execution.Args, "env", env, "dir", execution.Dir)

		if err := cmd.Run(); err != nil {
			log.Error("run failure", "stdout", stdout.String(), "stderr", stderr.String())
			return fmt.Errorf("running execution %q: %v", execution.Name, err)
		}

		log.Info("stdout", "message", stdout.String())
	}

	// Stop mock server
	log.Info("Stopping the mock server")
	if err := ctrl.MockServer.Stop(ctx); err != nil {
		return err
	}

	return nil
}
