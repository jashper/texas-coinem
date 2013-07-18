package Server

import (
	"fmt"
	db "github.com/christopherhesse/rethinkgo"
)

type DatabaseRequest struct {
	request db.Exp
	result  chan *db.Rows
}

type Database struct {
	session *db.Session
	queue   chan DatabaseRequest
}

func (this *Database) Connect(address, database string, bufCount int) (err error) {
	this.session, err = db.Connect(address, database)
	if err != nil {
		fmt.Println("CRITICAL : Failed to connect to " + database)
		return
	}

	this.queue = make(chan DatabaseRequest, bufCount)

	go this.run()

	return
}

func (this *Database) MakeRequest(request db.Exp) chan *db.Rows {
	result := make(chan *db.Rows, 1)
	dbRequest := DatabaseRequest{request, result}
	this.queue <- dbRequest
	return result
}

func (this *Database) run() {
	for {
		dbRequest := <-this.queue
		dbRequest.result <- dbRequest.request.Run(this.session)
	}
}
