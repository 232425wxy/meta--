package node

import (
	"github.com/232425wxy/meta--/config"
	"github.com/232425wxy/meta--/database"
)

func DefaultDBProvider(name string, cfg *config.Config) database.DB {
	db, err := database.NewDB(name, cfg.BasicConfig.DBPath(), database.BackendType(cfg.BasicConfig.DBBackend))
	if err != nil {
		panic(err)
	}
	return db
}
