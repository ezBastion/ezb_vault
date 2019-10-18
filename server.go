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
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
//	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"time"

	ezbevent "github.com/ezbastion/ezb_lib/eventlogmanager"
	"github.com/ezbastion/ezb_vault/Middleware"
	"github.com/ezbastion/ezb_vault/configuration"
	"github.com/ezbastion/ezb_vault/routes"
	"github.com/ezbastion/ezb_vault/setup"
	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"

)

var defaultconflisten string
var err error

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	log.Debugln("#### EXECUTE started #####")
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown
	changes <- svc.Status{State: svc.StartPending}
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	serverchan := make(chan bool)
	go MainGin(&serverchan)
loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				close(serverchan)
				break loop
			default:
				log.Errorln(fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

// RunService runs the service targeted by name. From 06/27/2019, debug is not needed as the debug is always done, log system will
// handle th level
func RunService(name string) {

	defer ezbevent.Close()
	run := debug.Run

	ezbevent.Info(fmt.Sprintf("starting the %s service", name))
	err = run(name, &myservice{})
	if err != nil {
		ezbevent.Error(fmt.Sprintf("%s service failed: %s", name, err.Error()))
		return
	}
	ezbevent.Info(fmt.Sprintf("%s service stopped", name))
}

// MainGin starts the server
func MainGin(serverchan *chan bool) {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	conf, err := setup.CheckConfig(true)
	if err != nil {
		panic(err)
	}

	// Backup logs
	log.Debugln("Backup log process")
	if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "log"), 0600)
		if err != nil {
			log.Errorln(err)
		}
	}

	ti := time.NewTicker(1 * time.Minute)
	defer ti.Stop()

	db, err := configuration.InitDB(conf, exPath)
	if err != nil {
		log.Fatalln(fmt.Sprintf("Error during InitDB Configuration : %s", err.Error()))
		panic(err)
	}

	// Init of the GIN Web HTTP framework
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(ginrus.Ginrus(log.StandardLogger(), time.RFC3339, true))
	r.Use(Middleware.AddHeaders)
	r.Use(Middleware.AuthJWT(db, conf))
	r.OPTIONS("*a", func(c *gin.Context) {
		c.AbortWithStatus(200)
	})
	r.Use(Middleware.DBMiddleware(db))
	routes.Routes(r)

	caCert, err := ioutil.ReadFile(path.Join(exPath, conf.CaCert))
	if err != nil {
		log.Fatalln(fmt.Sprintf("Error reading CaCert : %s", err.Error()))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	if conf.Listen == "" {
		log.Debugln(fmt.Sprintf("No listen port defined in conf file, settings default : %s", defaultconflisten))
		conf.Listen = defaultconflisten
	}
	tlsConfig := &tls.Config{}
	server := &http.Server{
		Addr:      conf.Listen,
		TLSConfig: tlsConfig,
		Handler:   r,
	}

	log.Infoln("Server EZB_VAULT started")
	go func() {
		if err := server.ListenAndServeTLS(path.Join(exPath, conf.PublicCert), path.Join(exPath, conf.PrivateKey)); err != nil {
			log.Infoln(fmt.Sprintf("listen: %s", err))
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Infoln("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		log.Fatalln(fmt.Sprintf("Reero during Server Shutdown : %s", err.Error()))
	}
	log.Infoln("Server exiting")
}
