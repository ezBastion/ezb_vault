// This file is part of ezBastion.

//     ezBastion is free software: you can redistribute it and/or modify
//     it under the terms of the GNU Affero General Public License as published by
//     the Free Software Foundation, either version 3 of the License, or
//     (at your option) any later version.

//     ezBastion is distributed in the hope that it will be useful,
//     but WITHOUT ANY WARRANTY; without even the implied warranty of
//     MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//     GNU Affero General Public License for more details.

//     You should have received a copy of the GNU Affero General Public License
//     along with ezBastion.  If not, see <https://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ezbastion/ezb_lib/logmanager"
	"github.com/ezbastion/ezb_lib/servicemanager"
	"github.com/ezbastion/ezb_vault/configuration"
	"github.com/ezbastion/ezb_vault/setup"

	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"

	fqdn "github.com/ShowMax/go-fqdn"
)

var logPath string
var exPath string
var conf configuration.Configuration
var firstcall bool
var isIntSess bool

func init() {

	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
	// Loading the conf if exists
	_fqdn := fqdn.Get()
	hostname, _ := os.Hostname()
	setup.CheckFolder()

	var err error
	conf, err = configuration.CheckConfig(true, exPath)

	if err != nil {
		firstcall = true
		conf.Listen = "localhost:5100"
		conf.ServiceFullName = "Easy Bastion Vault"
		conf.ServiceName = "ezb_vault"
		conf.LogLevel = "debug"
		conf.LogPath = ""
		conf.CaCert = "cert/ca.crt"
		conf.PrivateKey = "cert/ezb_vault.key"
		conf.PublicCert = "cert/ezb_vault.crt"
		conf.DB = "db/ezb_vault.db"
		conf.EzbPki = "localhost:6000"
		conf.StaPath = ""
		conf.JsonToStdout = false
		conf.ReportCaller = false
		conf.SAN = []string{_fqdn, hostname}
	}

	defaultconflisten = conf.Listen
	logmanager.StartWindowsEvent("ezb_vault")

	exe, _ := os.Executable()
	// logpath is not the same with a debug (exe folder) or service (%windor%\system32)
	if conf.LogPath == "" {
		logPath = filepath.Dir(exe) + string(os.PathSeparator) + "log"
	} else {
		logPath = conf.LogPath
	}
	isIntSess, _ = svc.IsAnInteractiveSession()
	logmanager.SetLogLevel(conf.LogLevel, logPath, "ezb_vault.log", 1024, 5, 10, isIntSess, conf.ReportCaller, conf.JsonToStdout)

	if !isIntSess {
		// if not in session, set a default log folder
		logmanager.Info("EZB_VAULT started by system command")
	}

	// c, _ := json.Marshal(conf)
	// ConfFile := path.Join(exPath, "conf/config.json")
	// if err := ioutil.WriteFile(ConfFile, c, 0600); err != nil {
	// 	logmanager.Fatal(err.Error())
	// }
	// logmanager.Info(fmt.Sprintf("%s saved", ConfFile), conf.JsonToStdout)
}

func main() {

	logmanager.Debug("EZB_VAULT, entering in main process")

	if !isIntSess && !firstcall {
		// if not in session, it is a start request
		logmanager.Debug(fmt.Sprintf("Service %s request to start ...", conf.ServiceName))
		RunService(conf.ServiceName, false)
		return
	}

	// from here, we are in session, handle the commands
	app := cli.NewApp()
	app.Name = "ezb_vault"
	app.Version = "0.1.0-rc1"
	app.Usage = "Manage ezBastion key/value vault storage."

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Genarate config file.",
			Action: func(c *cli.Context) error {
				err := setup.Setup(true, firstcall)
				return err
			},
		}, {
			Name:  "debug",
			Usage: "Start ezb_vault in console.",
			Action: func(c *cli.Context) error {
				logmanager.Debug("cli command debug started")
				configuration.CheckConfig(true, exPath)
				RunService(conf.ServiceName, true)
				return nil
			},
		}, {
			Name:  "install",
			Usage: "Add ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				logmanager.Debug("cli command install started")
				configuration.CheckConfig(true, exPath)
				return servicemanager.InstallService(conf.ServiceName, conf.ServiceFullName)
			},
		}, {
			Name:  "remove",
			Usage: "Remove ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				logmanager.Debug("cli command remove started")
				configuration.CheckConfig(true, exPath)
				return servicemanager.RemoveService(conf.ServiceName)
			},
		}, {
			Name:  "start",
			Usage: "Start ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				logmanager.Debug("cli command start started")
				configuration.CheckConfig(true, exPath)
				return servicemanager.StartService(conf.ServiceName)
			},
		}, {
			Name:  "stop",
			Usage: "Stop ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				logmanager.Debug("cli command stop started")
				configuration.CheckConfig(true, exPath)
				return servicemanager.ControlService(conf.ServiceName, svc.Stop, svc.Stopped)
			},
		},
	}

	cli.AppHelpTemplate = fmt.Sprintf(`
		
		███████╗███████╗██████╗  █████╗ ███████╗████████╗██╗ ██████╗ ███╗   ██╗
		██╔════╝╚══███╔╝██╔══██╗██╔══██╗██╔════╝╚══██╔══╝██║██╔═══██╗████╗  ██║
		█████╗    ███╔╝ ██████╔╝███████║███████╗   ██║   ██║██║   ██║██╔██╗ ██║
		██╔══╝   ███╔╝  ██╔══██╗██╔══██║╚════██║   ██║   ██║██║   ██║██║╚██╗██║
		███████╗███████╗██████╔╝██║  ██║███████║   ██║   ██║╚██████╔╝██║ ╚████║
		╚══════╝╚══════╝╚═════╝ ╚═╝  ╚═╝╚══════╝   ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
																			   
						██╗   ██╗ █████╗ ██╗   ██╗██╗  ████████╗               
						██║   ██║██╔══██╗██║   ██║██║  ╚══██╔══╝               
						██║   ██║███████║██║   ██║██║     ██║                  
						╚██╗ ██╔╝██╔══██║██║   ██║██║     ██║                  
						 ╚████╔╝ ██║  ██║╚██████╔╝███████╗██║                  
						  ╚═══╝  ╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝                  
																			  
%s
INFO:
		http://www.ezbastion.com		
		support@ezbastion.com
		`, cli.AppHelpTemplate)
	app.Run(os.Args)
}
