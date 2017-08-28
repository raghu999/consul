package config

import (
	"testing"

	"github.com/pascaldekloe/goe/verify"
)

func TestMerge(t *testing.T) {
	tests := []struct {
		desc  string
		files []ConfigFile
		want  ConfigFile
	}{
		{
			"top level fields",
			[]ConfigFile{
				{AdvertiseAddrLAN: pString("a")},
				{AdvertiseAddrLAN: pString("b")},
				{RaftProtocol: pInt(1)},
				{RaftProtocol: pInt(2)},
				{ServerMode: pBool(false)},
				{ServerMode: pBool(true)},
				{JoinAddrsLAN: []string{"a"}},
				{JoinAddrsLAN: []string{"b"}},
				{NodeMeta: map[string]string{"a": "b"}},
				{NodeMeta: map[string]string{"c": "d"}},
				{Ports: Ports{DNS: pInt(1)}},
				{Ports: Ports{DNS: pInt(2), HTTP: pInt(3)}},
			},
			ConfigFile{
				AdvertiseAddrLAN: pString("b"),
				RaftProtocol:     pInt(2),
				ServerMode:       pBool(true),
				JoinAddrsLAN:     []string{"a", "b"},
				NodeMeta:         map[string]string{"c": "d"},
				Ports:            Ports{DNS: pInt(2), HTTP: pInt(3)},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			got, want := Merge(tt.files), tt.want
			if !verify.Values(t, "", got, want) {
				t.FailNow()
			}
		})
	}
}
