// jshint esversion: 6

class Election {

    constructor(name, roster, proto) {
	this.name = name;
	this.roster = roster;
	this.proto = proto;
	
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
	    this.key = bufToHex(decoded.Key);
	    this.hash = bufToHex(decoded.Hash);
	});
    }

    cast(ballot) {
	const request = this.proto.lookup('CastRequest');
	const response = this.proto.lookup('CastResponse');
	const encryption = Encrypt(this.key, ballot);
	const data =  {
	    Election: this.name,
	    Ballot: {
		Alpha: hexToUint8Array(encryption[0]),
		Beta: hexToUint8Array(encryption[1])
	    }
	};
	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'CastRequest', request, data).then((data) => {
	    this.ballots.push({Alpha: encryption[0], Beta: encryption[1]});
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
		this.shuffles.push({Alpha: bufToHex(ballot.Alpha),
				    Beta: bufToHex(ballot.Beta)});
	    });
	});
    }
}

class Socket {

    static send(address, type, model, data) {
	return new Promise((resolve, reject) => {
	    const url = `ws://${extractUrl(address)}/nevv/${type}`;
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
