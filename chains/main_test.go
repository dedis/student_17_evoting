package chains

import (
	"testing"

	"gopkg.in/dedis/onet.v1"

	"github.com/qantik/nevv/crypto"
)

var master *Master
var election *Election

func TestMain(m *testing.M) {
	local := onet.NewLocalTest()
	defer local.CloseAll()

	_, roster, _ := local.GenBigTree(3, 3, 1, true)

	chain, _ := New(roster, nil)
	master = &Master{ID: chain.Hash, Roster: roster, Admins: []uint32{0, 1}}
	Store(master.Roster, master.ID, master)
	Store(master.Roster, master.ID, &Link{})

	chain, _ = New(roster, nil)
	b1 := &Ballot{User: 0, Alpha: crypto.Random(), Beta: crypto.Random()}
	b2 := &Ballot{User: 1, Alpha: crypto.Random(), Beta: crypto.Random()}
	box := &Box{Ballots: []*Ballot{b1, b2}}
	m1 := &Mix{Ballots: []*Ballot{b1, b2}, Proof: []byte{}}
	m2 := &Mix{Ballots: []*Ballot{b1, b2}, Proof: []byte{}}
	m3 := &Mix{Ballots: []*Ballot{b1, b2}, Proof: []byte{}}

	election = &Election{
		ID:     chain.Hash,
		Roster: roster,
		Key:    crypto.Random(),
		Data:   []byte{},
	}

	Store(election.Roster, election.ID, election)
	Store(election.Roster, election.ID, b1)
	Store(election.Roster, election.ID, b2)
	Store(election.Roster, election.ID, box)
	Store(election.Roster, election.ID, m1)
	Store(election.Roster, election.ID, m2)
	Store(election.Roster, election.ID, m3)

	m.Run()
}
