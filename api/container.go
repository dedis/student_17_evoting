package api

import (
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
)

type Election struct {
	ID     string       `protobuf:"1,req,id"`
	Admin  string       `protobuf:"2,req,admin"`
	Start  string       `protobuf:"3,req,start"`
	End    string       `protobuf:"4,req,end"`
	Data   []byte       `protobuf:"5,req,data"`
	Roster *onet.Roster `protobuf:"6,req,roster"`

	Users []string `protobuf:"7,rep,users"`

	Key         abstract.Point `protobuf:"8,opt,key"`
	Description string         `protobuf:"9,opt,description"`
}

type Ballot struct {
	User string `protobuf:"1,req,user"`

	Alpha abstract.Point `protobuf:"2,req,alpha"`
	Beta  abstract.Point `protobuf:"3,req,beta"`

	Clear []byte `protobuf:"4,opt,clear"`
}

type Box struct {
	Ballots []*Ballot `protobuf:"1,req,ballots"`
}
