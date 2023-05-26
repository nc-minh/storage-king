package utils

import (
	"github.com/spf13/viper"
)

type Web struct {
	ClientID                string   `json:"client_id"`
	ProjectID               string   `json:"project_id"`
	AuthURI                 string   `json:"auth_uri"`
	TokenURI                string   `json:"token_uri"`
	AuthProviderX509CertURL string   `json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `json:"client_secret"`
	RedirectURIs            []string `json:"redirect_uris"`
	JavaScriptOrigins       []string `json:"javascript_origins"`
}

type Config struct {
	Env               string `mapstructure:"ENV"`
	Port              string `mapstructure:"PORT"`
	DBDriver          string `mapstructure:"DB_DRIVER"`
	DBSource          string `mapstructure:"DB_SOURCE"`
	MigrationURL      string `mapstructure:"MIGRATION_URL"`
	HTTPServerAddress string `mapstructure:"HTTP_SERVER_ADDRESS"`
	TokenSymmetricKey string `mapstructure:"TOKEN_SYMMETRIC_KEY"`

	ClientID                string   `mapstructure:"CLIENT_ID" json:"client_id"`
	ProjectID               string   `mapstructure:"PROJECT_ID" json:"project_id"`
	AuthURI                 string   `mapstructure:"AUTH_URI" json:"auth_uri"`
	TokenURI                string   `mapstructure:"TOKEN_URI" json:"token_uri"`
	AuthProviderX509CertURL string   `mapstructure:"AUTH_PROVIDER_X509_CERT_URL" json:"auth_provider_x509_cert_url"`
	ClientSecret            string   `mapstructure:"CLIENT_SECRET" json:"client_secret"`
	RedirectURIs            []string `mapstructure:"REDIRECT_URIS" json:"redirect_uris"`
	JavaScriptOrigins       []string `mapstructure:"JAVASCRIPT_ORIGINS" json:"javascript_origins"`
	Web                     Web      `json:"web"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	config.Web = Web{
		ClientID:                config.ClientID,
		ProjectID:               config.ProjectID,
		AuthURI:                 config.AuthURI,
		TokenURI:                config.TokenURI,
		AuthProviderX509CertURL: config.AuthProviderX509CertURL,
		ClientSecret:            config.ClientSecret,
		RedirectURIs:            config.RedirectURIs,
		JavaScriptOrigins:       config.JavaScriptOrigins,
	}

	return config, err

}
