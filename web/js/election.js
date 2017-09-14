// jshint esversion: 6

class Election {

    constructor(name, roster, proto, curve) {
	this.name = name;
	this.roster = roster;
	this.proto = proto;
	this.curve = curve;

	const url = extractUrl(roster.servers[0].Address);
	this.socket = new Socket(url + '/nevv', proto);
	
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
	    this.hash = bufToHex(data.Hash);
	    this.key = unmarshal(this.curve, data.Key);

	    console.log(bufToHex(marshal(this.key)));
	});
    }

    cast() {
	const ballot = encrypt(this.curve, this.key);
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

function Socket(url, protobuf) {
    this.url = 'ws://' + url;
    this.protobuf = protobuf;

    this.send = (request, response, data) => {
	return new Promise((resolve, reject) => {
	    const ws = new WebSocket(this.url + '/' + request);
	    ws.binaryType = 'arraybuffer';

	    const requestModel = this.protobuf.lookup(request);
	    if (requestModel === undefined)
		reject(new Error('Model ' + request + ' not found'));
	    const responseModel = this.protobuf.lookup(response);
	    if (responseModel === undefined)
		reject(new Error('Model ' + response + ' not found'));

	    ws.onopen = () => {
		const message = requestModel.create(data);
		const marshal = requestModel.encode(message).finish();
		ws.send(marshal);
	    };

	    ws.onmessage = (event) => {
		ws.close();
		const buffer = new Uint8Array(event.data);
		const unmarshal = responseModel.decode(buffer);
		resolve(unmarshal);
	    };

	    ws.onclose = (event) => {
		if (!event.wasClean)
		    reject(new Error(event.reason));
	    };

	    ws.onerror = (error) => {
		reject(new Error('Could not connect to ' + this.url));
	    };
	});
    };
}
