package config

// ServerConfig SB GOLANG HAS NO GENERIC TYPE
type ServerConfig struct {
	Name      string
	Namespace string
	LogLevel  string `yaml:"logLevel"`
}

type DatabaseConfig struct {
	Mysql MysqlConfig
}

type MysqlConfig struct {
	Db    string
	Table string
	Host  string
	User  string
	Pwd   string
}
