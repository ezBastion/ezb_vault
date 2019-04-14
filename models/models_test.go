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

package models

import (
	"testing"
)

func TestAES(t *testing.T) {
	Raws := []KeyVal{
		{ID: 0, U: "user0", K: "key0", V: "098f6bcd4621d373cade4e832627b4f6"},
		{ID: 1, U: "user1", K: "key1", V: " "},
		{ID: 2, U: "user2", K: "key2", V: "@</;%^?_-☻.♥"},
	}
	for _, Raw := range Raws {
		r := Raw.Encrypt("d4621d373cad")
		o := r.Decrypt("d4621d373cad")
		if o.V != Raw.V {
			t.Errorf("TestAES was incorrect, got: <%s>, want: <%s>.", o.V, Raw.V)
		}
	}
}
