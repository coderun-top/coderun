package yaml

import (
	"io"
	"io/ioutil"
	"os"

	libcompose "github.com/docker/libcompose/yaml"
	"gopkg.in/yaml.v2"
)

type (
	// Config defines a pipeline configuration.
	Config struct {
		// Cache     libcompose.Stringorslice  `yaml:"cache,omitempty"`
		Cache     Cache                  `yaml:"cache,omitempty"`
		Platform  string                 `yaml:"platform,omitempty"`
		// 保留对之前版本的兼容
		Branches  Constraint             `yaml:"branches,omitempty"`
		Trigger   Constraints            `yaml:"trigger,omitempty"`
		Workspace Workspace              `yaml:"workspace,omitempty"`
		Clone     Containers             `yaml:"clone,omitempty"`
		Pipeline  Containers             `yaml:"steps"`
		Services  Containers             `yaml:"services,omitempty"`
		Networks  Networks               `yaml:"networks,omitempty"`
		Volumes   Volumes                `yaml:"volumes,omitempty"`
		Labels    libcompose.SliceorMap  `yaml:"labels,omitempty"`
		Trust     Constraint             `yaml:"trust,omitempty"`
		Agent     string                 `yaml:"agent,omitempty"`
		Vargs     map[string]interface{} `yaml:",inline"`
	}

	// Workspace defines a pipeline workspace.
	Workspace struct {
		Base string
		Path string
	}

	Cache struct {
		Key       string                   `yaml:"key,omitempty"`
		Untracked bool                     `yaml:"untracked,omitempty"`
		Paths     libcompose.Stringorslice `yaml:"paths,omitempty"`
	}
)

// Parse parses the configuration from bytes b.
func Parse(r io.Reader) (*Config, error) {
	out, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return ParseBytes(out)
}

// ParseBytes parses the configuration from bytes b.
func ParseBytes(b []byte) (*Config, error) {
	out := new(Config)
	err := yaml.Unmarshal(b, out)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// ParseString parses the configuration from string s.
func ParseString(s string) (*Config, error) {
	return ParseBytes(
		[]byte(s),
	)
}

// ParseFile parses the configuration from path p.
func ParseFile(p string) (*Config, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f)
}
