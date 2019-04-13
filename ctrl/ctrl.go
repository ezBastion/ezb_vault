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


package ctrl

import (
	"ezb_vault/models"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func Getdbconn(c *gin.Context) (db *gorm.DB, ret string) {
	db, _ = c.MustGet("db").(*gorm.DB)
	if db == nil {
		dberrmsg, ok := c.MustGet("dberr").(error)
		if ok {
			ret = dberrmsg.Error()
		} else {
			ret = string("unknow database connection error")
		}
		return nil, ret
	}
	return db, ""
}

func GetAll(c *gin.Context) {
	key := c.GetHeader("EZB-VAULT-KEY")
	var Raw []models.KeyVal
	var out []models.KeyVal
	db, err := Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	user, _ := c.MustGet("sub").(string)
	if err := db.Where("u = ? ", user).Find(&Raw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	for _, r := range Raw {
		o := r.Decrypt(key)
		if o.V != "" {
			out = append(out, o)
		}
	}
	fmt.Println("out: ", len(out))
	if len(out) == 0 {
		c.JSON(http.StatusNoContent, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

func GetVal(c *gin.Context) {
	key := c.GetHeader("EZB-VAULT-KEY")
	var Raw models.KeyVal
	db, err := Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	name := c.Param("name")
	user, _ := c.MustGet("sub").(string)
	if err := db.Where("u = ? AND k = ?", user, name).Find(&Raw).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNoContent, err.Error())
			return
		} else {
			c.JSON(http.StatusInternalServerError, err.Error())
			return
		}
	}
	out := Raw.Decrypt(key)
	if out.V == "" {
		c.JSON(http.StatusNoContent, out)
		return
	}
	c.JSON(http.StatusOK, out)
}

func AddVal(c *gin.Context) {
	key := c.GetHeader("EZB-VAULT-KEY")
	var Raw models.KeyVal
	db, err := Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if err := c.BindJSON(&Raw); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := c.MustGet("sub").(string)
	Raw.U = user
	newRaw := Raw.Encrypt(key)
	db.NewRecord(Raw)
	if err := db.Create(&newRaw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusCreated, Raw)
}
func UpdateVal(c *gin.Context) {
	key := c.GetHeader("EZB-VAULT-KEY")
	var NewRaw models.KeyVal
	var OldRaw models.KeyVal
	db, err := Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if err := c.BindJSON(&NewRaw); err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	user, _ := c.MustGet("sub").(string)
	name := c.Param("name")
	if err := db.Where("u = ? AND k = ?", user, name).Find(&OldRaw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	if NewRaw.K != "" {
		OldRaw.K = NewRaw.K
	}
	if NewRaw.V != "" {
		OldRaw.V = NewRaw.Encrypt(key).V
	}
	if err := db.Save(&OldRaw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, NewRaw)
}
func DeleteVal(c *gin.Context) {
	var Raw models.KeyVal
	db, err := Getdbconn(c)
	if err != "" {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	name := c.Param("name")
	user, _ := c.MustGet("sub").(string)
	if err := db.Where("u = ? AND k = ?", user, name).Delete(&Raw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusNoContent, Raw)
}
