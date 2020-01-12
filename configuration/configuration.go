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

package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path"
)

type Configuration struct {
	Listen          string   `json:"listen"`
	PrivateKey      string   `json:"privatekey"`
	PublicCert      string   `json:"publiccert"`
	CaCert          string   `json:"cacert"`
	StaPath         string   `json:"stapath"`
	DB              string   `json:"dbpath"`
	ServiceName     string   `json:"servicename"`
	ServiceFullName string   `json:"servicefullname"`
	LogLevel        string   `json:"loglevel"`
	LogPath         string   `json:"logpath"`
	EzbPki          string   `json:"ezb_pki"`
	ReportCaller    bool     `json:"reportcaller"`
	JsonToStdout    bool     `json:"jsonstdout"`
	SAN             []string `json:"san"`
}

func CheckConfig(isIntSess bool, exPath string) (conf Configuration, err error) {
	ConfFile := path.Join(exPath, "conf/config.json")
	raw, err := ioutil.ReadFile(ConfFile)
	if err != nil {
		fmt.Println("### error reading conf file " + ConfFile)
		return conf, err
	}
	err = json.Unmarshal(raw, &conf)
	if err != nil {
		fmt.Println("### error Unmarshal json file " + ConfFile)
		return conf, err
	}
	return conf, nil
}
