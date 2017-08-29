package service

import (
	"github.com/dedis/cothority/skipchain"
	"github.com/qantik/nevv/protocol"
)

// Election is the base data structure of the application. It comprises
// for each involved conode the genesis and the latest appended block as
// well as the generated shared secret from the distributed key generation
// protocol which is run at the inception of a new election.
type Election struct {
	Genesis *skipchain.SkipBlock
	Latest  *skipchain.SkipBlock

	*protocol.SharedSecret
}
