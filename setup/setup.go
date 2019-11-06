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

package setup

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"path/filepath"
	"github.com/ezbastion/ezb_lib/certmanager"
	"github.com/ezbastion/ezb_lib/logmanager"
	"github.com/ezbastion/ezb_vault/configuration"
	"github.com/ezbastion/ezb_lib/ez_stdio"
	"strings"

	fqdn "github.com/ShowMax/go-fqdn"
)

var exPath string

func init() {
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
}

func CheckFolder(isIntSess bool) {

	if _, err := os.Stat(path.Join(exPath, "cert")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "cert"), 0600)
		if err != nil {
			return
		}
		logmanager.Info("Make cert folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "log")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "log"), 0600)
		if err != nil {
			return
		}
		logmanager.Info("Make log folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "conf")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "conf"), 0600)
		if err != nil {
			return
		}
		logmanager.Info("Make conf folder.")
	}
	if _, err := os.Stat(path.Join(exPath, "db")); os.IsNotExist(err) {
		err = os.MkdirAll(path.Join(exPath, "db"), 0600)
		if err != nil {
			return
		}
		logmanager.Info("Make db folder.")
	}
}

func Setup(isIntSess bool) error {

	logmanager.Debug("Entering in setup process")
	_fqdn := fqdn.Get()
	quiet := true
	confFile := path.Join(exPath, "conf/config.json")
	hostname, _ := os.Hostname()
	CheckFolder(isIntSess)
	conf, err := configuration.CheckConfig(isIntSess, exPath)
	if err != nil {
		quiet = false
		conf.Listen = ":5100"
		conf.ServiceFullName = "Easy Bastion Vault"
		conf.ServiceName = "ezb_vault"
		conf.LogLevel = "warning"
		conf.CaCert = "cert/ca.crt"
		conf.PrivateKey = "cert/ezb_vault.key"
		conf.PublicCert = "cert/ezb_vault.crt"
		conf.DB = "db/ezb_vault.db"
		conf.EzbPki = "localhost:6000"
		conf.SAN = []string{_fqdn, hostname}
	}

	_, fica := os.Stat(path.Join(exPath, conf.CaCert))
	logmanager.Debug(fmt.Sprintf("Cacert sets to %s", fica))
	_, fipriv := os.Stat(path.Join(exPath, conf.PrivateKey))
	logmanager.Debug(fmt.Sprintf("Privatekey sets to %s", fipriv))
	_, fipub := os.Stat(path.Join(exPath, conf.PublicCert))
	logmanager.Debug(fmt.Sprintf("PublicCert sets to %s", fipub))

	if os.IsNotExist(fica) || os.IsNotExist(fipriv) || os.IsNotExist(fipub) {
		logmanager.Debug("Setting the certificate")
		keyFile := path.Join(exPath, conf.PrivateKey)
		certFile := path.Join(exPath, conf.PublicCert)
		caFile := path.Join(exPath, conf.CaCert)
		request := certmanager.NewCertificateRequest(conf.ServiceName, 730, conf.SAN)
		certmanager.Generate(request, conf.EzbPki, certFile, keyFile, caFile)
		logmanager.Debug("Certificate generated")
	}

	if quiet == false {
		logmanager.Info("\n\n")
		logmanager.Info("***********")
		logmanager.Info("*** PKI ***")
		logmanager.Info("***********")
		logmanager.Info("ezBastion nodes use elliptic curve digital signature algorithm ")
		logmanager.Info("(ECDSA) to communicate.")
		logmanager.Info("We need ezb_pki address and port, to request certificat pair.")
		logmanager.Info("ex: 10.20.1.2:6000 pki.domain.local:6000")

		for {
			p := ez_stdio.AskForValue("ezb_pki", conf.EzbPki, `^[a-zA-Z0-9-\.]+:[0-9]{4,5}$`)
			c := ez_stdio.AskForConfirmation(fmt.Sprintf("pki address (%s) ok?", p))
			if c {
				conn, err := net.Dial("tcp", p)
				if err != nil {
					logmanager.Error(fmt.Sprintf("## Failed to connect to %s ##\n", p))
				} else {
					conn.Close()
					conf.EzbPki = p
					logmanager.Debug(fmt.Sprintf("EZB_PKI set to %s", p))
					break
				}
			}
		}

		logmanager.Info("Certificat Subject Alternative Name.")
		logmanager.Info(fmt.Sprintf("By default using: <%s, %s> as SAN. Add more ?", _fqdn, hostname))
		for {
			tmp := conf.SAN

			san := ez_stdio.AskForValue("SAN (comma separated list)", strings.Join(conf.SAN, ","), `(?m)^[[:ascii:]]*,?$`)

			t := strings.Replace(san, " ", "", -1)
			tmp = strings.Split(t, ",")
			c := ez_stdio.AskForConfirmation(fmt.Sprintf("SAN list %s ok?", tmp))
			if c {
				conf.SAN = tmp
				logmanager.Debug(fmt.Sprintf("EZB_SAN set to %s", tmp))
				break
			}
		}

		c, _ := json.Marshal(conf)
		ioutil.WriteFile(confFile, c, 0600)
		logmanager.Info(fmt.Sprintf("%s saved",confFile))
	}

	return nil
}
