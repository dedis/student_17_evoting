/*
Package shuffle implements the the Neff shuffle protocol.

Each participating node creates a verifiable shuffle with a corresponding proof
of the last last block in election skipchain. This block is either a box of
encrypted ballot (in case of the root node) or a mix of the previous node. Every
newly created shuffle is appended to the chain before the next node is prompted
to create its shuffle. The leaf node notifies the root upon storing its shuffle,
which terminates the protocol. The individual mixes are not verified here but
only in a later stage of the election.

Schema:

        [Prompt]            [Prompt]            [Prompt]         [Terminate]
  Root ------------> Node1 ------------> Node2 --> ... --> Leaf ------------> Root

The protocol can only be started by the election's creator and is non-repeatable.
*/
package shuffle
