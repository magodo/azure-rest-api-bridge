package ctrl

type Config struct {
	Overrides  []Override  `hcl:"override,block"`
	Executions []Execution `hcl:"execution,block"`
}

type Override struct {
	PathPattern        string          `hcl:"path_pattern,attr"`
	ResponseSelector   string          `hcl:"response_selector,optional"`
	ResponseBody       string          `hcl:"response_body,optional"`
	ResponseMergePatch string          `hcl:"response_merge_patch,optional"`
	ResponseJSONPatch  string          `hcl:"response_json_patch,optional"`
	ExpanderOption     *ExpanderOption `hcl:"expander,block"`
	SynthOption        *SynthOption    `hcl:"synthesizer,block"`
}

type Execution struct {
	Name       string            `hcl:"name,label"`
	Skip       bool              `hcl:"skip,optional"`
	SkipReason string            `hcl:"skip_reason,optional"`
	Overrides  []Override        `hcl:"override,block"`
	Env        map[string]string `hcl:"env,optional"`
	Dir        string            `hcl:"dir,optional"`
	Path       string            `hcl:"path,attr"`
	Args       []string          `hcl:"args,optional"`
}

type SynthOption struct {
	UseEnumValue bool `hcl:"use_enum_value,optional"`
}

type ExpanderOption struct {
	EmptyObjAsStr bool `hcl:"empty_obj_as_str,optional"`
}
