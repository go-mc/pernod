package main

import (
	"flag"
	"log"
	"pernod"
	"pernod/playermodify"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/server"
	"github.com/google/uuid"
)

const ServerName = "Pernod"

func main() {
	flag.Parse()
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatalf("Parse config file error: %v", err)
	}

	motd := chat.Text(config.Description)
	playerList := server.NewPlayerList(ServerName, server.ProtocolVersion, config.MaxPlayersNum, &motd)

	var modifier playermodify.Modifier
	// create proxy objects
	destinations := make(map[string]*ModifyProxy, len(config.Destinations))
	for name, addr := range config.Destinations {
		destinations[name] = &ModifyProxy{
			Proxy: pernod.Proxy{
				PlayerList:              playerList,
				Destination:             addr.Address,
				ModifyServerboundPacket: modifier.ModifyServerboundPacket,
				ModifyClientboundPacket: modifier.ModifyClientboundPacket,
			},
			Modifier: &modifier,
		}
	}

	// start listeners
	var wg sync.WaitGroup
	wg.Add(len(config.Listeners))
	for _, listenCfg := range config.Listeners {
		s := server.Server{
			ListPingHandler: playerList,
			LoginHandler: &server.MojangLoginHandler{
				OnlineMode: listenCfg.OnlineMode,
				Threshold:  listenCfg.Threshold,
			},
			GamePlay: destinations[listenCfg.Destination],
		}
		go func(s server.Server, addr string) {
			if err := s.Listen(addr); err != nil {
				log.Printf("Listene at %s error: %v, goruntine exit", addr, err)
				return
			}
			wg.Done()
		}(s, listenCfg.ListenAt)
	}
	wg.Wait()
	log.Printf("All listener returned, program exit")
}

var configFile = flag.String("c", "config.toml", "config file name")

var config struct {
	MaxPlayersNum int
	Description   string
	Listeners     []struct {
		ListenAt    string
		Destination string
		OnlineMode  bool
		Threshold   int
	}
	Destinations map[string]struct {
		Address string
	}
	ProfileMappings []struct {
		Match struct {
			Name string
			UUID uuid.UUID
		}
		DisplayName string
		MapTo       string
		Skin        string
	}
}
