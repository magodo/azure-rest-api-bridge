package ctrl

type Config struct {
	Overrides  []Override  `hcl:"override,block"`
	Executions []Execution `hcl:"execution,block"`
}

type Override struct {
	PathPattern        string `hcl:"path_pattern,attr"`
	ResponseBody       string `hcl:"response_body,optional"`
	ResponseMergePatch string `hcl:"response_merge_patch,optional"`
	ResponseJSONPatch  string `hcl:"response_json_patch,optional"`
}

type Execution struct {
	Name      string            `hcl:"name,label"`
	Overrides []Override        `hcl:"override,block"`
	Env       map[string]string `hcl:"env,optional"`
	Dir       string            `hcl:"dir,optional"`
	Path      string            `hcl:"path,attr"`
	Args      []string          `hcl:"args,optional"`
}
