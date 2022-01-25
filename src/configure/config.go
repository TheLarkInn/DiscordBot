package configure

import (
	"bytes"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func checkErr(err error) {
	if err != nil {
		logrus.WithError(err).Fatal("config")
	}
}

func New() *Config {
	config := viper.New()

	// Default config
	b, _ := json.Marshal(Config{
		ConfigFile: "config.yaml",
	})
	tmp := viper.New()
	defaultConfig := bytes.NewReader(b)
	tmp.SetConfigType("json")
	checkErr(tmp.ReadConfig(defaultConfig))
	checkErr(config.MergeConfigMap(viper.AllSettings()))

	pflag.String("config", "config.yaml", "Config file location")
	pflag.Bool("noheader", false, "Disable the startup header")
	pflag.Parse()
	checkErr(config.BindPFlags(pflag.CommandLine))

	// File
	config.SetConfigFile(config.GetString("config"))
	config.AddConfigPath(".")
	err := config.ReadInConfig()
	if err != nil {
		logrus.Warning(err)
		logrus.Info("Using default config")
	} else {
		checkErr(config.MergeInConfig())
	}

	BindEnvs(config, Config{})

	// Environment
	config.AutomaticEnv()
	config.SetEnvPrefix("DISCORD_BOT")
	config.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.AllowEmptyEnv(true)

	// Print final config
	c := &Config{}
	checkErr(config.Unmarshal(&c))

	initLogging(c.Level)

	return c
}

func BindEnvs(config *viper.Viper, iface interface{}, parts ...string) {
	ifv := reflect.ValueOf(iface)
	ift := reflect.TypeOf(iface)
	for i := 0; i < ift.NumField(); i++ {
		v := ifv.Field(i)
		t := ift.Field(i)
		tv, ok := t.Tag.Lookup("mapstructure")
		if !ok {
			continue
		}
		switch v.Kind() {
		case reflect.Struct:
			BindEnvs(config, v.Interface(), append(parts, tv)...)
		default:
			_ = config.BindEnv(strings.Join(append(parts, tv), "."))
		}
	}
}

type Config struct {
	Level      string `mapstructure:"level" json:"level"`
	ConfigFile string `mapstructure:"config" json:"config"`
	NoHeader   bool   `mapstructure:"noheader" json:"noheader"`

	Discord struct {
		GuildID    string   `mapstructure:"guild_id" json:"guild_id"`
		Token      string   `mapstructure:"token" json:"token"`
		AdminRoles []string `mapstructure:"admin_roles" json:"admin_roles"`
	} `mapstructure:"discord" json:"discord"`

	Pod struct {
		Name string `mapstructure:"name" json:"name"`
	} `mapstructure:"pod" json:"pod"`

	Mongo struct {
		URI      string `mapstructure:"uri" json:"uri"`
		Database string `mapstructure:"database" json:"database"`
		Direct   bool   `mapstructure:"direct" json:"direct"`
	} `mapstructure:"mongo" json:"mongo"`

	Redis struct {
		Username   string   `mapstructure:"username" json:"username"`
		Password   string   `mapstructure:"password" json:"password"`
		MasterName string   `mapstructure:"master_name" json:"master_name"`
		Addresses  []string `mapstructure:"addresses" json:"addresses"`
		Database   int      `mapstructure:"database" json:"database"`
		Sentinel   bool     `mapstructure:"sentinel" json:"sentinel"`
	} `mapstructure:"redis" json:"redis"`

	Monitoring struct {
		Enabled bool       `mapstructure:"enabled" json:"enabled"`
		Bind    string     `mapstructure:"bind" json:"bind"`
		Labels  []KeyValue `mapstructure:"labels" json:"labels"`
	} `mapstructure:"monitoring" json:"monitoring"`

	Health struct {
		Enabled bool   `mapstructure:"enabled" json:"enabled"`
		Bind    string `mapstructure:"bind" json:"bind"`
	} `mapstructure:"health" json:"health"`

	Modules struct {
		Points struct {
			Enabled          bool   `mapstructure:"enabled" json:"enabled"`
			HourlyLimit      int    `mapstructure:"hourly_limit" json:"hourly_limit"`
			DailyLimit       int    `mapstructure:"daily_limit" json:"daily_limit"`
			WeeklyLimit      int    `mapstructure:"weekly_limit" json:"weekly_limit"`
			PointsPerMessage int    `mapstructure:"points_per_message" json:"points_per_message"`
			RequiredRoleID   string `mapstructure:"required_role_id" json:"required_role_id"`
			Roles            []struct {
				ID     string `mapstructure:"id" json:"id"`
				Points int    `mapstructure:"points" json:"points"`
			} `mapstructure:"roles" json:"roles"`
		} `mapstructure:"points" json:"points"`
		Common struct {
			Enabled         bool   `mapstructure:"enabled" json:"enabled"`
			DankRoleID      string `mapstructure:"dank_role_id" json:"dank_role_id"`
			BasedRoleID     string `mapstructure:"based_role_id" json:"based_role_id"`
			BasedRoleColors []int  `mapstructure:"based_role_colors" json:"based_role_colors"`
		} `mapstructure:"common" json:"common"`
		GoodNight struct {
			Enabled bool `mapstructure:"enabled" json:"enabled"`
		} `mapstructure:"goodnight" json:"goodnight"`
	} `mapstructure:"modules" json:"modules"`
}

type KeyValue struct {
	Key   string `mapstructure:"key" json:"key"`
	Value string `mapstructure:"value" json:"value"`
}
