package admin

import (
	"net/http"

	"github.com/daodao97/xgo/xadmin"
	"github.com/daodao97/xgo/xdb"
)

func RegHook() {
	xadmin.RegAfterList("user", func(r *http.Request, list []xdb.Row) []xdb.Row {
		return list
	})
}
