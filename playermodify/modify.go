package playermodify

import (
	"bytes"
	"fmt"
	"regexp"
	"sync"

	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
)

const blockMsg = "[BLOCKED MESSAGE]"

var filter = regexp.MustCompile(`\${(.*)}`)

type Modifier struct {
	ProfileMapping sync.Map // map[uuid.UUID]PlayerProperties
}

func (m *Modifier) ModifyClientboundPacket(p *packet.Packet) error {
	switch packetid.ClientboundPacketID(p.ID) {
	case packetid.ClientboundPlayerInfoUpdate:
		var (
			EntityID   packet.VarInt
			PlayerUUID packet.UUID
			X, Y, Z    packet.Double
			Yaw, Pitch packet.Byte
		)
		if err := p.Scan(&EntityID, &PlayerUUID, &X, &Y, &Z, &Yaw, &Pitch); err != nil {
			return err
		}
		if pp, ok := m.ProfileMapping.Load(uuid.UUID(PlayerUUID)); ok {
			pp := pp.(PlayerProperties)
			*p = packet.Marshal(p.ID, EntityID, packet.UUID(pp.ID), X, Y, Z, Yaw, Pitch)
		}
	case packetid.ClientboundPlayerInfoRemove:
		r := bytes.NewReader(p.Data)
		var (
			Action, NumOfPlayers packet.VarInt
			UUID                 packet.UUID
			Gamemode, Ping       packet.VarInt
			HasDisplayName       packet.Boolean
			DisplayName          chat.Message
		)
		var np bytes.Buffer
		for _, v := range [...]packet.Field{&Action, &NumOfPlayers} {
			if _, err := v.ReadFrom(r); err != nil {
				return err
			}
			if _, err := v.WriteTo(&np); err != nil {
				return err
			}
		}
		for i := 0; i < int(NumOfPlayers); i++ {
			if _, err := UUID.ReadFrom(r); err != nil {
				return err
			}
			pp, ok := m.ProfileMapping.Load(uuid.UUID(UUID))
			if ok {
				pp := pp.(PlayerProperties)
				UUID = packet.UUID(pp.ID)
			}
			if _, err := UUID.WriteTo(&np); err != nil {
				return err
			}
			switch Action {
			case 0: // 0: add player
				var Name packet.String
				var NumOfProperties packet.VarInt
				if _, err := Name.ReadFrom(r); err != nil {
					return err
				}
				if _, err := NumOfProperties.ReadFrom(r); err != nil {
					return err
				}
				properties := make([]Property, int(NumOfProperties))
				for i := 0; i < int(NumOfProperties); i++ {
					if _, err := properties[i].ReadFrom(r); err != nil {
						return err
					}
				}

				if ok {
					pp := pp.(PlayerProperties)
					Name = packet.String(pp.Name)
					if pp.DisplayName != "" {
						Name = packet.String(pp.DisplayName)
					}
					properties = pp.Properties
				}

				if _, err := Name.WriteTo(&np); err != nil {
					return err
				}
				if _, err := packet.VarInt(len(properties)).WriteTo(&np); err != nil {
					return err
				}
				for _, v := range properties {
					if _, err := v.WriteTo(&np); err != nil {
						return err
					}
				}

				for _, v := range [...]packet.Field{&Gamemode, &Ping} {
					if _, err := v.ReadFrom(r); err != nil {
						return err
					}
					if _, err := v.WriteTo(&np); err != nil {
						return err
					}
				}

				if _, err := HasDisplayName.ReadFrom(r); err != nil {
					return err
				}
				if HasDisplayName {
					if _, err := DisplayName.ReadFrom(r); err != nil {
						return err
					}
				}
				if ok {
					name := pp.(PlayerProperties).DisplayName
					HasDisplayName = name != ""
					DisplayName = chat.Message{Text: name}
				}

				if _, err := HasDisplayName.WriteTo(&np); err != nil {
					return err
				}
				if HasDisplayName {
					if _, err := DisplayName.WriteTo(&np); err != nil {
						return err
					}
				}

			case 1: // 1: update gamemode
				if _, err := Gamemode.ReadFrom(r); err != nil {
					return err
				}
				if _, err := Gamemode.WriteTo(&np); err != nil {
					return err
				}
			case 2: // 2: update latency
				if _, err := Ping.ReadFrom(r); err != nil {
					return err
				}
				if _, err := Ping.WriteTo(&np); err != nil {
					return err
				}
			case 3: // 3: update display name
				if _, err := HasDisplayName.ReadFrom(r); err != nil {
					return err
				}
				if HasDisplayName {
					if _, err := DisplayName.ReadFrom(r); err != nil {
						return err
					}
				}
				if ok {
					name := pp.(PlayerProperties).DisplayName
					HasDisplayName = name != ""
					DisplayName = chat.Message{Text: name}
				}
				if _, err := HasDisplayName.WriteTo(&np); err != nil {
					return err
				}
				if HasDisplayName {
					if _, err := DisplayName.WriteTo(&np); err != nil {
						return err
					}
				}
			case 4: // 4: remove player
				// No field
			default:
				return fmt.Errorf("unknown action: %d", Action)
			}
		}
		*p = packet.Packet{ID: p.ID, Data: np.Bytes()}
	}
	return nil
}

func (m *Modifier) ModifyServerboundPacket(p *packet.Packet) error {
	switch packetid.ServerboundPacketID(p.ID) {
	case packetid.ServerboundTeleportToEntity: // Spectate
		var TargetPlayer packet.UUID
		if err := p.Scan(&TargetPlayer); err != nil {
			return err
		}
		if pp, ok := m.ProfileMapping.Load(uuid.UUID(TargetPlayer)); ok {
			pp := pp.(PlayerProperties)
			*p = packet.Marshal(p.ID, packet.UUID(pp.ID))
		}
	}
	return nil
}
