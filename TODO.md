# Reconnect Player State

- Handle temporary disconnects entirely inside the authoritative Nakama match using in-memory state.
- Use `UserID` as the persistent player identity; treat `SessionID` and `runtime.Presence` as replaceable connection data.
- On `MatchLeave`, remove the presence but retain the player and spatial state for a 30-60 second reconnect grace period.
- Store a disconnect expiry tick on the retained player instead of deleting it immediately.
- On `MatchJoinAttempt`, allow the same `UserID` to reclaim its retained player even when the match has no free slot.
- On `MatchJoin`, replace the old `SessionID` and presence, update the player and spatial-grid indexes, clear the disconnect expiry, and preserve position, characters, health, weapon, and strategy.
- In `MatchLoop`, permanently remove disconnected players whose grace period has expired, then update the match label and player-count snapshot.
- Keep disconnected players out of movement input processing and personalized detection broadcasts while their state is retained.
- The client only needs to reconnect, authenticate, and join the same match; the server restores the retained player state by `UserID`.
- Add tests for reconnecting within the grace period, expiry cleanup, changed session IDs, duplicate active sessions, capacity handling, and spatial-grid consistency.

# Gameplay Protocol

- Add and decode a client input opcode for selecting or changing the player's strategy formation; the server currently always uses the default `compact` strategy.
- Replace the temporary `impact_ratio` values in `weapons.json` with values matched to the final Unity attack animations; combat currently uses these defaults to calculate `impact_tick`.
