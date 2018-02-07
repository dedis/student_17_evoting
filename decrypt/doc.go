/*
Package decrypt implements the decryption protocol.

Each participating node begins with verifying the integrity of each mix. If
If all mixes are correct a partial decryption of the last mix is performed using
the node's shared secret from the DKG. The result is the appended to the election
skipchain before prompting the next node. If at least one mix can no be verified
the node won't create a decryption but appends a flag indicating the failure
on the skipchain. The leaf node notifies the root upon completing its turn, which
terminates the protocol.

Schema:

        [Prompt]            [Prompt]            [Prompt]         [Terminate]
  Root ------------> Node1 ------------> Node2 --> ... --> Leaf ------------> Root

The protocol can only be started by the election's creator and is non-repeatable.
*/
package decrypt
