package config

import (
	"log"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/spf13/viper"
)

const (
	ConfigPath = "config"
)

var (
	Config	configType
)


type configType struct {

	Title				string	`json:"title"`

	Validator struct {
		OperatorAddr		string	`json:"operatorAddr"`
	}
	Feeder struct {
		Password		string	`json:"password"`
	}
	APIs struct {
		KRW struct {
			Coinone		string	`json:"coinone"`
		}

		USD struct {
			Dunamu		string	`json:"dunamu"`
		}

		MNT struct {
			Currencylayer	string	`json:"currencylayer"`
		}

		SDR struct {
			IMF		string	`json:"imf"`
		}
	}
	Options	struct {
		Interval			time.Duration	`json:"interval"`
		ChangeRateLimit struct {
			Soft			float64	`json:"soft"`
			Hard			float64	`json:"hard"`
		}
	}
}


func Init() {

	Config = readConfig()
}

func readConfig() configType {

        var config configType

	path := viper.GetString(ConfigPath)+"/config.toml"

        if _, err := toml.DecodeFile(path, &config); err != nil{

                log.Fatal("Config file is missing: ", config)
        }


	return config

}
