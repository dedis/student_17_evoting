// Package shuffle contains the shuffle protocol of the nevv service.
// It uses a cothority tree instance where every node has exactly one child,
// this means it must be initialized with the fanout argument set to 1.
// Upon start the root node collects the ballots from SkipChain and creates a
// shuffles before appending it to the chain. It then invokes its child node with
// a Prompt message containing said new SkipBlock. The child again creates a mix and
// stores it on the SkipChain before prompting its child.
// The leaf node sends Terminate message back up to the root node after it has completed
// its shuffle. The root node then signals the end of the protocol to the service.
//
//                      Root -> n(1) -> n(2) -> ... -> Leaf
//                       |                              |
//                           <-----------------------
//
// This protocol is in accordance with the mixnet proposed in the Helios whitepaper.
// [https://www.usenix.org/legacy/event/sec08/tech/full_papers/adida/adida.pdf]
package shuffle
