import { WsSession } from "./wssession";

const session = new WsSession(
	"ws://localhost:8080",
	"WebCrypto-raw.EC.P-256$BQYAAACBnQAAABgAAABYA",
	(data) => {
		return sign(data, keyPair);
	}
);

session.messageEvents.addEventListener((event) => {
	console.log(event.data);
});
