package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	flags "github.com/cosmos/cosmos-sdk/client/flags"

	cfg "github.com/node-a-team/terra-oracle/config"
)

func registerSystemdCmd(config cfg.ConfigType) *cobra.Command {
        registerSystemdCmd := &cobra.Command{
                Use:   "register-systemd",
                Short: "Register systemd service(oracle.service)",
                Run: func(cmd *cobra.Command, args []string)  {

			// Read in configuration file for local config.toml
                        cfg.Init()

			GOPATH := os.Getenv("GOPATH")
			HOME := os.Getenv("HOME")
			USER := os.Getenv("USER")
			FeederName := cfg.Config.Feeder.Name
			FeederPasswd := cfg.Config.Feeder.Password


			fmt.Println("config: ", cfg.Config)
			fmt.Println("cfg.Feeder.Password: ", FeederPasswd)
			fmt.Println("FeederName: ", FeederName)


			// Create oracle_starter
			runCmd := `
echo '#!/usr/bin/expect -f

set timeout -1

` +"spawn " +GOPATH +"/bin/terra-oracle service --from=J --broadcast-mode=block --config=" +HOME +" --vote-mode aggregate" +`

expect {
  "Enter keyring passphrase:" {
    send "` +FeederPasswd +`\r"; exp_continue
  }
}' > $GOPATH/bin/terra-oracle_starter

chmod +x ` +GOPATH +`/bin/terra-oracle_starter
`
			shellCmd(runCmd)

			// Check oracle_starter
			fmt.Printf("Create "+GOPATH +"/terra-oracle_starter\n")
			runCmd = `
cat ` +GOPATH +`/bin/terra-oracle_starter
`
			shellCmd(runCmd)



			// Create terra-oracle.service
			runCmd = `
sudo tee /etc/systemd/system/terra-oracle.service > /dev/null <<EOF 
[Unit]
Description=Terra Oracle
After=network-online.target

[Service]
User=` +USER +`
ExecStart=` +GOPATH +`/bin/terra-oracle_starter
StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=terra-oracle
Restart=always
RestartSec=3

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable terra-oracle.service
`
                        shellCmd(runCmd)

			// Check terra-oracle.service
                        fmt.Printf("Create /etc/systemd/system/terra-oracle.service\n")
                        runCmd = `
cat /etc/systemd/system/terra-oracle.service
`
			shellCmd(runCmd)

                },
        }

	registerSystemdCmd.Flags().String(cfg.ConfigPath, "", "Directory for config.toml")
        registerSystemdCmd.MarkFlagRequired(cfg.ConfigPath)

	registerSystemdCmd = flags.PostCommands(registerSystemdCmd)[0]
        registerSystemdCmd.MarkFlagRequired(flags.FlagFrom)

        return registerSystemdCmd
}


func shellCmd(cmd string) {
        out, err := exec.Command("/bin/bash", "-c", cmd).Output()
        if err != nil {
                fmt.Printf("SHELL cmd error : ")
                fmt.Printf(" %s\n", cmd)
                fmt.Println(" out : ", string(out[:]))
        } else {
//                fmt.Printf("..\nSHELL out : ")
                fmt.Printf(string(out[:]) + "\n")
        }
}
