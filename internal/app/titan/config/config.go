package config

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
	"os"
	"strings"
)

const (
	envPrefix  = "GUS_PS"
	configName = "config"
)

var C *Config

// Main config group
type Config struct {
	Env      	string
	Log      	Logging
	Database 	Database
	Grpc     	Grpc
	Http     	Http
	ServiceNow 	ServiceNow
	Cors     	Cors
	OAuthCredentials OAuthCredentials
}

// subgroups
type Logging struct {
	Level             string
	Encoding          string
	IncludeCaller     bool
	IncludeStacktrace bool
}

type Database struct {
	Host               string
	Port               int
	DBName             string
	User               string
	Password           string
	MaxOpenConnections int
	MaxIdleConnections int
}

type OAuthCredentials struct {
	WebTitanBaseUrl string
	ConsumerKey     string
	ConsumerSecret  string
	TokenKey        string
	TokenSecret     string
}

type Grpc struct {
	GatewayAddr string
	ListenAddr  string
}

type Http struct {
	ListenAddr string
}

type ServiceNow struct {
	ApiURL string
	Username string
	Password string
	Timeout string
}

type Cors struct {
	AllowedOriginsCSV string
	AllowedHeadersCSV string
	AllowedMethodsCSV string
}

// Initialize the global config
func init() {

	// use the folder from env var, or default to ./configs
	// 	* this is needed for the tests
	configFolder := os.Getenv("CONFIG_FOLDER")
	if configFolder == "" {
		configFolder = "./configs"
		// For Testing Locally
		//configFolder = "./../../../../configs"
	}

	var err error
	C, err = NewConfig(configFolder)
	if err != nil {
		panic("Error creating global config; exception: " + err.Error())
	} else {
		printConfig()
	}
}

// NewConfig creates a new config
func NewConfig(configFolder string) (*Config, error) {

	// from the environment
	viper.SetEnvPrefix(envPrefix)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match

	// from a config file
	viper.SetConfigName(configName)
	viper.AddConfigPath(configFolder)

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return nil, errors.Wrap(err, "Error reading config file")
	}

	// Parse variables to Config struct
	var c *Config
	if err := viper.Unmarshal(&c); err != nil {
		return nil, errors.Wrap(err, "Error loading Configurations")
	}

	return c, nil
}

func printConfig() {
	y, err := yaml.Marshal(C)
	if err != nil {
		panic(fmt.Sprintf("Unable to log the config [%+v]; exception: %s", C, err.Error()))
	}
	fmt.Println("============== CONFIG ==============")
	fmt.Println()
	fmt.Println(string(y))
	fmt.Println("============== CONFIG ==============")
	fmt.Println()
}
