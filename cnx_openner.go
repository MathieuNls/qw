package gosqlwrapper

import (
	"database/sql"
	"fmt"
)

//CnxOpener defines the requiered method for each
//cnx drivers
type CnxOpener interface {
	OpenCnx([]string) (*sql.DB, error)
}

//MySQLCnxOpenner is CnxOpener for MySql
type MySQLCnxOpenner struct {
}

//OpenCnx opens a connection to a MySQL server.
//It iterates over dbCons until it can open a
//connection.
func (CnxOpener MySQLCnxOpenner) OpenCnx(dbCons []string) (*sql.DB, error) {

	var db *sql.DB
	var err error

	for index := 0; index < len(dbCons); index++ {
		//Check dns format
		db, err = sql.Open("mysql", dbCons[index])
		if err != nil {
			fmt.Println(dbCons[index] + " failed to open")
		} else {
			//Check database connectivity
			err = db.Ping()
			if err != nil {
				fmt.Println(dbCons[index] + " failed to answer ping")
			} else {
				//Database is answering, break here
				break
			}
			defer db.Close()

		}
	}
	return db, err
}
