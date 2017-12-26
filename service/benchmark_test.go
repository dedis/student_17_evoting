package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/qantik/nevv/api"
	"github.com/qantik/nevv/chains"
)

func TestBenchmark(t *testing.T) {
	t.Skip()
	n := 50

	lr, _ := service.Link(&api.Link{service.pin, roster, nil, []chains.User{0}})

	users := make([]chains.User, n)
	tokens := make([]string, n)
	for i := 0; i < n; i++ {
		users[i] = chains.User(i)
		lor, _ := service.Login(&api.Login{lr.Master, chains.User(i), []byte{}})
		tokens[i] = lor.Token
	}

	election := &chains.Election{Name: "", Creator: 0, Users: users}
	or, _ := service.Open(&api.Open{tokens[0], lr.Master, election})

	for i := 0; i < n; i++ {
		k, c := encrypt(or.Key, []byte{byte(i)})
		b := &chains.Ballot{chains.User(i), k, c}
		service.Cast(&api.Cast{tokens[i], or.Genesis, b})
	}

	service.Shuffle(&api.Shuffle{tokens[0], or.Genesis})
	r, _ := service.Decrypt(&api.Decrypt{tokens[0], or.Genesis})

	for i := 0; i < n; i++ {
		assert.Equal(t, byte(r.Decrypted.Texts[i].User), r.Decrypted.Texts[i].Data[0])
	}
}
