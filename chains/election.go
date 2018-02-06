package chains

import (
	"errors"

	"gopkg.in/dedis/cothority.v1/skipchain"
	"gopkg.in/dedis/crypto.v0/abstract"
	"gopkg.in/dedis/onet.v1"
	"gopkg.in/dedis/onet.v1/network"
)

const (
	STAGE_RUNNING = iota
	STAGE_SHUFFLED
	STAGE_DECRYPTED
	STAGE_FINISHED
	STAGE_CORRUPT
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
	network.RegisterMessages(Election{}, Ballot{}, Box{}, Mix{})
}

// FetchElection retrieves the election object from its skipchain and sets its stage.
func FetchElection(roster *onet.Roster, id skipchain.SkipBlockID) (*Election, error) {
	chain, err := chain(roster, id)
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(chain[1].Data)
	election := blob.(*Election)

	n := len(election.Roster.List)
	num_boxes, num_mixes, num_partials := 0, 0, 0

	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if _, ok := blob.(*Box); ok {
			num_boxes++
		} else if _, ok := blob.(*Mix); ok {
			num_mixes++
		} else if _, ok := blob.(*Partial); ok {
			num_partials++
		}
	}

	if num_boxes == 0 && num_mixes == 0 && num_partials == 0 {
		election.Stage = STAGE_RUNNING
	} else if num_boxes == 1 && num_mixes == n && num_partials == 0 {
		election.Stage = STAGE_SHUFFLED
	} else if num_boxes == 1 && num_mixes == n && num_partials == n {
		election.Stage = STAGE_DECRYPTED
	} else {
		election.Stage = STAGE_CORRUPT
	}
	return election, nil
}

// Ballots accumulates all the casted ballots while only keeping the last ballot
// for each user.
func (e *Election) Ballots() (*Box, error) {
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

// Box returns all casted ballots wrapped in a box structure.
func (e *Election) Box() (*Box, error) {
	if e.Stage == STAGE_RUNNING {
		return e.Ballots()
	}

	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	for _, block := range chain {
		_, blob, _ := network.Unmarshal(block.Data)
		if box, ok := blob.(*Box); ok {
			return box, nil
		}
	}

	return nil, errors.New("Could not create box")
}

// Latest returns the last block of the election skipchain.
func (e *Election) Latest() (network.Message, error) {
	chain, err := chain(e.Roster, e.ID)
	if err != nil {
		return nil, err
	}

	_, blob, _ := network.Unmarshal(chain[len(chain)-1].Data)
	return blob, nil
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
