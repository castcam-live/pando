import { connect } from "ws-key-auth";

type EventEmitter<T> = {
	addEventListener: (listener: (value: T) => void) => () => void;
};

type Subject<T> = EventEmitter<T> & {
	emit: (value: T) => void;
};

function createSubject<T>(): Subject<T> {
	const listeners = new Set<(value: T) => void>();

	return {
		addEventListener: (listener: (value: T) => void) => {
			listeners.add(listener);
			return () => {
				listeners.delete(listener);
			};
		},
		emit: (value: T) => {
			for (const listener of listeners) {
				listener(value);
			}
		},
	};
}

type States = "CONNECTING" | "CONNECTED" | "DISCONNECTED";

/**
 * A simple state machine that can be used to track the state of a WebSocket
 */
class ConnectionState {
	private emitter: Subject<States> = createSubject();

	constructor(private state: States) {}

	public get(): States {
		return this.state;
	}

	public set(state: States) {
		this.state = state;
		this.emitter.emit(state);
	}

	public getStateChangeEvents(): EventEmitter<States> {
		return this.emitter;
	}
}

/**
 * A WebSocket session that uses the ws-key-auth protocol to authenticate, and
 * if the WebSocket connection closes, automatically reconnects, along with the
 * authentication handshake.
 */
export class WsSession {
	private connectionState: ConnectionState = new ConnectionState(
		"DISCONNECTED"
	);
	private closed: boolean = false;
	private _messageEvents: Subject<MessageEvent> = createSubject();

	constructor(
		private address: string | URL,
		private clientId: string,
		private sign: (data: ArrayBuffer) => Promise<ArrayBuffer>
	) {
		// TODO: do something about the error
		this.connect().catch(console.error);
	}

	private messageBuffer: string[] = [];

	private ws: WebSocket | null = null;

	private async connect() {
		if (this.closed) return;

		this.ws = new WebSocket(this.address);

		this.ws.addEventListener("close", () => {
			this.connectionState.set("DISCONNECTED");
			this.connect();
		});

		// TODO: handle other cases when the connection closes

		this.ws.addEventListener("message", (event) => {
			if (this.connectionState.get() === "CONNECTED") {
				// Only messages received after a successful connection is what we care
				// about

				this._messageEvents.emit(event);
			}
		});

		this.connectionState.set("CONNECTING");

		await connect(this.ws, this.clientId, this.sign);

		this.connectionState.set("CONNECTED");

		for (const message of this.messageBuffer) {
			this.ws.send(message);
		}
	}

	/**
	 * Sends data over the WebSocket connection
	 * @param data The data to send
	 */
	send(data: string) {
		if (this.connectionState.get() === "CONNECTED") {
			this.ws?.send(data);
		} else {
			this.messageBuffer.push(data);
		}
	}

	/**
	 * Closes the WebSocket connection, as well as never reconnecting afterwards
	 */
	close() {
		this.closed = true;
		this.ws?.close();
	}

	/**
	 * Returns an event emitter that emits whenever the connection state changes
	 */
	get stateChangeEvents(): EventEmitter<States> {
		return this.connectionState.getStateChangeEvents();
	}

	/**
	 * Returns an event emitter that emits whenever a message is received
	 */
	get messageEvents(): EventEmitter<MessageEvent> {
		return this._messageEvents;
	}
}
