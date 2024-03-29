package pernod

import (
	"errors"
	"log"

	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/net/packet"
	"github.com/Tnze/go-mc/server"
	"github.com/google/uuid"
)

// Proxy is an implementation of github.com/Tnze/go-mc/server.GamePlay
type Proxy struct {
	PlayerList              *server.PlayerList
	Destination             string
	ModifyClientboundPacket func(p *packet.Packet) error
	ModifyServerboundPacket func(p *packet.Packet) error
}

type Player struct {
	conn *net.Conn
	c    *bot.Client
}

func (p Player) SendDisconnect(reason chat.Message) {
	p.conn.WritePacket(packet.Marshal(packetid.ClientboundDisconnect, reason))
}

func (p *Proxy) AcceptPlayer(name string, id uuid.UUID, _ int32, conn *net.Conn) {
	c := bot.NewClient()
	c.Auth.Name = name
	c.Auth.UUID = id.String()
	// forward all packet from server to player
	c.Events.AddGeneric(bot.PacketHandler{
		Priority: 100,
		F: func(pk packet.Packet) error {
			if p.ModifyClientboundPacket != nil {
				err := p.ModifyClientboundPacket(&pk)
				if err != nil {
					return err
				}
			}
			return conn.WritePacket(pk)
		},
	})
	if err := c.JoinServer(p.Destination); err != nil {
		log.Printf("Connecting to server[%s] error: %v", p.Destination, err)
		var disconnectErr *bot.DisconnectErr
		if errors.As(err, &disconnectErr) {
			_ = conn.WritePacket(packet.Marshal(
				packetid.ClientboundDisconnect,
				(*chat.Message)(disconnectErr),
			))
		}
		return
	}
	defer c.Close()

	player := Player{conn, c}
	defer player.conn.Close()
	p.PlayerList.ClientJoin(player, server.PlayerSample{
		Name: c.Name,
		ID:   c.UUID,
	})
	defer p.PlayerList.ClientLeft(player)

	go func() {
		// forward all packet from player to server
		var pk packet.Packet
		var err error
		for {
			err = conn.ReadPacket(&pk)
			if err != nil {
				break
			}
			if p.ModifyServerboundPacket != nil {
				err = p.ModifyServerboundPacket(&pk)
				if err != nil {
					break
				}
			}
			err = c.Conn.WritePacket(pk)
			if err != nil {
				break
			}
		}
		log.Printf("Forward packets from player[%s] to server[%s] error: %v", name, p.Destination, err)
	}()
	if err := c.HandleGame(); err != nil {
		log.Printf("Forward packets from server[%s] to player[%s] error: %v", p.Destination, name, err)
	}
}
