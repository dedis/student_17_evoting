// jshint esversion: 6

class Election {

    constructor(name, roster, proto, curve) {
	this.name = name;
	this.roster = roster;
	this.proto = proto;
	this.curve = curve;

	const url = misc.extractHostFromUrl(roster.servers[0].Address);
	this.socket = new net.Socket(url + '/nevv', proto);
	
	this.key = null;
	this.hash = null;
	this.ballots = [];
	this.shuffles = [];
	this.shuffled = false;
    }

    generate() {
	const data = {
	    Name: this.name,
	    Roster: {
		List: this.roster.servers
	    }
	};
	return this.socket.send('GenerateRequest', 'GenerateResponse', data).then((data) => {
	    this.hash = misc.uint8ArrayToHex(data.Hash);
	    this.key = crypto.unmarshal(data.Key);

	    console.log(misc.uint8ArrayToHex(crypto.marshal(this.key)));
	});
    }

    cast() {
	const ballot = crypto.elgamalEncrypt(this.key, new Uint8Array([7, 7, 7, 7, 7]));
	const data =  {
	    Election: this.name,
	    Ballot: ballot
	};
	return this.socket.send('CastRequest', 'CastResponse', data).then((data) => {
	    this.ballots.push(ballot);
	});
    }

    shuffle() {
	const data = { Election: this.name };
	return this.socket.send('ShuffleRequest', 'ShuffleResponse', data).then((data) => {
	    this.shuffled = true;
	});
    }

    fetch(node) {
	let order = -1;
	$.each(this.roster.servers, (index, server) => {
	    if (server.Address == node)
		order = index;
	});

	if (order == -1)
	    throw `${node} not part of roster`;

	const data = { Election: this.name, Block: this.ballots.length + order + 1 };
	return this.socket.send('FetchRequest', 'FetchResponse', data).then((data) => {
	    this.shuffles = [];
	    $.each(data.Ballots, (index, ballot) => {
		this.shuffles.push(ballot);
	    });
	});
    }

    decrypt() {
	const data =  {
	    Election: this.name
	};
	return this.socket.send('DecryptionRequest', 'DecryptionResponse', data);
    }
}
