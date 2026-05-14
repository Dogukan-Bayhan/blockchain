# P2P Deferred Tasks

This file tracks P2P improvements intentionally postponed while the basic peer
discovery and gossip foundation is being built.

## Peer Address Management

- Replace `knownPeers map[string]time.Time` with a richer structure:

  ```go
  type KnownPeer struct {
      Addr         string
      LastSeen     time.Time
      LastAttempt  time.Time
      FailureCount int
  }
  ```

- Add retry backoff for known peer dialing.
- Increase `FailureCount` after failed dial attempts.
- Reset `FailureCount` after a successful connection.
- Stop aggressively retrying peers that recently failed.
- Remove or deprioritize peers after repeated failures.
- Decide the cleanup policy, for example:
  - remove after 5 failed attempts in the learning version;
  - use longer stale-peer rules in a more realistic version.

## Duplicate Connection Policy

- Use `remoteNodeID` as the primary duplicate detection key after handshake.
- Close duplicate connections between the same two nodes.
- Decide which duplicate connection survives:
  - keep outbound;
  - keep inbound;
  - or use deterministic tie-break by node ID.

## Peer State Safety

- Protect mutable `Peer` fields with a mutex or atomic accessors:
  - `handshakeDone`
  - `remoteNodeID`
  - `remoteP2PAddr`
  - handshake flags
- Update `Manager` reads to use safe peer snapshot methods.
- Run `go test -race ./...` after adding synchronization.

## Connection Limits

- Add maximum inbound peer count.
- Add maximum outbound peer count.
- Reject or close excess connections.
- Avoid dialing known peers if the outbound limit is already reached.

## Peer Scoring

- Track basic peer behavior:
  - invalid message payloads;
  - message before handshake;
  - repeated connection failures;
  - useful addr responses;
  - successful handshakes.
- Use scores to prioritize dialing and future gossip.

## Ban / Misbehavior Policy

- Add a temporary ban list for clearly bad peers.
- Ban malformed or abusive peers for a fixed duration.
- Skip banned peers in known peer dialing.

## Persistent Peer Storage

- Store known peers on disk.
- Reload known peers on node startup.
- Keep last seen, last attempt, and failure count metadata.

## Node Identity

- Replace generated string IDs such as `node-5005` with cryptographic identity.
- Add an Ed25519 node key pair.
- Derive `node_id` from the public key.
- Add handshake proof so a peer can prove it owns its node ID.

## NAT / Advertised Address

- Distinguish listen address from advertised address.
- Add a config flag for advertised P2P address.
- Learn observed address from connected peers.
- Later investigate UPnP, NAT-PMP, hole punching, or relay support.

## Graceful Shutdown Improvements

- Add `sync.WaitGroup` to wait for manager goroutines to exit.
- Make `Manager.Stop()` idempotent with `sync.Once`.
- Add configurable shutdown timeout.

## Gossip Foundation

- Add `Peer.SendMessage` as a public/safe send method.
- Add `Manager.Broadcast(msg, exceptPeer)` for gossip.
- Add seen caches:
  - `seenTx`
  - `seenBlocks`
- Start with transaction gossip:
  - `inv`
  - `getdata`
  - `tx`
- Then add block gossip:
  - `inv`
  - `getdata`
  - `block`
