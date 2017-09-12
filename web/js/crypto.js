// jshint esversion: 6

function embed(curve, message) {
    for(let i = 0; i < 100; i++) {
	console.log(i);
	let random = curve.genKeyPair().getPublic();
	let x = random.x.toString(16, 2).substring(0, 58) + '111111';
	// let x = random.x.toString(16, 2).substring(0, 58) + '616263';
	try {
	    let key = curve.curve.pointFromX(x);
	    let key1 = key.mul(curve.n);
	    if (key.validate() && key1.isInfinity())
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

function eencrypt(curve, key) {
    // let message = curve.genKeyPair().getPublic();
    let message = embed(curve);
    console.log(message);
    console.log(reverse(message.x.toString(16, 2)));
    console.log(reverse(message.y.toString(16, 2)));
    console.log(reverse(message.z.toString(16, 2)));

    const k = curve.genKeyPair().getPrivate();
    const K = curve.g.mul(k);
    console.log(K);
    console.log(reverse(K.x.toString(16, 2)));
    console.log(reverse(K.y.toString(16, 2)));
    console.log(reverse(K.z.toString(16, 2)));

    const S = key.mul(k);
    const C = S.add(message);
    console.log(C);
    console.log(reverse(C.x.toString(16, 2)));
    console.log(reverse(C.y.toString(16, 2)));
    console.log(reverse(C.z.toString(16, 2)));

    const sx = hexToUint8Array(K.x.toString(16, 2)).reverse();
    const sy = hexToUint8Array(K.y.toString(16, 2)).reverse();
    const sz = hexToUint8Array(K.z.toString(16, 2)).reverse();

    const cx = hexToUint8Array(C.x.toString(16, 2)).reverse();
    const cy = hexToUint8Array(C.y.toString(16, 2)).reverse();
    const cz = hexToUint8Array(C.z.toString(16, 2)).reverse();

    // console.log(sx.length, sy.length, sz.length, cx.length, cy.length, cz.length);

    const z = new Uint8Array([1, 0, 0, 0, 0, 0, 0, 0,
			      0, 0, 0, 0, 0, 0, 0, 0,
			      0, 0, 0, 0, 0, 0, 0, 0,
			      0, 0, 0, 0, 0, 0, 0, 0]);

    return { Alpha: { X: sx, Y: sy, Z: sz.length != 32 ? z : sz },
	     Beta: { X: cx, Y: cy, Z: cz.length != 32 ? z : cz },
	     Alpha1: marshal(K), Beta1: marshal(C)};
}

function check(pair) {
    return pair.Alpha.X.length == 32 && pair.Alpha.Y.length == 32 &&
	pair.Alpha.Z.length == 32 && pair.Beta.X.length == 32 &&
	pair.Beta.Y.length == 32 && pair.Beta.Z.length == 32;
}

function encrypt(curve, key) {
    let pair;
    do {
	pair = eencrypt(curve, key);
    } while (!check(pair));

    return pair;
}
