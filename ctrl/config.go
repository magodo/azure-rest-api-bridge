package ctrl

type Config struct {
	Overrides  []Override  `hcl:"override,block"`
	Executions []Execution `hcl:"execution,block"`
}

type Override struct {
	PathPattern           string            `hcl:"path_pattern,attr"`
	ResponseSelectorMerge string            `hcl:"response_selector_merge,optional"`
	ResponseSelectorJSON  string            `hcl:"response_selector_json,optional"`
	ResponseBody          string            `hcl:"response_body,optional"`
	ResponsePatchMerge    string            `hcl:"response_patch_merge,optional"`
	ResponsePatchJSON     string            `hcl:"response_patch_json,optional"`
	ResponseHeader        map[string]string `hcl:"response_header,optional"`
	ExpanderOption        *ExpanderOption   `hcl:"expander,block"`
	SynthOption           *SynthOption      `hcl:"synthesizer,block"`
}

type Execution struct {
	Name string `hcl:"name,label"`
	Type string `hcl:"type,label"`

	Skip       bool              `hcl:"skip,optional"`
	SkipReason string            `hcl:"skip_reason,optional"`
	Overrides  []Override        `hcl:"override,block"`
	Env        map[string]string `hcl:"env,optional"`
	Dir        string            `hcl:"dir,optional"`
	Path       string            `hcl:"path,attr"`
	Args       []string          `hcl:"args,optional"`
}

func (exec Execution) String() string {
	return exec.Name + "." + exec.Type
}

type SynthOption struct {
	UseEnumValue     bool               `hcl:"use_enum_value,optional"`
	DuplicateElement []DuplicateElement `hcl:"duplicate_element,block"`
}

type ExpanderOption struct {
	EmptyObjAsStr bool `hcl:"empty_obj_as_str,optional"`
	DisableCache  bool `hcl:"disable_cache,optional"`
}

type DuplicateElement struct {
	Count *int   `hcl:"count,optional"`
	Addr  string `hcl:"addr,attr"`
}
