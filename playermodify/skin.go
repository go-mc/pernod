package playermodify

import (
	"encoding/json"
	"fmt"
	pk "github.com/Tnze/go-mc/net/packet"
	"github.com/google/uuid"
	"io"
	"net/http"
)

type PlayerProperties struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	DisplayName string     `json:"-"`
	Properties  []Property `json:"properties"`
}

type Property struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

func GetPlayerProperties(UUID uuid.UUID) (result PlayerProperties, err error) {
	resp, err := http.Get("https://sessionserver.mojang.com/session/minecraft/profile/" + UUID.String())
	if err != nil {
		return PlayerProperties{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		err = json.NewDecoder(resp.Body).Decode(&result)
		return
	} else if resp.StatusCode == http.StatusNoContent {
		return
	}
	return PlayerProperties{}, fmt.Errorf("http status error: %s", resp.Status)
}

func (pp *Property) WriteTo(w io.Writer) (n int64, err error) {
	has := pk.Boolean(pp.Signature != "")
	for _, v := range []pk.FieldEncoder{
		pk.String(pp.Name),
		pk.String(pp.Value),
		has, pk.Opt{
			Has:   &has,
			Field: pk.String(pp.Signature),
		},
	} {
		nn, err := v.WriteTo(w)
		n += nn
		if err != nil {
			return n, err
		}
	}
	return
}

func (pp *Property) ReadFrom(r io.Reader) (n int64, err error) {
	has := pk.Boolean(pp.Signature != "")
	for _, v := range []pk.FieldDecoder{
		(*pk.String)(&pp.Name),
		(*pk.String)(&pp.Value),
		&has, pk.Opt{
			Has:   &has,
			Field: (*pk.String)(&pp.Signature),
		},
	} {
		nn, err := v.ReadFrom(r)
		n += nn
		if err != nil {
			return n, err
		}
	}
	return
}
