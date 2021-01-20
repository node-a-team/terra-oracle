package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
)

const (
	ConfigPath = "config"
	VoteMode   = "vote-mode"
)

var (
	Config configType
)

type configType struct {
	Title string `json:"title"`

	Validator struct {
		OperatorAddr string `json:"operatorAddr"`
	}
	Feeder struct {
		Password string `json:"password"`
	}
	APIs struct {
		Luna struct {
			Krw struct {
				Coinone string `json:"coinone"`
			}
		}

		Stables struct {
			Currencylayer string `json:"currencylayer"`
		}

		Sdr struct {
			Imf string `json:"imf"`
		}

		Band struct {
			Active bool `json:"active"`
			Band string `json:"band"`
		}
	}
	Options struct {
		Interval time.Duration `json:"interval"`
	}
}

func Init() {

	Config = readConfig()
}

func readConfig() configType {

	var config configType

	path := viper.GetString(ConfigPath) + "/config.toml"

	if _, err := toml.DecodeFile(path, &config); err != nil {

		log.Fatal("Config file is missing: ", config)
	}

	return config

}
