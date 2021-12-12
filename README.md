# Pernod

Pernod is a minecraft proxy server.

- Add an external authentication to offline-mode server.
- Display online player's skin on both online or offline player's client.

This program implement this feature by modifies the `ClientboundPlayerInfoPacket`, 
`ClientboundAddPlayerPacket` and `ServerboundTeleportToEntityPacket`.