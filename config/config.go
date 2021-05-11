package config

import (
	"fmt"
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
	Config ConfigType
)

type ConfigType struct {
	Title string `json:"title"`

	Validator struct {
		OperatorAddr string `json:"operatorAddr"`
	}
	Feeder struct {
		Name string `json:"name"`
		Password string `json:"password"`
	}
	APIs struct {
		Luna struct {
			Krw struct {
				Coinone string `json:"coinone"`
				Bithumb string `json:"bithumb"`
			}
			Usd struct {
				Binance string `json:"binance"`
			}
		}

		Stables struct {
			Currencylayer string `json:"currencylayer"`
		}

		Sdr struct {
			Imf string `json:"imf"`
		}

		Band struct {
			Active bool   `json:"active"`
			Band   string `json:"band"`
		}
	}
	Options struct {
		Interval struct {
			Luna	time.Duration `json:"luna"`
			Stables	time.Duration `json:"stables"`
		}
	}
}

func Init() {

	Config = readConfig()
}

func readConfig() ConfigType {

	var config ConfigType

	fmt.Println("ConfigPath: ", ConfigPath)
	path := viper.GetString(ConfigPath) + "/config.toml"

	if _, err := toml.DecodeFile(path, &config); err != nil {

		log.Fatal("Config file is missing: ", config)
	}

	return config

}
