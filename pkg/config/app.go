package config

type App struct {
	Env     string `mapstructure:"env" json:"env" yaml:"env"`
	License string `mapstructure:"license" json:"license" yaml:"license"`
	Port    string `mapstructure:"port" json:"port" yaml:"port"`
	Support string `mapstructure:"support" json:"support" yaml:"support"`
	Group   string `mapstructure:"group" json:"group" yaml:"group"`
}
