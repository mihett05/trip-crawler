package config

const (
	EnvProduction  = "production"
	EnvDevelopment = "development"
)

type AppConfig struct {
	Name        string `env:"NAME"`
	Environment string `env:"ENV" envDefault:"production"`
}
