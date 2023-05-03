import { WsSession } from "./wssession";

const session = new WsSession(
	"ws://localhost:8080",
	"WebCrypto-raw.EC.P-256$BQYAAACBnQAAABgAAABYA",
	(data) => {
		return sign(data, keyPair);
	}
);

// ### Receive list of neighbors
//
// Look at list of existing neighbours, and compare to new list.
//
// For all neighbours that are gone, delete them from the local cache, and
// notify the remaining nodes that it has lost those nodes
//
// For all neighbours that are new, add them to the local cache, and notify
// the remote nodes of the new

session.messageEvents.addEventListener((event) => {
	console.log(event.data);
});
