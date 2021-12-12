package main

import (
	"github.com/Tnze/go-mc/net"
	"github.com/Tnze/go-mc/offline"
	"github.com/google/uuid"
	"log"
	"pernod"
	"pernod/playermodify"
)

type ModifyProxy struct {
	pernod.Proxy
	*playermodify.Modifier
}

func (m *ModifyProxy) AcceptPlayer(name string, id uuid.UUID, protocol int32, conn *net.Conn) {
	if offline.NameToUUID(name) != id {
		pp, err := playermodify.GetPlayerProperties(id)
		if err != nil {
			log.Printf("Get player properties error: %v", err)
		}
		offlineUUID := offline.NameToUUID(name)
		m.Modifier.ProfileMapping.Store(offlineUUID, pp)
		defer m.ProfileMapping.Delete(offlineUUID)
	}

	m.Proxy.AcceptPlayer(name, id, protocol, conn)
}
