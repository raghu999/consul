package config

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/pascaldekloe/goe/verify"
)

// TestParseFlags tests whether command line flags are properly parsed
// into the Flags/File structure. It does not test the conversion into
// the final runtime configuration. See TestConfig for that.
func TestParseFlags(t *testing.T) {
	tests := []struct {
		args  []string
		flags Flags
		err   error
	}{
		{},
		{
			args:  []string{`-bind`, `a`},
			flags: Flags{File: ConfigFile{BindAddr: pString("a")}},
		},
		{
			args:  []string{`-bootstrap`},
			flags: Flags{File: ConfigFile{Bootstrap: pBool(true)}},
		},
		{
			args:  []string{`-bootstrap=true`},
			flags: Flags{File: ConfigFile{Bootstrap: pBool(true)}},
		},
		{
			args:  []string{`-bootstrap=false`},
			flags: Flags{File: ConfigFile{Bootstrap: pBool(false)}},
		},
		{
			args:  []string{`-bootstrap`, `true`},
			flags: Flags{File: ConfigFile{Bootstrap: pBool(true)}},
		},
		{
			args:  []string{`-config-file`, `a`, `-config-dir`, `b`, `-config-file`, `c`, `-config-dir`, `d`},
			flags: Flags{ConfigFiles: []string{"a", "b", "c", "d"}},
		},
		{
			args:  []string{`-datacenter`, `a`},
			flags: Flags{File: ConfigFile{Datacenter: pString("a")}},
		},
		{
			args:  []string{`-dns-port`, `1`},
			flags: Flags{File: ConfigFile{Ports: Ports{DNS: pInt(1)}}},
		},
		{
			args:  []string{`-join`, `a`, `-join`, `b`},
			flags: Flags{File: ConfigFile{JoinAddrsLAN: []string{"a", "b"}}},
		},
		{
			args:  []string{`-node-meta`, `a:b`, `-node-meta`, `c:d`},
			flags: Flags{File: ConfigFile{NodeMeta: map[string]string{"a": "b", "c": "d"}}},
		},
	}

	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			flags, err := ParseFlags(tt.args)
			if got, want := err, tt.err; !reflect.DeepEqual(got, want) {
				t.Fatalf("got error %v want %v", got, want)
			}
			if !verify.Values(t, "flag", flags, tt.flags) {
				t.FailNow()
			}
		})
	}
}

