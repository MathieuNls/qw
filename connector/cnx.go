package qw

//Cnx defines the requiered method for each
//cnx drivers
type Cnx interface {

	//Returns a connector
	OpenCnx([]string) (interface{}, error)
}
