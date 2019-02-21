package foreman

import (
	"log"
	"net/url"
	"os"

	"github.com/wayfair/terraform-provider-foreman/foreman/api"
	logger "github.com/wayfair/terraform-provider-utils/log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
	"github.com/hashicorp/terraform/terraform"
)

// Environment variables the provider recognizes for configuration
const (
	// Environment variable to configure the provider_loglevel attribute
	ProviderLogLevelEnv string = "FOREMAN_PROVIDER_LOGLEVEL"
	// Environment variable to configure the provider_logfile attribute
	ProviderLogFileEnv string = "FOREMAN_PROVIDER_LOGFILE"
	// Environment variable to configure the client_username attribute
	ClientUsernameEnv string = "FOREMAN_CLIENT_USERNAME"
	// Environment variable to configure the client_password attribute
	ClientPasswordEnv string = "FOREMAN_CLIENT_PASSWORD"
)

// Provider configuration default values
const (
	// Default log level if one is not provided
	DefaultProviderLogLevel string = "INFO"
	// Default output log file if one is not provided
	DefaultProviderLogFile string = "terraform-provider-foreman.log"
)

// Log file constants
const (
	// Specifying the log file as "-" preserves the standard behavior of the
	// Golang stdlib log package.
	LogFileStdLog string = "-"
)

// Configuration options for the provider logging
type LoggingConfig struct {
	// The log level to use
	LogLevel logger.LogLevel
	// The path to the log file
	LogFile string
}

// Provider : Defines params for provider in terraform and available resources
func Provider() terraform.ResourceProvider {
	return &schema.Provider{

		Schema: map[string]*schema.Schema{

			// -- Provider Logging --

			"provider_loglevel": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					ProviderLogLevelEnv,
					DefaultProviderLogLevel,
				),
				ValidateFunc: validation.StringInSlice([]string{
					"DEBUG",
					"TRACE",
					"INFO",
					"WARNING",
					"ERROR",
					"NONE",
					// NOTE(ALL): false - do not ignore case when comparing values
				}, false),
				Description: "The level of verbosity for the provider's log file. This " +
					"setting determines which types of log messages are written and which " +
					"are ignored. Possible values (from most verbose to least verbose) " +
					"include 'DEBUG', 'TRACE', 'INFO', 'WARNING', 'ERROR', and 'NONE'.  The " +
					"provider's logs will be written to the location specified by " +
					"`provider_logfile`. This can also be set through the environment " +
					"variable `FOREMAN_PROVIDER_LOGLEVEL`. Defaults to `'INFO'`.",
			},
			"provider_logfile": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					ProviderLogFileEnv,
					DefaultProviderLogFile,
				),
				Description: "Where to deirect provider-specific log output. " +
					"A value of '-' preserves the default behavior of the log package " +
					"from Golang stdlib and will be combined with the main terraform.log " +
					"file produced by terraform. If the desired output file does not " +
					"exist, it will be created. If the file already exists, logs will be " +
					"appended to the file. This can also be set through the enviornment " +
					"variable `FOREMAN_PROVIDER_LOGFILE`. Defaults to " +
					"`'terraform-provider-foreman.log'`.",
			},

			// -- API Server configuration --

			"server_hostname": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Description: "The hostname / IP address of the Foreman REST API server",
			},
			"server_protocol": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "https",
				Description: "The protocol the Foreman REST API server is using for " +
					"communication. Defaults to `\"https\"`.",
			},

			// -- REST client configuration --

			"client_tls_insecure": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				Description: "Whether or not to verify the server's certificate. " +
					"Defaults to `false`.",
			},

			// -- client credentials --

			"client_username": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					ClientUsernameEnv,
					"",
				),
				Description: "The username to authenticate against Foreman. This can " +
					"also be set through the environment variable `FOREMAN_CLIENT_USERNAME`. " +
					"Defaults to `\"\"`.",
			},
			"client_password": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					ClientPasswordEnv,
					"",
				),
				Description: "The username to authenticate against Foreman. This can " +
					"also be set through the environment variable `FOREMAN_CLIENT_PASSWORD`. " +
					"Defaults to `\"\"`.",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"foreman_architecture":         resourceForemanArchitecture(),
			"foreman_host":                 resourceForemanHost(),
			"foreman_hostgroup":            resourceForemanHostgroup(),
			"foreman_media":                resourceForemanMedia(),
			"foreman_model":                resourceForemanModel(),
			"foreman_partitiontable":       resourceForemanPartitionTable(),
			"foreman_provisioningtemplate": resourceForemanProvisioningTemplate(),
			"foreman_smartproxy":           resourceForemanSmartProxy(),
			"foreman_computeresource":      resourceForemanComputeResource(),
		},

		DataSourcesMap: map[string]*schema.Resource{
			"foreman_architecture":         dataSourceForemanArchitecture(),
			"foreman_domain":               dataSourceForemanDomain(),
			"foreman_environment":          dataSourceForemanEnvironment(),
			"foreman_hostgroup":            dataSourceForemanHostgroup(),
			"foreman_media":                dataSourceForemanMedia(),
			"foreman_model":                dataSourceForemanModel(),
			"foreman_operatingsystem":      dataSourceForemanOperatingSystem(),
			"foreman_partitiontable":       dataSourceForemanPartitionTable(),
			"foreman_provisioningtemplate": dataSourceForemanProvisioningTemplate(),
			"foreman_smartproxy":           dataSourceForemanSmartProxy(),
			"foreman_subnet":               dataSourceForemanSubnet(),
			"foreman_templatekind":         dataSourceForemanTemplateKind(),
			"foreman_computeresource":      dataSourceForemanComputeResource(),
		},
		ConfigureFunc: providerConfigure,
	}

}

