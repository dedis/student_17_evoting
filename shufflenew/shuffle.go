package shufflenew

import (
	"github.com/dedis/cothority/skipchain"
	"github.com/dedis/onet/network"
	"gopkg.in/dedis/onet.v1"
)

type Protocol struct {
	*onet.TreeNodeInstance

	Genesis  *skipchain.SkipBlock
	Finished chan bool
}

func init() {
	network.RegisterMessage(Prompt{})
	network.RegisterMessage(Terminate{})
	// _, _ = onet.GlobalProtocolRegister(Name, New)
}

// func New(node *onet.TreeNodeInstance) (onet.ProtocolInstance, error) {
// 	protocol := &Protocol{TreeNodeInstance: node, Finished: make(chan bool)}

// 	for _, handler := range []interface{protocol.HandlePrompt, protocol.HandleTerminate} {
// 		if err := protocol.RegisterHandler(handler); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return protocol, nil
// }

// func (p *Protocol) Start() error {
// 	return nil
// }
