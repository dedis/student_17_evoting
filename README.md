# nevv

[![Build Status](https://travis-ci.org/dedis/student_17_evoting.svg?branch=master)](https://travis-ci.org/dedis/student_17_evoting)
[![Coverage Status](https://coveralls.io/repos/github/dedis/student_17_evoting/badge.svg?branch=master)](https://coveralls.io/github/dedis/student_17_evoting?branch=master&service=github)

nevv is a decentralized and distributed e-voting architecture based on Neff shuffles. It is based
on the ideas laid out in the Helios project replacing the conventional database storage
infrastructure by a Skipchain, a blockchain alternative developped at DEDIS.

### Installation
```shell
git clone https://github.com/dedis/student_17_evoting
cd student_17_evoting
go test -v ./...
```

The system's backbone is a master Skipchain containing the frontend's public key,
a list of servers (nodes) handling future elections and a list of administrators who have
priviledged possibilities like setting up elections. The master Skipchain can only be
created through the command line tool thus only with physical access to one of the servers.

### Creation of the master Skipchain
```shell
# Setup 5 local servers (nodes) with logging level 3.
# This generates a roster file (public.toml) with the server identities. The roster in the master
# master Skipchain must be a subset of the roster in public.toml. For simplicity,
# we reuse the roster in public.toml for the master Skipchain. 
./setup.sh run 5 3
cd cli
# Get the pin of a server
go run cli.go -roster=../public.toml
# Create the master Skipchain.
# Key has to be in base 64 representation.
# Admins has to be a list of comma-separated numbers, i.e 100,200,300
go run cli.go -pin=[pin] -roster=../public.toml -key=[frontend key] -admins=[list of admins]
# If the creation was successful the identifier of the master Skipchain is returned.
```

An election's life cycle is then determined by the following steps.

1. An administrator logs in and creates a new election. This prompts the servers in the
master Skipchain to set up a new Skipchain for the election and run a distributed key
generation protocol to generate a public, private key pair where the private key where each
node hold a share of the private key. This key is not to be reconstructed (for decryption)
until after the election is completed and shuffled. The identifier of the election Skipchain
is appended to the master Skipchain.
The generated public key and the identifier of the election Skipchain are then returned to the
frontend.

2. User which are part of the election's voter list can now log in to the service and cast
their ballots encrypted with public key from the distributed key generation protocol. Similar
to the Helios project nevv does not aim to obfuscate the voter's identity but only the data
of his actual ballot. A user is able to vote as many time as he wishes, only the last ballot
is taken into account.

3. To end an election the creator logs in and initiates the shuffle protocol. In this protocol
all nodes, one after another, re-encrypt the ballots. The resulting shuffle is then appended
to the election's Skipchain. The shuffle purpose is to be verifiable by all participators in
an election, meaning every voter can check if his ballot has been included by the system without
being compromised in any malicious way.

4. An election is finally terminated when the creator prompts the system to decrypt the
shuffled ballots. The nodes accumulate their shared secret for the decryption and the resulting
plaintexts are appended to the election's Skipchain. The evaluation of the election's result
is left to the frontend.

**Note:** At all points during an election's life cycle the users are able to retrieve the
encrpyted ballots. When the shuffle and decryption have been performed they can also be
requested.

### API
The communication with the servers is handled over protocol buffers. The API is laid out
in ```api/api.proto```.

### References
