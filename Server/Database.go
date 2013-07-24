package Server

import (
	"fmt"
	r "github.com/christopherhesse/rethinkgo"
)

/*
	This class connects to a RethinkDB database and
	provides a method for outside goprocesses to
	make asynch read and write requests to the database

	Side-Note:  Safe to call from simultaneous goprocesses
*/

// ############################################
//     Helper Structs
// ############################################

type DatabaseRequest struct {
	request r.Exp
	result  chan *r.Rows
}

// ############################################
//     Constructor Struct & Init
// ############################################

type Database struct {
	session *r.Session
	queue   chan DatabaseRequest
}

func (this *Database) Init(address, database string, bufCount int) (err error) {
	this.session, err = r.Connect(address, database)
	if err != nil {
		fmt.Println("CRITICAL : Failed to connect to DB " + database)
		return
	}

	this.queue = make(chan DatabaseRequest, bufCount)

	go this.run()

	return
}

func (this *Database) run() {
	for {
		dbRequest := <-this.queue
		dbRequest.result <- dbRequest.request.Run(this.session)
	}
}

func (this *Database) MakeRequest(request r.Exp) *r.Rows {
	result := make(chan *r.Rows, 1)
	dbRequest := DatabaseRequest{request, result}
	this.queue <- dbRequest
	return <-result
}
