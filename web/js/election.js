// jshint esversion: 6

class Election {

    constructor(name, roster, proto) {
	this.name = name;
	this.roster = roster;
	this.proto = proto;
	
	this.key = null;
	this.hash = null;
	this.ballots = [];
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
	const data =  {
	    Name: this.name,
	    Ballot: ballot
	};

	const address = this.roster.servers[0].Address;
	return Socket.send(address, 'CastRequest', request, data).then((data) => {
	    this.ballots.push(ballot);
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

	    socket.onerror = (error) => {
	        reject(new Error(`Could not connect to ${url}`));
	    };
	});
    }

}
