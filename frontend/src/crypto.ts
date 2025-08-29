import nacl from 'tweetnacl';

export async function decrypt(peerPublicKey: Uint8Array, privateKey: Uint8Array, encryptedData: Uint8Array): Promise<Uint8Array> {
    const sharedSecret = nacl.box.before(peerPublicKey, privateKey);

    const iv = encryptedData.slice(0, 12);

    const ciphertext = encryptedData.slice(12);

    const key = await window.crypto.subtle.importKey(
        'raw',
        sharedSecret as BufferSource,
        'AES-GCM',
        false,
        ['decrypt']
    );

    const decryptedData = await window.crypto.subtle.decrypt(
        {
            name: 'AES-GCM',
            iv: iv
        },
        key,
        ciphertext
    );

    return new Uint8Array(decryptedData)
}

export async function encrypt(peerPublicKey: Uint8Array, privateKey: Uint8Array, data: Uint8Array): Promise<Uint8Array> {
    const sharedSecret = nacl.box.before(peerPublicKey, privateKey);

    const iv = crypto.getRandomValues(new Uint8Array(12));

    const key = await window.crypto.subtle.importKey(
        'raw',
        sharedSecret as BufferSource,
        'AES-GCM',
        false,
        ['encrypt']
    );

    const encryptedData = await window.crypto.subtle.encrypt(
        {
            name: 'AES-GCM',
            iv: iv
        },
        key,
        data as BufferSource,
    );

    const ciphertext = new Uint8Array(iv.byteLength + encryptedData.byteLength);
    ciphertext.set(iv, 0);
    ciphertext.set(new Uint8Array(encryptedData), iv.byteLength);

    return ciphertext;
}