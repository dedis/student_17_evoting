package chains

import (
	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

const (
	// Election stages.
	RUNNING = iota
	SHUFFLED
	DECRYPTED
	FINISHED
	CORRUPT
)

// Election is the base object for a voting procedure. It is stored
// in the second Skipblock right after the (empty) genesis block. A reference
// to the election Skipchain is appended to the master Skipchain upon opening.
type Election struct {
	Name    string   // Name of the election.
	Creator uint32   // Creator is the election responsible.
	Users   []uint32 // Users is the list of registered voters.

	ID     skipchain.SkipBlockID // ID is the hash of the genesis block.
	Roster *onet.Roster          // Roster is the set of responsible nodes
	Key    abstract.Point        // Key is the DKG public key.
	Data   []byte                // Data can hold any marshable structure.
	Stage  uint32                // Stage indicates the phase of the election.

	Description string // Description in string format.
	End         string // End (termination) date.
}

func init() {
	network.RegisterMessages(Election{}, Ballot{}, Box{}, Mix{}, Partial{})
}

// FetchElection retrieves the election object from its skipchain and sets its stage.
func FetchElection(roster *onet.Roster, id skipchain.SkipBlockID) (*Election, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(chain[1].Data)
	election := blob.(*Election)

	n, num_mixes, num_partials := len(election.Roster.List), 0, 0
	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if _, ok := blob.(*Mix); ok {
			num_mixes++
		} else if _, ok := blob.(*Partial); ok {
			num_partials++
		}
	}

	if num_mixes == 0 && num_partials == 0 {
		election.Stage = RUNNING
	} else if num_mixes == n && num_partials == 0 {
		election.Stage = SHUFFLED
	} else if num_mixes == n && num_partials == n {
		election.Stage = DECRYPTED
	} else {
		election.Stage = CORRUPT
	}
	return election, nil
}

// Store appends a given structure to the election skipchain.
func (e *Election) Store(data interface{}) error {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return err
	}

	if _, err := client.StoreSkipBlock(chain[len(chain)-1], e.Roster, data); err != nil {
		return err
	}
	return nil
}

// Ballots accumulates all the ballots while only keeping the last ballot for each user.
func (e *Election) Box() (*Box, error) {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	// Use map to only included a user's last ballot.
	mapping := make(map[uint32]*Ballot)
	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if ballot, ok := blob.(*Ballot); ok {
			mapping[ballot.User] = ballot
		}
	}

	ballots := make([]*Ballot, 0)
	for _, ballot := range mapping {
		ballots = append(ballots, ballot)
	}
	return &Box{Ballots: ballots}, nil
}

// Mixes returns all mixes created by the roster conodes.
func (e *Election) Mixes() ([]*Mix, error) {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	mixes := make([]*Mix, 0)
	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if mix, ok := blob.(*Mix); ok {
			mixes = append(mixes, mix)
		}
	}

	return mixes, nil
}

// Partials returns the partial decryption for each roster conode.
func (e *Election) Partials() ([]*Partial, error) {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	partials := make([]*Partial, 0)
	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if partial, ok := blob.(*Partial); ok {
			partials = append(partials, partial)
		}
	}

	return partials, nil
}

// IsUser checks if a given user is a registered voter for the election.
func (e *Election) IsUser(user uint32) bool {
	for _, u := range e.Users {
		if u == user {
			return true
		}
	}
	return false
}

// IsUser checks if a given user is the creator of the election.
func (e *Election) IsCreator(user uint32) bool {
	return user == e.Creator
}

// storeBallots appends a list of ballots to the election skipchain.
func (e *Election) storeBallots(ballots []*Ballot) error {
	for _, ballot := range ballots {
		if err := e.Store(ballot); err != nil {
			return err
		}
	}
	return nil
}

// storeBallots appends a list of mixes to the election skipchain.
func (e *Election) storeMixes(mixes []*Mix) error {
	for _, mix := range mixes {
		if err := e.Store(mix); err != nil {
			return err
		}
	}
	return nil
}

// storeBallots appends a list of partials to the election skipchain.
func (e *Election) storePartials(partials []*Partial) error {
	for _, partial := range partials {
		if err := e.Store(partial); err != nil {
			return err
		}
	}
	return nil
}
