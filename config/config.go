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
		KRW struct {
			Coinone string `json:"coinone"`
		}

		USD struct {
			Dunamu string `json:"dunamu"`
		}

		AUD struct {
			Dunamu string `json:"dunamu"`
		}

		CAD struct {
			Dunamu string `json:"dunamu"`
		}

		CHF struct {
			Dunamu string `json:"dunamu"`
		}

		CNY struct {
			Dunamu string `json:"dunamu"`
		}

		EUR struct {
			Dunamu string `json:"dunamu"`
		}

		GBP struct {
			Dunamu string `json:"dunamu"`
		}

		HKD struct {
			Dunamu string `json:"dunamu"`
		}

		INR struct {
			Dunamu string `json:"dunamu"`
		}

		JPY struct {
			Dunamu string `json:"dunamu"`
		}

		MNT struct {
			Currencylayer string `json:"currencylayer"`
		}

		SDR struct {
			IMF string `json:"imf"`
		}

		SGD struct {
			Dunamu string `json:"dunamu"`
		}

		Band struct {
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
