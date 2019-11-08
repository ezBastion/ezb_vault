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
	conf "github.com/ezbastion/ezb_vault/configuration"
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

func Setup(isIntSess bool, quiet bool) error {

	logmanager.Debug("Entering in setup process")
	CheckFolder(isIntSess)

	if quiet == false {
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
			p := ez_stdio.AskForValue("ezb_pki", conf.Conf.EzbPki, `^[a-zA-Z0-9-\.]+:[0-9]{4,5}$`)
			c := ez_stdio.AskForConfirmation(fmt.Sprintf("pki address (%s) ok?", p))
			if c {
				conn, err := net.Dial("tcp", p)
				if err != nil {
					logmanager.Error(fmt.Sprintf("## Failed to connect to %s ##\n", p))
				} else {
					conn.Close()
					conf.Conf.EzbPki = p
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
			tmp := conf.Conf.SAN

			san := ez_stdio.AskForValue("SAN (comma separated list)", strings.Join(conf.Conf.SAN, ","), `(?m)^[[:ascii:]]*,?$`)

			t := strings.Replace(san, " ", "", -1)
			tmp = strings.Split(t, ",")
			c := ez_stdio.AskForConfirmation(fmt.Sprintf("SAN list %s ok?", tmp))
			if c {
				conf.Conf.SAN = tmp
				logmanager.Debug(fmt.Sprintf("EZB_SAN set to %s", tmp))
				break
			}
		}

		_, fica := os.Stat(path.Join(exPath, conf.Conf.CaCert))
		logmanager.Debug(fmt.Sprintf("Cacert sets to %s", fica))
		_, fipriv := os.Stat(path.Join(exPath, conf.Conf.PrivateKey))
		logmanager.Debug(fmt.Sprintf("Privatekey sets to %s", fipriv))
		_, fipub := os.Stat(path.Join(exPath, conf.Conf.PublicCert))
		logmanager.Debug(fmt.Sprintf("PublicCert sets to %s", fipub))
	
		if os.IsNotExist(fica) || os.IsNotExist(fipriv) || os.IsNotExist(fipub) {
			logmanager.Debug("Setting the certificate")
			keyFile := path.Join(exPath, conf.Conf.PrivateKey)
			certFile := path.Join(exPath, conf.Conf.PublicCert)
			caFile := path.Join(exPath, conf.Conf.CaCert)
			request := certmanager.NewCertificateRequest(conf.Conf.ServiceName, 730, conf.Conf.SAN)
			certmanager.Generate(request, conf.Conf.EzbPki, certFile, keyFile, caFile)
			logmanager.Debug("Certificate generated")
		}
	
		// we have to handle the sta certificate
		stacert := ""
		if conf.Conf.StaCert != "" {
			stacert = conf.Conf.StaCert
		} else {
			stacert = "ezb_sta.crt"
			conf.Conf.StaCert = "ezb_sta.crt"
		}
	
		staca := "" 
		if conf.Conf.StaPath == "" {
			staca = path.Join(exPath, stacert)
			conf.Conf.StaPath = exPath
		} else {
			staca = path.Join(conf.Conf.StaPath, conf.Conf.StaCert)
		}
		logmanager.Info("********************************", true)
		logmanager.Info("*** STA public cert settings ***", true)
		logmanager.Info("********************************", true)
		logmanager.Debug(fmt.Sprintf("sta public key path set to %s",conf.Conf.StaPath))
		logmanager.Debug(fmt.Sprintf("sta public key file set to %s",conf.Conf.StaCert))
		_, stapub := os.Stat(staca)
		tpath := conf.Conf.StaPath
		tcert := conf.Conf.StaCert
		for {
			if os.IsNotExist(stapub) {
				if ez_stdio.AskForConfirmation("STA public cert is not found, do you want to set an alternate path and file ?"){
					tpath = ez_stdio.AskForStringValue("ezb_sta public cert path :")
					tcert = ez_stdio.AskForStringValue("ezb_sta cert file name :")
					staca = path.Join(tpath, tcert)
					_, stapub = os.Stat(staca)
				} else {
					displaySTAError(stacert,conf.Conf.StaPath)
					break
				}
			} else {
				conf.Conf.StaPath = tpath
				conf.Conf.StaCert = tcert
				logmanager.Info(fmt.Sprintf("STA certificate found and set in configuration file %s in folder %s",tpath,tcert),true)
				break
			}
		}
	
		c, _ := json.Marshal(conf.Conf)
		ioutil.WriteFile(conf.ConfFile, c, 0600)
		logmanager.Info(fmt.Sprintf("%s saved",conf.ConfFile))
	}

	return nil
}

func displaySTAError(tcert, tpath string) {
	logmanager.Warning(fmt.Sprintf("STA certificate %s not found at %s",tcert, tpath))
	fmt.Printf("\nSTA public certificate is not found, please change configuration file after STA cert is copied")
	fmt.Printf("\nStaCert => name of the certificate")
	fmt.Printf("\nSTAPath => path of the certificate")
	fmt.Printf("\n")
}