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
	"ezb_vault/Middleware"
	"ezb_vault/configuration"
	"ezb_vault/routes"
	"ezb_vault/setup"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/gin-gonic/contrib/ginrus"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

func mainGin(serverchan *chan bool) {
	ex, _ := os.Executable()
	exPath := filepath.Dir(ex)
	conf, err := setup.CheckConfig(true)
	if err != nil {
		panic(err)
	}
	/* log */
	outlog := true
	gin.DisableConsoleColor()
	log.SetFormatter(&log.JSONFormatter{})
	switch conf.LogLevel {
	case "debug":
		log.SetLevel(log.DebugLevel)
		break
	case "info":
		log.SetLevel(log.InfoLevel)
		break
	case "warning":
		log.SetLevel(log.WarnLevel)
		break
	case "error":
		log.SetLevel(log.ErrorLevel)
		break
	case "critical":
		log.SetLevel(log.FatalLevel)
		break
	default:
		outlog = false
	}
	if outlog {
		if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
			err = os.MkdirAll(path.Join(exPath, "log"), 0600)
			if err != nil {
				log.Println(err)
			}
		}
		t := time.Now().UTC()
		l := fmt.Sprintf("log/ezb_vault-%d%d.log", t.Year(), t.YearDay())
		f, _ := os.OpenFile(path.Join(exPath, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		defer f.Close()
		log.SetOutput(io.MultiWriter(f))
		ti := time.NewTicker(1 * time.Minute)
		defer ti.Stop()
		go func() {
			for range ti.C {
				t := time.Now().UTC()
				l := fmt.Sprintf("log/ezb_vault-%d%d.log", t.Year(), t.YearDay())
				f, _ := os.OpenFile(path.Join(exPath, l), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				defer f.Close()
				log.SetOutput(io.MultiWriter(f))
			}
		}()
	}
	/* log */

	db, err := configuration.InitDB(conf, exPath)
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
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
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	if conf.Listen == "" {
		conf.Listen = "localhost:5100"
	}
	tlsConfig := &tls.Config{}
	server := &http.Server{
		Addr:      conf.Listen,
		TLSConfig: tlsConfig,
		Handler:   r,
	}

	go func() {
		if err := server.ListenAndServeTLS(path.Join(exPath, conf.PublicCert), path.Join(exPath, conf.PrivateKey)); err != nil {
			log.Info("listen: %s\n", err)
		}
	}()
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Info("Shutdown Server ...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Info("Server exiting")
}
