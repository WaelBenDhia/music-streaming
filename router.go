package main

import (
	"database/sql"
	"log"

	"github.com/astaxie/beego/config"
)

//Env context for handlers
type Env struct {
	Host         string
	pdb          *sql.DB
	Configurator config.Configer
	infoLog      *log.Logger
	warningLog   *log.Logger
	errorLog     *log.Logger
}
