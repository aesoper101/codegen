package custom

import (
	"gopkg.in/yaml.v3"
	"os"
)

type GenerateMode string

type BehaviorType string

const (
	Custom  GenerateMode = "custom"
	Package GenerateMode = "package"
	Service GenerateMode = "service"
	Method  GenerateMode = "method"
)

const (
	Skip   BehaviorType = "skip"
	Cover  BehaviorType = "cover"
	Append BehaviorType = "append"
)

func (t GenerateMode) String() string {
	return string(t)
}

func (t GenerateMode) IsValid() bool {
	switch t {
	case Custom, Package, Service, Method:
		return true
	default:
		return false
	}
}

func (t BehaviorType) String() string {
	return string(t)
}

func (t BehaviorType) IsValid() bool {
	switch t {
	case Skip, Cover, Append:
		return true
	default:
		return false
	}
}

type AppendKey string

const (
	AppendKeyMethod  AppendKey = "method"
	AppendKeyService AppendKey = "service"
)

func (t AppendKey) String() string {
	return string(t)
}

func (t AppendKey) IsValid() bool {
	switch t {
	case AppendKeyMethod, AppendKeyService:
		return true
	default:
		return false
	}
}

type Template struct {
	Path           string         `yaml:"path"`            // The generated path and its filename, such as biz/handler/ping.go
	Delims         [2]string      `yaml:"delims"`          // Template Action Instruction Identifier, default: "{{}}"
	Body           string         `yaml:"body"`            // Render template, currently only supports go template syntax
	Type           GenerateMode   `yaml:"type"`            // Template type: file/dir
	UpdateBehavior UpdateBehavior `yaml:"update_behavior"` // Update command behavior; 0:unchanged, 1:regenerate, 2:append
	Extras         Extras         `yaml:"extras"`          // Extra information
}

type Extras map[string]string

func (e Extras) Get(key string) string {
	v, _ := e[key]
	return v
}

type UpdateBehavior struct {
	Type BehaviorType `yaml:"type"` // Update behavior type: skip/cover/append
	// the following variables are used for append update
	AppendKey AppendKey `yaml:"append_key"`         // Append content based in key; for example: 'method'/'service'
	InsertKey string    `yaml:"insert_key"`         // Insert content by "insert_key"
	AppendTpl string    `yaml:"append_content_tpl"` // Append content if UpdateBehavior is "append"
	ImportTpl []string  `yaml:"import_tpl"`         // Import insert template
}

type TemplateConfig struct {
	Layouts []*Template `yaml:"layouts"`
	Extras  Extras      `yaml:"extras"`
}

func loadTemplates(paths ...string) ([]*TemplateConfig, error) {
	var tpls []*TemplateConfig
	for _, path := range paths {
		tplConfig, err := loadTemplate(path)
		if err != nil {
			return nil, err
		}
		tpls = append(tpls, tplConfig)
	}
	return tpls, nil
}

func loadTemplate(path string) (*TemplateConfig, error) {
	var config TemplateConfig
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = yaml.Unmarshal(b, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
