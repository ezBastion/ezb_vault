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
	"fmt"
	"path"

	m "github.com/ezbastion/ezb_vault/models"

	"github.com/jinzhu/gorm"
)

func InitDB(conf Configuration, exPath string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	db, err = gorm.Open("sqlite3", path.Join(exPath, conf.DB))

	if err != nil {
		fmt.Printf("sql.Open err: %s\n", err)
		return nil, err
	}
	db.Exec("PRAGMA foreign_keys = OFF")

	db.SingularTable(true)
	if !db.HasTable(&m.KeyVal{}) {
		db.CreateTable(&m.KeyVal{})
		db.Model(&m.KeyVal{}).AddUniqueIndex("idx_keyval_id", "id")
	}
	db.AutoMigrate(&m.KeyVal{})
	return db, nil
}
