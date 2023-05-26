package ctrl

type ExecSpec struct {
	Overrides  []Override  `hcl:"override,block"`
	Executions []Execution `hcl:"execution,block"`
}

type Override struct {
	PathPattern string `hcl:"path_pattern,attr"`
	Response    string `hcl:"response,attr"`
}

type Execution struct {
	Name      string            `hcl:"name,label"`
	Overrides []Override        `hcl:"override,block"`
	Env       map[string]string `hcl:"env,optional"`
	Dir       string            `hcl:"dir,optional"`
	Path      string            `hcl:"path,attr"`
	Args      []string          `hcl:"args,optional"`
}
