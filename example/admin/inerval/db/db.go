package db

import (
	_ "embed"

	"github.com/daodao97/xgo/xdb"
	_ "github.com/go-sql-driver/mysql"
)

var UserModel xdb.Model
var ConvModel xdb.Model

func InitDB(dbConf *xdb.Config) error {
	err := xdb.Init(map[string]*xdb.Config{
		"default": dbConf,
	})

	if err != nil {
		return err
	}

	UserModel = xdb.New("user",
		xdb.WithFakeDelKey("is_deleted"),
	)

	return nil
}
