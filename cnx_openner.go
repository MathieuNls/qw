package gosqlwrapper

import (
	"database/sql"
	"fmt"
)

type CnxOpener interface {
	OpenCnx([]string) (*sql.DB, error)
}

type MySQLCnxOpenner struct {
}

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
