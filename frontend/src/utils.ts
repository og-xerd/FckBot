export function sleep(ms: number) {
  return new Promise(resolve => setTimeout(resolve, ms));
}

export function base64RawUrlDecode(base64url: string): Uint8Array {
  let base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');

  while (base64.length % 4 !== 0) {
    base64 += '=';
  }

  const binaryStr = atob(base64);

  const bytes = new Uint8Array(binaryStr.length);
  for (let i = 0; i < binaryStr.length; i++) {
    bytes[i] = binaryStr.charCodeAt(i);
  }

  return bytes;
}


export function base64RawUrlEncode(data: any): string {
  let str;
  if (data instanceof Uint8Array) {
    str = String.fromCharCode(...data);
  } else if (typeof data === 'string') {
    str = data;
  } else {
    throw new Error('Unsupported data type');
  }

  const base64 = btoa(str);

  return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function uint8ArrayToHex(bytes: Uint8Array): string {
  return Array.from(bytes)
    .map(byte => byte.toString(16).padStart(2, '0'))
    .join('');
}

export function hexToUint8Array(hex: string): Uint8Array {
  if (hex.length % 2 !== 0) {
    throw new Error("Hex string must have even length");
  }
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.slice(i, i + 2), 16);
  }
  return bytes;
}