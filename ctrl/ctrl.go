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
	ExecFile    string
	MockSrvAddr string
	MockSrvPort int
}

type Ctrl struct {
	MockSrvAddr string
	MockSrvPort int
	ExecSpec    ExecSpec
}

func NewCtrl(opt Option) (*Ctrl, error) {
	parser := hclparse.NewParser()
	f, diags := parser.ParseHCLFile(opt.ExecFile)
	if diags.HasErrors() {
		return nil, fmt.Errorf("parsing %s: %v", opt.ExecFile, diags.Error())
	}
	var execSpec ExecSpec
	ctx := &hcl.EvalContext{
		Variables: map[string]cty.Value{
			"server_addr": cty.StringVal(fmt.Sprintf("%s:%d", opt.MockSrvAddr, opt.MockSrvPort)),
		},
	}
	if diags := gohcl.DecodeBody(f.Body, ctx, &execSpec); diags.HasErrors() {
		return nil, fmt.Errorf("decoding %s: %v", opt.ExecFile, diags.Error())
	}

	return &Ctrl{
		MockSrvAddr: opt.MockSrvAddr,
		MockSrvPort: opt.MockSrvPort,
		ExecSpec:    execSpec,
	}, nil
}

func (ctrl *Ctrl) Run(ctx context.Context) error {
	// Start mock server
	srv, closeCh, err := mockserver.Serve(fmt.Sprintf("%s:%d", ctrl.MockSrvAddr, ctrl.MockSrvPort))
	if err != nil {
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

		log.Debug("command", "path", execution.Path, "args", execution.Args, "env", env, "dir", execution.Dir)

		if err := cmd.Run(); err != nil {
			log.Error("run failure", "stdout", stdout.String(), "stderr", stderr.String())
			return fmt.Errorf("running execution %q: %v", execution.Name, err)
		}

		log.Info(stdout.String())
	}

	// Stop mock server
	if err := srv.Shutdown(ctx); err != nil {
		return err
	}
	<-closeCh

	return nil
}
