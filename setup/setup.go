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
	ex, _ := os.Executable()
	exPath = filepath.Dir(ex)
	ConfFile := path.Join(exPath, "conf/config.json")
	conf, err := configuration.CheckConfig(true, exPath)

	CheckFolder(isIntSess)
	// If StaCert is not set, we have a default conf file
	if (conf.StaCert == "" || err != nil ) {
		logmanager.Info("**************************", true)
		logmanager.Info("** EZB VAULT SETUP MODE **", true)
		logmanager.Info("**************************", true)
		logmanager.Info("Entering in the setup mode. Please answer the following requests ", true)
		logmanager.Info("\n",true)
		logmanager.Info("********************", true)
		logmanager.Info("*** PKI settings ***", true)
		logmanager.Info("********************", true)
		logmanager.Info("ezBastion nodes use elliptic curve digital signature algorithm ", true)
		logmanager.Info("(ECDSA) to communicate.",true)
		logmanager.Info("We need ezb_pki address and port, to request certificat pair.",true)
		logmanager.Info("ex: 10.20.1.2:6000 pki.domain.local:6000",true)

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
		_fqdn := fqdn.Get()
		hostname, _ := os.Hostname()
		logmanager.Info("********************", true)
		logmanager.Info("*** SAN settings ***", true)
		logmanager.Info("********************", true)
		logmanager.Info("Certificat Subject Alternative Name.",true)
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
	
		// We set the sta path by mandatory to cert
		conf.StaPath = path.Join(exPath, "cert")
		logmanager.Info("********************************", true)
		logmanager.Info("*** STA public cert settings ***", true)
		logmanager.Info("********************************", true)
		logmanager.Debug(fmt.Sprintf("sta public key path set mandatory to %s",conf.StaPath))
		_, stapub := os.Stat(staca)
		tpath := conf.StaPath
		tcert := "ezb_sta.crt"
		staca := path.Join(tpath, tcert)
		conf.StaPath = tpath
		_, stapub = os.Stat(staca)
		for {
			if os.IsNotExist(stapub) {
				if ez_stdio.AskForConfirmation("STA public cert is not found, do you want to set an alternate path ?"){
					tpath = ez_stdio.AskForStringValue("ezb_sta public cert path :")
					tcert = "ezb_sta.crt"
					staca = path.Join(tpath, tcert)
					_, stapub = os.Stat(staca)
				} else {
					logmanager.Warning(fmt.Sprintf("STA certificate not found in folder %s, please fix it after setup process", tpath))
					fmt.Printf("\nSTA public certificate is not found, please copy STA cert in %s path", tpath)
					conf.StaPath = tpath
					break
				}
			} else {
				conf.StaPath = tpath
				logmanager.Info(fmt.Sprintf("STA certificate found and set in folder %s",tpath),true)
				break
			}
		}
	
		c, _ := json.Marshal(conf)
		ioutil.WriteFile(ConfFile, c, 0600)
		logmanager.Info(fmt.Sprintf("%s saved",ConfFile))
	}

	return nil
}
