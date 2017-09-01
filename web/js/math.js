// jshint esversion: 6

function encrypt(curve, key) {
    const message = curve.genKeyPair().getPublic();

    const k = curve.genKeyPair().getPrivate();
    const K = curve.g.mul(k);
    const S = key.mul(k);
    const C = S.add(message);

    const sx = hexToUint8Array(S.x.toString(16, 2)).reverse();
    const sy = hexToUint8Array(S.y.toString(16, 2)).reverse();
    const sz = hexToUint8Array(S.z.toString(16, 2)).reverse();

    const cx = hexToUint8Array(C.x.toString(16, 2)).reverse();
    const cy = hexToUint8Array(C.y.toString(16, 2)).reverse();
    const cz = hexToUint8Array(C.z.toString(16, 2)).reverse();

    return { Alpha: { X: sx, Y: sy, Z: sz }, Beta: { X: cx, Y: cy, Z: cz } };
}
