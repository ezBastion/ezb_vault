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

package Middleware

import (
	"crypto/ecdsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/ezbastion/ezb_vault/configuration"

	"net/http"
	"os"
	"path"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"

	"github.com/ezbastion/ezb_lib/logmanager"
)

type Payload struct {
	JTI string `json:"jti"`
	ISS string `json:"iss"`
	SUB string `json:"sub"`
	AUD string `json:"aud"`
	EXP int    `json:"exp"`
	IAT int    `json:"iat"`
}

func AuthJWT(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		logmanager.WithFields("Middleware","jwt")
		var err error
		authHead := c.GetHeader("Authorization")
		bearer := strings.Split(authHead, " ")
		if len(bearer) != 2 {
			logmanager.Error(fmt.Sprintf("bad Authorization #V0001: authHead:'%s'",authHead))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0001"))
			return
		}
		if strings.Compare(strings.ToLower(bearer[0]), "bearer") != 0 {
			logmanager.Error(fmt.Sprintf("bad Authorization #V0002: %s",authHead))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0002"))
			return
		}
		tokenString := bearer[1]
		parts := strings.Split(tokenString, ".")
		if len(parts) != 3 {
			logmanager.Error("Bad bearer format.")
			c.AbortWithError(http.StatusForbidden, errors.New("#V0012"))
			return
		}
		p, err := base64.RawStdEncoding.DecodeString(parts[1])
		if err != nil {
			logmanager.Error(fmt.Sprintf("Unable to decode payload: ", err.Error()))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0009"))
			return
		}
		var payload Payload
		err = json.Unmarshal(p, &payload)
		if err != nil {
			logmanager.Error(fmt.Sprintf("Unable to parse payload: ", err.Error()))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0011"))
			return
		}
		jwtkeyfile := fmt.Sprintf("%s.crt", payload.ISS)
		// change the path to the configuration file
		// TODO check for cert name itself, in the payload
		jwtpubkey := path.Join(configuration.Conf.StaPath, jwtkeyfile)
		logmanager.Debug(fmt.Sprintf("sta public certificate set to %s", jwtpubkey))

		if _, err := os.Stat(jwtpubkey); os.IsNotExist(err) {
			logmanager.Error(fmt.Sprintf("Unable to load sta public certificate: ", err.Error()))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0010"))
			return
		}

		key, _ := ioutil.ReadFile(jwtpubkey)
		var ecdsaKey *ecdsa.PublicKey
		if ecdsaKey, err = jwt.ParseECPublicKeyFromPEM(key); err != nil {
			logmanager.Error(fmt.Sprintf("Unable to parse ECDSA public key: ", err.Error()))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0003"))
		}
		methode := jwt.GetSigningMethod("ES256")
		// parts := strings.Split(tokenString, ".")
		err = methode.Verify(strings.Join(parts[0:2], "."), parts[2], ecdsaKey)
		if err != nil {
			logmanager.Error(fmt.Sprintf("Error while verifying key: %s", err.Error()))
			c.AbortWithError(http.StatusForbidden, errors.New("#V0004"))
		}
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return jwt.ParseECPublicKeyFromPEM(key)
		})
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// log.Println(claims["iss"], claims["sub"])
			c.Set("sub", claims["sub"])
		} else {
			c.AbortWithError(http.StatusForbidden, errors.New("#V0005"))
			logmanager.Error(err.Error())
			return
		}
		c.Next()
	}
}
