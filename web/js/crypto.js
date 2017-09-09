// jshint esversion: 6

function embed(curve, message) {
    for(let i = 0; i < 100; i++) {
	let random = curve.genKeyPair().getPublic();
	let x = random.x.toString(16, 2).substring(0, 58) + '616263';
	try {
	    let key = curve.curve.pointFromX(x);
	    if (key.validate())
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
    let message = curve.genKeyPair().getPublic();
    console.log(reverse(message.x.toString(16, 2)));
    console.log(reverse(message.y.toString(16, 2)));
    console.log(reverse(message.z.toString(16, 2)));

    //for (let i = 0; i < 100; i++) {
    //	// const message = embed(curve, '');
    //	message = curve.genKeyPair().getPublic();
    //	console.log(reverse(message.x.toString(16, 2)));
    //	console.log(reverse(message.y.toString(16, 2)));
    //	console.log(reverse(message.z.toString(16, 2)));
    //}

    const k = curve.genKeyPair().getPrivate();
    const K = curve.g.mul(k);
    // K.normalize();
    const S = key.mul(k);
    const C = S.add(message);
    // C.normalize();

    const sx = hexToUint8Array(K.x.toString(16, 2)).reverse();
    const sy = hexToUint8Array(K.y.toString(16, 2)).reverse();
    const sz = hexToUint8Array(K.z.toString(16, 2)).reverse();

    const cx = hexToUint8Array(C.x.toString(16, 2)).reverse();
    const cy = hexToUint8Array(C.y.toString(16, 2)).reverse();
    const cz = hexToUint8Array(C.z.toString(16, 2)).reverse();

    console.log(sx.length, sy.length, sz.length, cx.length, cy.length, cz.length);

    const z = new Uint8Array([1, 0, 0, 0, 0, 0, 0, 0,
			     0, 0, 0, 0, 0, 0, 0, 0,
			     0, 0, 0, 0, 0, 0, 0, 0,
			     0, 0, 0, 0, 0, 0, 0, 0]);

    return { Alpha: { X: sx, Y: sy, Z: sz.length != 32 ? z : sz },
	     Beta: { X: cx, Y: cy, Z: cz.length != 32 ? z : cz } };
}
