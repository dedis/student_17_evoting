// jshint esversion: 6

function bufToHex(buffer) {
    return Array.from(buffer).map((element, index) => {
	return ('00' + element.toString(16)).slice(-2);
    }).join('');
}

function hexToUint8Array(string) {
    return new Uint8Array(Math.ceil(string.length / 2)).map((element, index) => {
	return parseInt(string.substr(index * 2, 2), 16);
    });
}

function unmarshal(curve, bytes) {
    const odd = (bytes[31] >> 7) === 1;
    if (odd)
	bytes[0] -= 19;

    return curve.curve.pointFromY(bytes.reverse(), odd);
}

function marshal(point) {
    point.normalize();

    const buffer = hexToUint8Array(point.y.toString(16, 2));
    buffer[0] ^= (point.x.isOdd() ? 1 : 0) << 7;

    return buffer.reverse();
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
	let alpha = bufToHex(element.Alpha.X);
	let beta = bufToHex(element.Beta.X);
	$(table).append(`<tr><td>${alpha}<br>${beta}</td></tr>`);
    });
}
