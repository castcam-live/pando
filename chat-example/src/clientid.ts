import { encodeBase64 } from "ws-key-auth";

export async function generateKeys() {
	return await crypto.subtle.generateKey(
		{ name: "ECDSA", namedCurve: "P-256" },
		false,
		["sign", "verify"]
	);
}

export async function getClientId(keyPair: CryptoKeyPair) {
	const algo = keyPair.publicKey.algorithm;
	if (algo.name !== "ECDSA" && algo.name !== "ECDH") {
		throw new Error(
			`Unexpected key algorithm "${keyPair.publicKey.algorithm.name}"`
		);
	}
	const encodedRaw = encodeBase64(
		await crypto.subtle.exportKey("raw", keyPair.publicKey)
	);
	return `WebCrypto-raw.EC.${(algo as any).namedCurve}$${encodedRaw}`;
}

export async function sign(
	data: ArrayBuffer,
	keyPair: CryptoKeyPair
): Promise<ArrayBuffer> {
	const signature = await crypto.subtle.sign(
		{ name: "ECDSA", hash: "SHA-256" },
		keyPair.privateKey,
		data
	);
	return signature;
}
