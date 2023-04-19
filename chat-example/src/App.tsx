import { useEffect, useRef } from "react";
import { Routes, Route, redirect, useParams } from "react-router-dom";
import { generateKeys, getClientId, sign } from "./clientid";
import { WsSession } from "./wssession";

/**
 * Gets a random string of characters
 * @param count The number of characters to get from the random string
 * @returns A random string of characters
 */
function getRandomString(count: number) {
	let text = "";
	const possible =
		"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
	for (let i = 0; i < count; i++) {
		text += possible.charAt(Math.floor(Math.random() * possible.length));
	}
	return text;
}

function Home() {
	useEffect(() => {
		redirect(`/${getRandomString(32)}`);
	}, []);
	return <></>;
}

function ChatRoom() {
	const params = useParams();
	const isCancelledRef = useRef(false);

	useEffect(() => {
		if (!params.chatId) {
			redirect("/");
			return;
		}

		generateKeys().then(async (keys) => {
			const clientId = await getClientId(keys);

			const url = new URL("ws://localhost:3333");
			url.pathname = `/tree/${encodeURIComponent(clientId)}`;

			new WsSession(url, clientId, (data) =>
				sign(data, keys)
			).messageEvents.addEventListener((message) => {
				console.log(message);
			});
		});
	}, []);

	useEffect(() => {
		return () => {
			isCancelledRef.current = true;
		};
	}, []);

	return <></>;
}

function App() {
	return (
		<Routes>
			<Route path="/" element={<Home />} />
			<Route path="/:chatId" element={<Home />} />
		</Routes>
	);
}

export default App;
