package hertz

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aesoper101/codegen/internal/generator"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser/protobuf"
	"github.com/aesoper101/codegen/internal/generator/hertz/parser/thrift"
	"github.com/aesoper101/x/configx"
)

type GenerateMode string

type BehaviorType string

const (
	Custom  GenerateMode = "custom"
	Package GenerateMode = "package"
	IDL     GenerateMode = "idl"
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
	case Custom, Package, Service, Method, IDL:
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
	Path           string         `json:"path"`            // The generated path and its filename, such as biz/handler/ping.go
	Delims         [2]string      `json:"delims"`          // Template Action Instruction Identifier, default: "{{}}"
	Body           string         `json:"body"`            // Render template, currently only supports go template syntax
	Type           GenerateMode   `json:"type"`            // Template type: file/dir
	UpdateBehavior UpdateBehavior `json:"update_behavior"` // Update command behavior; 0:unchanged, 1:regenerate, 2:append
	Extras         Extras         `json:"extras"`          // Extra information
}

func (tpl *Template) Preprocess() {
	if tpl.Extras != nil {
		tpl.Extras = Extras{}
	}
	if tpl.Type == "" {
		tpl.Type = Custom
	}
	if tpl.UpdateBehavior.Type == "" {
		tpl.UpdateBehavior.Type = Skip
	}
}

type UpdateBehavior struct {
	Type BehaviorType `json:"type"` // Update behavior type: skip/cover/append
	// the following variables are used for append update
	AppendKey AppendKey `json:"append_key"`         // Append content based in key; for example: 'method'/'service'
	InsertKey string    `json:"insert_key"`         // Insert content by "insert_key"
	AppendTpl string    `json:"append_content_tpl"` // Append content if UpdateBehavior is "append"
	ImportTpl []string  `json:"import_tpl"`         // Import insert template
}

type GenerateConfig struct {
	Templates []*Template       `json:"templates"`
	GoModule  string            `json:"go_module"`
	ModuleDir string            `json:"module_dir"`
	IDLType   generator.IdlType `json:"idl_type"`
	Files     []string          `json:"files"` // idl files, theses files will be used to generate the code
	Extras    Extras            `json:"extras"`
	Options   parser.Options    `json:"options"`
}

func (config *GenerateConfig) UnmarshalJSON(data []byte) error {
	type Data struct {
		IDLType generator.IdlType `json:"idl_type"`
		Options json.RawMessage
	}

	var d Data
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}

	switch d.IDLType {
	case generator.Thrift:
		opts := thrift.DefaultOptions()
		if err := json.Unmarshal(d.Options, &opts); err != nil {
			return err
		}
		config.Options = &opts
	case generator.Proto:
		opts := protobuf.DefaultOptions()
		if err := json.Unmarshal(d.Options, &opts); err != nil {
			return err
		}
		config.Options = &opts
	default:
		return fmt.Errorf("unsupported idl type: %s", d.IDLType)
	}
	return nil
}

func loadConfig(ctx context.Context, files ...string) (*GenerateConfig, error) {
	configProvider, err := configx.New(ctx, configx.EnableEnvLoading(""), configx.WithConfigFiles(files...))
	if err != nil {
		return nil, err
	}

	var config GenerateConfig

	if err := configProvider.Unmarshal("", &config); err != nil {
		return nil, err
	}

	return &config, nil
}
