package config

type Telegram struct {
	Token          string  `mapstructure:"token" json:"token" yaml:"token"`
	TronScanApiKey string  `mapstructure:"tron_scan_api_key" json:"tron_scan_api_key" yaml:"tron_scan_api_key"`
	GridApiKey     string  `mapstructure:"grid_api_key" json:"grid_api_key" yaml:"grid_api_key"`
	AliasKey       string  `mapstructure:"alias_key" yaml:"alias_key"`
	PrivateKey     string  `mapstructure:"private_key" yaml:"private_key"`
	ReceiveAddress string  `mapstructure:"receive_address" yaml:"receive_address"`
	SendAddress    string  `mapstructure:"send_address" yaml:"send_address"`
	EnableApi      string  `mapstructure:"enable_api" yaml:"enable_api"`
	Ratio          float64 `mapstructure:"ratio" yaml:"ratio"`
	ApiID          int     `mapstructure:"api_id" yaml:"api_id"`
	AppKey         string  `mapstructure:"app_key" yaml:"app_key"`
}