// TestConfig tests whether a combination of command line flags and
// config files creates the correct runtime configuration. The tests do
// not use the default configuration as basis as this would provide a
// lot of redundancy in the test results.
//
// The tests are grouped and within the groups are ordered alphabetically.
func TestConfig(t *testing.T) {
	tests := []struct {
		desc  string
		fmt   string // json or hcl
		def   ConfigFile
		files []string
		flags []string
		cfg   Config
		err   error
	}{
		{
			desc: "default config",
			def:  defaultFile,
			cfg:  defaultConfig,
		},

		// cmd line flags
		{
			flags: []string{`-bind`, `1.2.3.4`},
			cfg:   Config{BindAddrs: []string{"1.2.3.4"}},
		},
		{
			flags: []string{`-bootstrap`},
			cfg:   Config{Bootstrap: true},
		},
		{
			flags: []string{`-datacenter`, `a`},
			cfg:   Config{Datacenter: "a"},
		},
		{
			flags: []string{`-dns-port`, `123`, `-bind`, `0.0.0.0`},
			cfg: Config{
				BindAddrs:   []string{"0.0.0.0"},
				DNSPort:     123,
				DNSAddrsUDP: []string{":123"},
				DNSAddrsTCP: []string{":123"},
			},
		},
		{
			flags: []string{`-join`, `a`, `-join`, `b`},
			cfg:   Config{JoinAddrsLAN: []string{"a", "b"}},
		},
		{
			flags: []string{`-node-meta`, `a:b`, `-node-meta`, `c:d`},
			cfg:   Config{NodeMeta: map[string]string{"a": "b", "c": "d"}},
		},

		// json cfg file
		{
			fmt:   "json",
			files: []string{`{"bootstrap":true}`},
			cfg:   Config{Bootstrap: true},
		},
		{
			fmt:   "json",
			files: []string{`{"check_update_interval":"5m"}`},
			cfg:   Config{CheckUpdateInterval: 5 * time.Minute},
		},
		{
			fmt:   "json",
			files: []string{`{"datacenter":"a"}`},
			cfg:   Config{Datacenter: "a"},
		},
		{
			fmt:   "json",
			files: []string{`{"bind_addr":"0.0.0.0","ports":{"dns":123}}`},
			cfg: Config{
				BindAddrs:   []string{"0.0.0.0"},
				DNSPort:     123,
				DNSAddrsUDP: []string{":123"},
				DNSAddrsTCP: []string{":123"},
			},
		},
		{
			fmt:   "json",
			files: []string{`{"start_join":["a"]}`, `{"start_join":["b"]}`},
			cfg:   Config{JoinAddrsLAN: []string{"a", "b"}},
		},
		{
			fmt:   "json",
			files: []string{`{"node_meta":{"a":"b"}}`},
			cfg:   Config{NodeMeta: map[string]string{"a": "b"}},
		},
		{
			fmt:   "json",
			files: []string{`{"node_meta":{"a":"b"}}`, `{"node_meta":{"c":"d"}}`},
			cfg:   Config{NodeMeta: map[string]string{"c": "d"}},
		},

		// hcl cfg file
		{
			fmt:   "hcl",
			files: []string{`bootstrap = true`},
			cfg:   Config{Bootstrap: true},
		},
		{
			fmt:   "hcl",
			files: []string{`check_update_interval = "5m"`},
			cfg:   Config{CheckUpdateInterval: 5 * time.Minute},
		},
		{
			fmt:   "hcl",
			files: []string{`datacenter = "a"`},
			cfg:   Config{Datacenter: "a"},
		},
		{
			fmt: "hcl",
			files: []string{`
				bind_addr = "0.0.0.0"
				ports { dns = 123 }`},
			cfg: Config{
				BindAddrs:   []string{"0.0.0.0"},
				DNSPort:     123,
				DNSAddrsUDP: []string{":123"},
				DNSAddrsTCP: []string{":123"},
			},
		},
		{
			fmt:   "hcl",
			files: []string{`start_join = ["a"]`, `start_join = ["b"]`},
			cfg:   Config{JoinAddrsLAN: []string{"a", "b"}},
		},
		{
			fmt:   "hcl",
			files: []string{`node_meta { a = "b" }`},
			cfg:   Config{NodeMeta: map[string]string{"a": "b"}},
		},
		{
			fmt:   "hcl",
			files: []string{`node_meta { a = "b" }`, `node_meta { c = "d" }`},
			cfg:   Config{NodeMeta: map[string]string{"c": "d"}},
		},

		// precedence rules
		{
			fmt:   "json",
			files: []string{`{"bootstrap":true}`, `{"bootstrap":false}`},
			cfg:   Config{Bootstrap: false},
		},
		{
			fmt:   "json",
			files: []string{`{"bootstrap":true}`},
			flags: []string{`-bootstrap=false`},
			cfg:   Config{Bootstrap: false},
		},
		{
			fmt:   "hcl",
			files: []string{`bootstrap=true`, `bootstrap=false`},
			cfg:   Config{Bootstrap: false},
		},
		{
			fmt:   "hcl",
			files: []string{`bootstrap=true`},
			flags: []string{`-bootstrap=false`},
			cfg:   Config{Bootstrap: false},
		},
	}

	for _, tt := range tests {
		var desc []string
		if tt.desc != "" {
			desc = append(desc, tt.desc)
		}
		if len(tt.files) > 0 {
			s := tt.fmt + ":" + strings.Join(tt.files, ",")
			desc = append(desc, s)
		}
		if len(tt.flags) > 0 {
			s := "flags:" + strings.Join(tt.flags, " ")
			desc = append(desc, s)
		}

		t.Run(strings.Join(desc, ";"), func(t *testing.T) {
			// start with default config
			files := []ConfigFile{tt.def}

			// add files in order
			for _, s := range tt.files {
				f, err := ParseFile(s)
				if err != nil {
					t.Fatalf("ParseFile failed for %q: %s", s, err)
				}
				files = append(files, f)
			}

			// add flags
			flags, err := ParseFlags(tt.flags)
			if err != nil {
				t.Fatalf("ParseFlags failed: %s", err)
			}
			files = append(files, flags.File)

			// merge files and build config
			cfg, err := NewConfig(Merge(files))
			if err != nil {
				t.Fatalf("NewConfig failed: %s", err)
			}

			// fmt.Printf("cfg: %#v\n", cfg)

			if !verify.Values(t, "", cfg, tt.cfg) {
				t.FailNow()
			}
		})
	}
}