// providerConfigure uses the configuration values from the terraform file to
// configure the provider.  Returns an authenticated REST client for
// communication with Foreman.
func providerConfigure(d *schema.ResourceData) (interface{}, error) {

	var ok bool

	// parsing log level
	var logLevelStr string
	if logLevelStr, ok = d.Get("provider_loglevel").(string); !ok || logLevelStr == "" {
		log.Printf(
			"[INFO ] Log level not set from the configuration option "+
				"'provider_loglevel'. Got [%s]. Defaulting to [%s]",
			logLevelStr,
			DefaultProviderLogLevel,
		)
		logLevelStr = DefaultProviderLogLevel
	}
	// convert the string form of the log level into the logger.LogLevel type
	logLevel, logLvlErr := logger.LogLevelFromString(logLevelStr)
	if logLvlErr != nil {
		log.Printf(
			"[WARN ] Invalid log level value found for provider configuration: [%s]. ",
			logLevelStr,
		)
	}

	// parsing log file
	var logFile string
	if logFile, ok = d.Get("provider_logfile").(string); !ok || logFile == "" {
		log.Printf(
			"[INFO ] Log file not set from the configuration option "+
				"'provider_logfile'. Got [%s]. Defaulting to [%s]",
			logFile,
			DefaultProviderLogFile,
		)
		logFile = DefaultProviderLogFile
	}

	// Construct the logging configuration and initialize the logging
	logConfig := LoggingConfig{
		LogLevel: logLevel,
		LogFile:  logFile,
	}
	log.Printf(
		"[DEBUG] LoggingConfig: [%+v]",
		logConfig,
	)
	InitLogger(logConfig)
	log.Printf(
		"[INFO ] Provider logger properly initialized. The log level is "+
			"set to [%s].",
		logConfig.LogLevel.String(),
	)

	config := Config{
		// -- server configuration --
		Server: api.Server{
			URL: url.URL{
				Scheme: d.Get("server_protocol").(string),
				Host:   d.Get("server_hostname").(string),
			},
		},
		// -- client configuration --
		ClientTLSInsecure: d.Get("client_tls_insecure").(bool),
		ClientCredentials: api.ClientCredentials{
			Username: d.Get("client_username").(string),
			Password: d.Get("client_password").(string),
		},
	}

	return config.Client()
}

// InitLogger initialize the provider's shared logging instance. The shared
// logger will attempt to log to a file.  If an error is encountered while
// trying to set up the log file , the error is captured with Golang stdlib
// "log" and the default log writer is used.
func InitLogger(logConfig LoggingConfig) {
	// Set the log level. If the log level is set to 'NONE', then return
	// and do not continue with file logging
	logger.SetLevel(logConfig.LogLevel)
	if logConfig.LogLevel == logger.LevelNone {
		return
	}
	// If the log file is set to stdlog, return. The logger package uses
	// stdlog by default
	if logConfig.LogFile == LogFileStdLog {
		return
	}
	// attempt to open the file for writing.  If the file doesn't already
	// exist, feel free to create it for us.  If the file already exists,
	// open it in append mode.  If an error is encountered, fall back to the
	// default writer.
	fileFlags := os.O_WRONLY | os.O_CREATE | os.O_APPEND
	file, fileErr := os.OpenFile(logConfig.LogFile, fileFlags, 0775)
	if fileErr != nil {
		log.Printf(
			"[ERROR] Could not initialize provider's file log file [%s]. "+
				"Error: [%s].",
			logConfig.LogFile,
			fileErr.Error(),
		)
		log.Printf(
			"[INFO] Sending provider's log output to default io.Writer",
		)
		return
	}
	// No file errors - set the standard logger to write to the file.
	logger.SetOutput(file)
	log.Printf(
		"[INFO ] Provider logger set to write to [%s]",
		logConfig.LogFile,
	)
}
