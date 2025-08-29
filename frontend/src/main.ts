import { sha256 } from '@noble/hashes/sha2.js';
import { sha3_256 } from '@noble/hashes/sha3.js';
import { blake3 } from '@noble/hashes/blake3.js';
import { decrypt, encrypt } from "./crypto";
import { hexToUint8Array, base64RawUrlDecode, base64RawUrlEncode, sleep } from "./utils";
import nacl from 'tweetnacl';
import { FckBot } from './config';

function countLeadingZeros(hashBuffer: Uint8Array) {
    let zeros = 0;

    for (let byte of hashBuffer) {
        for (let i = 7; i >= 0; i--) {
            if ((byte >> i) & 1) return zeros;
            zeros++;
        }
    }
    return zeros;
}

function hashMeetsDifficulty(hashBuffer: Uint8Array, difficulty: number): boolean {
    return countLeadingZeros(hashBuffer) >= difficulty;
}

function solveChallenge(type: string, challenge: string, difficulty: number, hashAlgorithm: string): number {
    const challengeEnc = hexToUint8Array(challenge);

    if (challengeEnc.length != 16) {
      throw new Error("challenge length error");
    }

    if (type == "pow") {
        let counter = 0;

        while (true) {
            const nonceBytes = new Uint8Array(4);
            const view = new DataView(nonceBytes.buffer);
            view.setUint32(0, counter);

            const input = new Uint8Array(challengeEnc.length + nonceBytes.length);
            input.set(challengeEnc, 0);
            input.set(nonceBytes, challengeEnc.length);

            var hashBuffer: Uint8Array = new Uint8Array();

            if (hashAlgorithm == "sha256") {
                hashBuffer = sha256(input);
            } else if (hashAlgorithm == "sha3-256") {
                hashBuffer = sha3_256(input);
            } else if (hashAlgorithm == "blake3") {
                hashBuffer = blake3(input);
            }

            if (hashMeetsDifficulty(hashBuffer, difficulty)) {
                return counter;
            }

            counter++;
        }
    }

    return 0;
}

async function getChallenge(): Promise<{ challenge: any; peerPublicKey: Uint8Array; keyPair: nacl.BoxKeyPair }> {
    const keypair = nacl.box.keyPair();

    const response = await fetch(FckBot.config.challengeUrl, {
        method: "POST",
        body: base64RawUrlEncode(keypair.publicKey),
    });

    if (response.status == 200) {
        const body = await response.json();

        const publicKey = base64RawUrlDecode(body.publicKey);
        const challenge = base64RawUrlDecode(body.challenge);

        const data = await decrypt(publicKey ,keypair.secretKey, challenge);

        const decodeData = new TextDecoder().decode(data);

        return { challenge: JSON.parse(decodeData), peerPublicKey: publicKey, keyPair: keypair };
        
    } else {
        throw new Error("getChallenge: Unknown Error")
    }
}

async function makeAnswer(challenge: any, answer: number, peerPublicKey: Uint8Array, keyPair: nacl.BoxKeyPair): Promise<string> {
    const response = JSON.stringify({
        ...challenge,
        answer: answer,
    })

    const encodedResponse = new TextEncoder().encode(response);
    
    const encryptedResponse = await encrypt(peerPublicKey, keyPair.secretKey, encodedResponse);

    const encryptedAnswer = new Uint8Array(keyPair.publicKey.byteLength + encryptedResponse.byteLength);
    encryptedAnswer.set(keyPair.publicKey, 0);
    encryptedAnswer.set(new Uint8Array(encryptedResponse), keyPair.publicKey.byteLength);

    return base64RawUrlEncode(encryptedAnswer);
}

FckBot.fetch = async (url: string | URL, options: RequestInit = {}): Promise<Response> => {
    const { challenge, peerPublicKey, keyPair } = await getChallenge();

    const answer = solveChallenge(challenge.type, challenge.challenge, challenge.difficulty, challenge.algorithm);

    const response = await makeAnswer(challenge, answer, peerPublicKey, keyPair);

    await sleep(challenge.latency);

    const customHeader = {
        "x-answer": response,
    };

    const headers = {
        ...(options.headers instanceof Headers
            ? Object.fromEntries(options.headers.entries())
            : options.headers || {}),
        ...customHeader,
    };

    const finalOptions: RequestInit = {
        ...options,
        headers,
    };

    return fetch(url, finalOptions);
}