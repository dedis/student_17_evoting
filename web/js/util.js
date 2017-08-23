// jshint esversion: 6

function bufToHex(buffer) {
    return Array.prototype.map.call(
	new Uint8Array(buffer), x => ('00' + x.toString(16)).slice(-2)).join('');
}

function hexToUint8Array(string) {
    return new Uint8Array(Math.ceil(string.length / 2)).map((element, index) => {
	return parseInt(string.substr(index * 2, 2), 16);
    });
}

function readFile(input) {
    if (!input.files || !input.files[0])
	throw 'File not found';

    const file = input.files[0];
    const reader = new FileReader();

    return new Promise((resolve, reject) => {
	reader.onload = () => {
	    resolve(reader.result);
	};

	reader.onerror = (error) => {
	    reject(error);
	};

	reader.readAsText(file);
    });
}

function parseRoster(toml) {
    const roster = topl.parse(toml);
    roster.servers.forEach((server) => {
        const pub = Uint8Array.from(atob(server.Public), c => c.charCodeAt(0));
        const url = 'https://dedis.epfl.ch/id/' + bufToHex(pub);
        server.Id = new UUID(5, 'ns:URL', url).export();
    });

    return roster;
}

function extractUrl(string) {
    let url = string.replace('tcp://', '').split(':');
    url[1] = parseInt(url[1]) + 1;

    return url.join(':');
}

function color() {
    const letters = '0123456789ABCDEF';
    let color = '#';
    for (let i = 0; i < 6; i++)
        color += letters[Math.floor(Math.random() * 16)];

    return color;
}

function populate(table, data) {
    $(`${table} tbody tr`).remove();
    $.each(data, (index, element) => {
	$(table).append(`<tr><td>${element.Alpha}<br>${element.Beta}</td></tr>`);
    });
}
