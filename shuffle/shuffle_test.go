package shuffle

// var serviceID onet.ServiceID

// var box = &chains.Box{
// 	Ballots: []*chains.Ballot{
// 		&chains.Ballot{
// 			User:  chains.User(0),
// 			Alpha: crypto.Suite.Point(),
// 			Beta:  crypto.Suite.Point(),
// 		},
// 		&chains.Ballot{
// 			User:  chains.User(1),
// 			Alpha: crypto.Suite.Point(),
// 			Beta:  crypto.Suite.Point(),
// 		},
// 	},
// }

// type service struct {
// 	*onet.ServiceProcessor

// 	secret *dkg.SharedSecret
// }

// func init() {
// 	serviceID, _ = onet.RegisterNewService(Name, newService)
// }

// func TestProtocol(t *testing.T) {
// 	local := onet.NewLocalTest()
// 	defer local.CloseAll()
// 	nodes, _, tree := local.GenBigTree(5, 5, 1, true)

// 	services := local.GetServices(nodes, serviceID)

// 	instance, _ := services[0].(*service).CreateProtocol(Name, tree)
// 	protocol := instance.(*Protocol)
// 	protocol.Key = crypto.Suite.Point()
// 	protocol.Box = box
// 	protocol.Start()

// 	select {
// 	case <-protocol.Finished:
// 		assert.Equal(t, 2, len(protocol.Mix.Ballots))
// 		assert.Equal(t, protocol.Name(), protocol.Mix.Node)
// 	case <-time.After(2 * time.Second):
// 		t.Fatal("Protocol timeout")
// 	}
// }

// func (s *service) NewProtocol(n *onet.TreeNodeInstance, c *onet.GenericConfig) (
// 	onet.ProtocolInstance, error) {

// 	switch n.ProtocolName() {
// 	case Name:
// 		instance, err := New(n)
// 		if err != nil {
// 			return nil, err
// 		}
// 		protocol := instance.(*Protocol)
// 		protocol.Key = crypto.Suite.Point()
// 		protocol.Box = box

// 		go func() {
// 			<-protocol.Finished
// 			fmt.Println(protocol.Mix.Node)
// 		}()

// 		return protocol, nil
// 	default:
// 		return nil, errors.New("Unknown protocol")
// 	}
// }

// func newService(ctx *onet.Context) onet.Service {
// 	return &service{ServiceProcessor: onet.NewServiceProcessor(ctx)}
// }
