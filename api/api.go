package api

import "gopkg.in/dedis/onet.v1"

// ID is used for registration on the onet.
const ID = "nevv"

// Client structure for communication with the CoSi service.
type Client struct {
	*onet.Client
}

// NewClient instantiates a new cosi.Client.
func NewClient() *Client {
	return &Client{Client: onet.NewClient(ID)}
}
