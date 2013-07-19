package Server

import (
	"fmt"
	r "github.com/christopherhesse/rethinkgo"
)

type DatabaseRequest struct {
	request r.Exp
	result  chan *r.Rows
}

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

func (this *Database) MakeRequest(request r.Exp) *r.Rows {
	result := make(chan *r.Rows, 1)
	dbRequest := DatabaseRequest{request, result}
	this.queue <- dbRequest
	return <-result
}

func (this *Database) run() {
	for {
		dbRequest := <-this.queue
		dbRequest.result <- dbRequest.request.Run(this.session)
	}
}
