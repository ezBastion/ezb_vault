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
	"github.com/ezbastion/ezb_vault/setup"

	log "github.com/sirupsen/logrus"
	ezbevent "github.com/ezbastion/ezb_lib/eventlogmanager"

	"github.com/urfave/cli"
	"golang.org/x/sys/windows/svc"
)

var logPath string

func init() {
	exe, _ := os.Executable()
	// logpath is not the same with a debug (exe folder) or service (%windor%\system32)
	logPath = filepath.Dir(exe)

	logmanager.SetLogLevel("debug", logPath, "ezb_vault.log", 1024, 5, 10)
	defaultconflisten = "localhost:5100"
	ezbevent.Open("ezb_vault")
}

func main() {

	 log.Debugln("EZB_VAULT, entering in main process")
	 isIntSess, err := svc.IsAnInteractiveSession()
	 if err != nil {
	 	log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	 }
	 if !isIntSess {
		conf, err := setup.CheckConfig(false)
		if err == nil {
			log.Debugln(fmt.Sprintf("Service %s request to start ...",conf.ServiceName))
			RunService(conf.ServiceName, false)
		}
		return
	}

	app := cli.NewApp()
	app.Name = "ezb_vault"
	app.Version = "0.1.0-rc1"
	app.Usage = "Manage ezBastion key/value vault storage."

	app.Commands = []cli.Command{
		{
			Name:  "init",
			Usage: "Genarate config file.",
			Action: func(c *cli.Context) error {
				err := setup.Setup(true)
				return err
			},
		}, {
			Name:  "debug",
			Usage: "Start ezb_vault in console.",
			Action: func(c *cli.Context) error {
				log.Debugln("cli command debug started")
				conf, _ := setup.CheckConfig(true)
				RunService(conf.ServiceName,true)
				return nil
			},
		}, {
			Name:  "install",
			Usage: "Add ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				log.Debugln("cli command install started")
				conf, _ := setup.CheckConfig(true)
				return servicemanager.InstallService(conf.ServiceName, conf.ServiceFullName)			},
		}, {
			Name:  "remove",
			Usage: "Remove ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				log.Debugln("cli command remove started")
				conf, _ := setup.CheckConfig(true)
				return servicemanager.RemoveService(conf.ServiceName)			},
		}, {
			Name:  "start",
			Usage: "Start ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				log.Debugln("cli command start started")
				conf, _ := setup.CheckConfig(true)
				return servicemanager.StartService(conf.ServiceName)			},
		}, {
			Name:  "stop",
			Usage: "Stop ezb_vault deamon windows service.",
			Action: func(c *cli.Context) error {
				log.Debugln("cli command stop started")
				conf, _ := setup.CheckConfig(true)
				return servicemanager.ControlService(conf.ServiceName, svc.Stop, svc.Stopped)			},
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
