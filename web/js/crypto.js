// jshint esversion: 6

function embed(curve, message) {
    if (message.constructor !== Uint8Array)
	throw "Message type is not Uint8Array";
    if (message.length > 29)
	throw "Oversized (> 29 bytes) message";

    const size = message.length;
    
    for (;;) {
	let random = curve.genKeyPair().getPublic();
	let bytes = hexToUint8Array(random.y.toString(16, 2)).reverse();
	bytes[0] = size;
	bytes.set(message, 1);

	try {
	    let key = unmarshal(curve, bytes);
	    if (key.validate() && key.mul(curve.n).isInfinity())
		return key;
	} catch(err) {}
    }
}

function reverse(string) {
    let reversed = '';
    for (let i = 0; i < string.length-1; i += 2)
	reversed = string.substring(i, i+2) + reversed;

    return reversed;
}

function encrypt(curve, key) {
    let message = embed(curve, new Uint8Array([7, 7, 7, 7, 7]));
    console.log(bufToHex(marshal(message)));

    const k = curve.genKeyPair().getPrivate();
    const K = curve.g.mul(k);
    const S = key.mul(k);
    const C = S.add(message);

    return { Alpha: marshal(K), Beta: marshal(C) };
}
