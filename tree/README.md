# tree

This is a document that describes the distributed tree management server.

Summary: nodes will not be auto-arranging themselves into a tree topology; instead it's a server that will hold a representation of how the clients will be related to each other.

Properties of the server's internal representation of a tree will be the following:

- each client will be given IDs of 3 neighbouring nodes
- the graph will be acyclic
- the graph will be undirected

By the definition of the last two-points, the graph will therefore be a tree.

If two non-neighbouring nodes need to communicate with eachother, either the application utilizing the tree will need to design the application to support the act of relaying messages from node-to-node, or the application will need to utilize a third-party server, unrelated to this server

## Protocol

### 1 Connection and authentication

**Step 1**

Client will connect via WebSocket to /tree/{id}/{clientid}, where `{id}` is the tree ID, and `{clientid}` is a ECDSA secp256r1 [NIST-formatted public key]() for the various Key ID versions that are supported).

**Step 2**

Server will then send a [JTD](https://github.com/railtown/rfc/blob/main/type-data-json.md) message of type `CHALLENGE`, containing the data `message`, and it will be a base64-encoded binary data.

An example server `CHALLENGE` message:

```json
{
  "type": "CHALLENGE",
  "data": {
    "message": "kaN9OArkHprc3FC7Iwlp4561oiD+CnNA4oEGOVcDlKk="
  }
}
```

**Step 3**

Client will then sign the extracted base64 data, and respond to the server with a JTD message of type `CHALLENGE_RESPONSE`, containing the fields `message` and `signature`.

An example `CHALLENGE_RESPONSE` message:

```json
{
  "type": "CHALLENGE_RESPONSE",
  "data": {
    "message": "",
    "signature": ""
  }
}
```
