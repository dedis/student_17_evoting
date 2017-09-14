// jshint esversion: 6

class Election {

    constructor(name, roster, proto, curve) {
	this.name = name;
	this.roster = roster;
	this.proto = proto;
	this.curve = curve;
	
	this.key = null;
	this.hash = null;
	this.ballots = [];
	this.shuffles = [];
	this.shuffled = false;
    }

    generate() {
	const request = this.proto.lookup('GenerateRequest');
	const response = this.proto.lookup('GenerateResponse');
	const data = {
	    Name: this.name,
	    Roster: {
		List: this.roster.servers
	    }
	};

	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'GenerateRequest', request, data).then((data) => {
	    const buffer = new Uint8Array(data);
	    const decoded = response.decode(buffer);

	    this.hash = bufToHex(decoded.Hash);
	    this.key = unmarshal(this.curve, decoded.Key);
	    console.log(reverse(this.key.x.toString(16, 2)));
	    console.log(reverse(this.key.y.toString(16, 2)));
	    console.log(reverse(this.key.z.toString(16, 2)));

	    console.log(bufToHex(marshal(this.key)));
	});
    }

    cast() {
	const request = this.proto.lookup('CastRequest');
	const response = this.proto.lookup('CastResponse');
	const ballot = encrypt(this.curve, this.key);
	const data =  {
	    Election: this.name,
	    Ballot: ballot
	};
	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'CastRequest', request, data).then((data) => {
	    this.ballots.push(ballot);
	});
    }

    shuffle() {
	const request = this.proto.lookup('ShuffleRequest');
	const response = this.proto.lookup('ShuffleResponse');
	const data = { Election: this.name };

	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'ShuffleRequest', request, data).then(() => {
	    this.shuffled = true;
	});
    }

    fetch(node) {
	const request = this.proto.lookup('FetchRequest');
	const response = this.proto.lookup('FetchResponse');

	let order = -1;
	$.each(this.roster.servers, (index, server) => {
	    if (server.Address == node)
		order = index;
	});

	if (order == -1)
	    throw `${node} not part of roster`;

	const data = { Election: this.name, Block: this.ballots.length + order + 1 };

	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'FetchRequest', request, data).then((data) => {
	    const buffer = new Uint8Array(data);
	    const decoded = response.decode(buffer);
	    this.shuffles = [];
	    $.each(decoded.Ballots, (index, ballot) => {
		this.shuffles.push(ballot);
	    });
	});
    }

    decrypt() {
	const request = this.proto.lookup('DecryptionRequest');
	const response = this.proto.lookup('DecryptionResponse');
	const data =  {
	    Election: this.name
	};
	const address = this.roster.servers[0].Address;

	return Socket.send(address, 'DecryptionRequest', request, data);
    }
}

class Socket {

    static send(address, type, model, data) {
	return new Promise((resolve, reject) => {
	    let hostname = address.replace('tcp://', '').split(':');
	    hostname[1] = parseInt(hostname[1]) + 1;
	    hostname = hostname.join(':');
	    
	    const url = `ws://${hostname}/nevv/${type}`;
	    const socket = new WebSocket(url);
	    socket.binaryType = 'arraybuffer';

	    const message = model.create(data);
	    const encoding = model.encode(message).finish();
	    
	    socket.onopen = () => {
	        socket.send(encoding);
	    };

	    socket.onmessage = (event) => {
	        socket.close();
	        resolve(event.data);
	    };

	    socket.onclose = (event) => {
		if (!event.wasClean)
		    reject(new Error(event.reason));
	    };
 
	    socket.onerror = (error) => {
	        reject(new Error(`Could not connect to ${url}`));
	    };
	});
    }

}
