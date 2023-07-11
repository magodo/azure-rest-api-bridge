package ctrl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/magodo/azure-rest-api-bridge/log"
	"github.com/magodo/azure-rest-api-bridge/mockserver"
	"github.com/magodo/azure-rest-api-bridge/mockserver/swagger"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/function/stdlib"
)

type Option struct {
	ConfigFile    string
	ContinueOnErr bool
	ServerOption  mockserver.Option
}

type Ctrl struct {
	ExecSpec      Config
	ContinueOnErr bool
	MockServer    mockserver.Server
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
		Functions: map[string]function.Function{
			"jsonencode": stdlib.JSONEncodeFunc,
		},
		Variables: map[string]cty.Value{
			"home":        cty.StringVal(homedir),
			"server_addr": cty.StringVal(fmt.Sprintf("%s:%d", opt.ServerOption.Addr, opt.ServerOption.Port)),
		},
	}
	if diags := gohcl.DecodeBody(f.Body, ctx, &execSpec); diags.HasErrors() {
		return nil, fmt.Errorf("decoding %s: %v", opt.ConfigFile, diags.Error())
	}

	if err := validateExecSpec(execSpec); err != nil {
		return nil, fmt.Errorf("invalid exec spec: %v", err)
	}

	srv, err := mockserver.New(opt.ServerOption)
	if err != nil {
		return nil, fmt.Errorf("creating mock server: %v", err)
	}

	return &Ctrl{
		ExecSpec:      execSpec,
		ContinueOnErr: opt.ContinueOnErr,
		MockServer:    *srv,
	}, nil
}

func validateExecSpec(spec Config) error {
	validateOverride := func(ovs []Override) error {
		for _, ov := range ovs {
			if ov.ResponseBody+ov.ResponseSelector+ov.ResponseJSONPatch+ov.ResponseMergePatch == "" && ov.ExpanderOption == nil && ov.SynthOption == nil {
				return fmt.Errorf("empty override block is not allowed")
			}
			if ov.ResponseBody != "" {
				if ov.ResponseSelector+ov.ResponseJSONPatch+ov.ResponseMergePatch != "" || ov.ExpanderOption != nil || ov.SynthOption != nil {
					return fmt.Errorf("`response_body` can only be exclusive specified")
				}
				continue
			}
			if ov.ResponseJSONPatch != "" && ov.ResponseMergePatch != "" {
				return fmt.Errorf("`response_merge_patch` conflicts with `response_json_patch`")
			}
		}
		return nil
	}

	if err := validateOverride(spec.Overrides); err != nil {
		return err
	}

	execNames := make(map[string]bool)
	for _, exec := range spec.Executions {
		if exec.Skip && exec.SkipReason == "" {
			return fmt.Errorf("skipped execution %s must have a skip_reason", exec.Name)
		}
		if err := validateOverride(exec.Overrides); err != nil {
			return err
		}
		if _, exist := execNames[exec.Name]; exist {
			return fmt.Errorf("duplicated execution %s", exec.Name)
		}
		execNames[exec.Name] = true
	}

	return nil
}

func (ctrl *Ctrl) Run(ctx context.Context) error {
	// Start mock server
	log.Info("Starting the mock server")
	if err := ctrl.MockServer.Start(); err != nil {
		return err
	}

	outputs := make(map[string]interface{})

	execTotal := len(ctrl.ExecSpec.Executions)
	execSucceed := 0
	execFail := 0

	// Launch each execution
	for i, execution := range ctrl.ExecSpec.Executions {
		run := func(execution Execution) error {
			if execution.Skip {
				log.Info(fmt.Sprintf("Skipping %s (%d/%d): %s", execution.Name, i+1, execTotal, execution.SkipReason))
				return nil
			}

			overrides := append([]Override{}, execution.Overrides...)
			overrides = append(overrides, ctrl.ExecSpec.Overrides...)

			var ovs []mockserver.Override
			for _, override := range overrides {
				ov := mockserver.Override{
					PathPattern:        *regexp.MustCompile(override.PathPattern),
					ResponseSelector:   override.ResponseSelector,
					ResponseBody:       override.ResponseBody,
					ResponseMergePatch: override.ResponseMergePatch,
					ResponseJSONPatch:  override.ResponseJSONPatch,
				}
				if opt := override.SynthOption; opt != nil {
					ov.SynthOption = &swagger.SynthesizerOption{
						UseEnumValues: opt.UseEnumValue,
					}
				}
				if opt := override.ExpanderOption; opt != nil {
					ov.ExpanderOption = &swagger.ExpanderOption{
						EmptyObjAsStr: opt.EmptyObjAsStr,
					}
				}

				ovs = append(ovs, ov)
			}

			ctrl.MockServer.InitExecution(ovs)

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

			log.Info(fmt.Sprintf("Executing %s (%d/%d)", execution.Name, i+1, execTotal))

			log.Debug("execution detail", "path", execution.Path, "args", execution.Args, "env", env, "dir", execution.Dir)

			if err := cmd.Run(); err != nil {
				log.Error("run failure", "stdout", stdout.String(), "stderr", stderr.String())
				return fmt.Errorf("running execution %q: %v", execution.Name, err)
			}

			log.Debug("execution result", "stdout", stdout.String())

			var appModel interface{}
			if err := json.Unmarshal(stdout.Bytes(), &appModel); err != nil {
				log.Error("post-execution unmarshal failure", "error", err, "stdout", stdout.String())
				return fmt.Errorf("post-execution %q unmarshal: %v", execution.Name, err)
			}

			m, err := MapModels(appModel, ctrl.MockServer.Records()...)
			if err != nil {
				log.Error("post-execution map models", "error", err)
				return fmt.Errorf("post-execution %q map models: %v", execution.Name, err)
			}

			if err := m.AddLink(ctrl.MockServer.Idx.Commit, ctrl.MockServer.Specdir); err != nil {
				log.Error("post-execution model map adding link", "error", err)
				return fmt.Errorf("post-execution model map adding link: %v", err)
			}

			outputs[execution.Name] = m

			return nil
		}

		if err := run(execution); err != nil {
			if ctrl.ContinueOnErr {
				execFail++
				continue
			}
			return err
		} else {
			execSucceed++
		}
	}

	if ctrl.ContinueOnErr {
		log.Info("Summary", "total", execTotal, "succeed", execSucceed, "fail", execFail)
	}

	b, err := json.MarshalIndent(outputs, "", "  ")
	if err != nil {
		log.Error("post-execution marshalling map", "error", err)
		return fmt.Errorf("post-execution marshalling map: %v", err)
	}

	fmt.Println(string(b))

	// Stop mock server
	log.Info("Stopping the mock server")
	if err := ctrl.MockServer.Stop(ctx); err != nil {
		return err
	}

	return nil
}
