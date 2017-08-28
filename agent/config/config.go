package config

import (
	"flag"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/hashicorp/hcl"
)

// ConfigFile defines the format of a config file.
//
// All fields are specified as pointers to simplify merging multiple
// File structures since this allows to determine whether a field has
// been set.
type ConfigFile struct {
	AdvertiseAddrLAN        *string
	AdvertiseAddrWAN        *string
	BindAddr                *string `json:"bind_addr" hcl:"bind_addr"`
	Bootstrap               *bool
	BootstrapExpect         *int
	CheckUpdateInterval     *string `json:"check_update_interval" hcl:"check_update_interval"`
	ClientAddr              *string
	DNSDomain               *string
	DNSRecursors            []string
	DataDir                 *string
	Datacenter              *string
	DevMode                 *bool
	DisableHostNodeID       *bool
	DisableKeyringFile      *bool
	EnableScriptChecks      *bool
	EnableSyslog            *bool
	EnableUI                *bool
	EncryptKey              *string
	JoinAddrsLAN            []string `json:"start_join" hcl:"start_join"`
	JoinAddrsWAN            []string
	LogLevel                *string           `json:"log_level" hcl:"log_level"`
	NodeID                  *string           `json:"node_id" hcl:"node_id"`
	NodeMeta                map[string]string `json:"node_meta" hcl:"node_meta"`
	NodeName                *string           `json:"node_name" hcl:"node_name"`
	NonVotingServer         *bool
	PidFile                 *string
	Ports                   Ports
	RPCProtocol             *int
	RaftProtocol            *int
	RejoinAfterLeave        *bool
	RetryJoinIntervalLAN    *time.Duration
	RetryJoinIntervalWAN    *time.Duration
	RetryJoinLAN            []string
	RetryJoinMaxAttemptsLAN *int
	RetryJoinMaxAttemptsWAN *int
	RetryJoinWAN            []string
	SerfBindAddrLAN         *string
	SerfBindAddrWAN         *string
	ServerMode              *bool
	UIDir                   *string

	DeprecatedRetryJoinAzure RetryJoinAzure
	DeprecatedRetryJoinEC2   RetryJoinEC2
	DeprecatedRetryJoinGCE   RetryJoinGCE
}

type Ports struct {
	DNS     *int
	HTTP    *int
	HTTPS   *int
	SerfLAN *int `json:"serf_lan" hcl:"serf_lan"`
	SerfWAN *int `json:"serf_wan" hcl:"serf_wan"`
	Server  *int

	DeprecatedRPC *int `json:"rpc" hcl:"rpc"`
}

type RetryJoinAzure struct {
	TagName         *string `json:"tag_name" hcl:"tag_name"`
	TagValue        *string `json:"tag_value" hcl:"tag_value"`
	SubscriptionID  *string `json:"subscription_id" hcl:"subscription_id"`
	TenantID        *string `json:"tenant_id" hcl:"tenant_id"`
	ClientID        *string `json:"client_id" hcl:"client_id"`
	SecretAccessKey *string `json:"secret_access_key", hcl:"secret_access_key"`
}

type RetryJoinEC2 struct {
	Region          *string `json:"region" hcl:"region"`
	TagKey          *string `json:"tag_key" hcl:"tag_key"`
	TagValue        *string `json:"tag_value" hcl:"tag_value"`
	AccessKeyID     *string `json:"access_key_id" hcl:"access_key_id"`
	SecretAccessKey *string `json:"secret_access_key" hcl:"secret_access_key"`
}

type RetryJoinGCE struct {
	ProjectName     *string `json:"project_name" hcl:"project_name"`
	ZonePattern     *string `json:"zone_pattern" hcl:"zone_pattern"`
	TagValue        *string `json:"tag_value" hcl:"tag_value"`
	CredentialsFile *string `json:"credentials_file" hcl:"credentials_file"`
}

// ParseFile decodes a configuration file in JSON or HCL format.
func ParseFile(s string) (ConfigFile, error) {
	var f ConfigFile
	if err := hcl.Decode(&f, s); err != nil {
		return ConfigFile{}, err
	}
	return f, nil
}

// Flags defines the command line flags.
//
// All fields are specified as pointers to simplify merging multiple
// File structures since this allows to determine whether a field has
// been set.
type Flags struct {
	File        ConfigFile
	ConfigFiles []string

	DeprecatedDatacenter          *string
	DeprecatedAtlasInfrastructure *string
	DeprecatedAtlasJoin           *bool
	DeprecatedAtlasToken          *string
	DeprecatedAtlasEndpoint       *string
}

// ParseFlag parses the arguments into a Flags struct.
func ParseFlags(args []string) (Flags, error) {
	var f Flags
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	AddFlags(fs, &f)
	if err := fs.Parse(args); err != nil {
		return Flags{}, err
	}
	return f, nil
}

// NewFlagSet creates the set of command line flags for the agent.
func AddFlags(fs *flag.FlagSet, f *Flags) {
	add := func(p interface{}, name, help string) {
		switch x := p.(type) {
		case **bool:
			fs.Var(newBoolPtrValue(x), name, help)
		case **time.Duration:
			fs.Var(newDurationPtrValue(x), name, help)
		case **int:
			fs.Var(newIntPtrValue(x), name, help)
		case **string:
			fs.Var(newStringPtrValue(x), name, help)
		case *[]string:
			fs.Var(newStringSliceValue(x), name, help)
		case *map[string]string:
			fs.Var(newStringMapValue(x), name, help)
		default:
			panic(fmt.Sprintf("invalid type: %T", p))
		}
	}

	// command line flags ordered by flag name
	add(&f.File.AdvertiseAddrLAN, "advertise", "Sets the advertise address to use.")
	add(&f.File.AdvertiseAddrWAN, "advertise-wan", "Sets address to advertise on WAN instead of -advertise address.")
	add(&f.File.BindAddr, "bind", "Sets the bind address for cluster communication.")
	add(&f.File.Bootstrap, "bootstrap", "Sets server to bootstrap mode.")
	add(&f.File.BootstrapExpect, "bootstrap-expect", "Sets server to expect bootstrap mode.")
	add(&f.File.ClientAddr, "client", "Sets the address to bind for client access. This includes RPC, DNS, HTTP and HTTPS (if configured).")
	add(&f.ConfigFiles, "config-dir", "Path to a directory to read configuration files from. This will read every file ending in '.json' as configuration in this directory in alphabetical order. Can be specified multiple times.")
	add(&f.ConfigFiles, "config-file", "Path to a JSON file to read configuration from. Can be specified multiple times.")
	add(&f.File.DataDir, "data-dir", "Path to a data directory to store agent state.")
	add(&f.File.Datacenter, "datacenter", "Datacenter of the agent.")
	add(&f.File.DevMode, "dev", "Starts the agent in development mode.")
	add(&f.File.DisableHostNodeID, "disable-host-node-id", "Setting this to true will prevent Consul from using information from the host to generate a node ID, and will cause Consul to generate a random node ID instead.")
	add(&f.File.DisableKeyringFile, "disable-keyring-file", "Disables the backing up of the keyring to a file.")
	add(&f.File.Ports.DNS, "dns-port", "DNS port to use.")
	add(&f.File.DNSDomain, "domain", "Domain to use for DNS interface.")
	add(&f.File.EnableScriptChecks, "enable-script-checks", "Enables health check scripts.")
	add(&f.File.EncryptKey, "encrypt", "Provides the gossip encryption key.")
	add(&f.File.Ports.HTTP, "http-port", "Sets the HTTP API port to listen on.")
	add(&f.File.JoinAddrsLAN, "join", "Address of an agent to join at start time. Can be specified multiple times.")
	add(&f.File.JoinAddrsWAN, "join-wan", "Address of an agent to join -wan at start time. Can be specified multiple times.")
	add(&f.File.LogLevel, "log-level", "Log level of the agent.")
	add(&f.File.NodeName, "node", "Name of this node. Must be unique in the cluster.")
	add(&f.File.NodeID, "node-id", "A unique ID for this node across space and time. Defaults to a randomly-generated ID that persists in the data-dir.")
	add(&f.File.NodeMeta, "node-meta", "An arbitrary metadata key/value pair for this node, of the format `key:value`. Can be specified multiple times.")
	add(&f.File.NonVotingServer, "non-voting-server", "(Enterprise-only) This flag is used to make the server not participate in the Raft quorum, and have it only receive the data replication stream. This can be used to add read scalability to a cluster in cases where a high volume of reads to servers are needed.")
	add(&f.File.PidFile, "pid-file", "Path to file to store agent PID.")
	add(&f.File.RPCProtocol, "protocol", "Sets the protocol version. Defaults to latest.")
	add(&f.File.RaftProtocol, "raft-protocol", "Sets the Raft protocol version. Defaults to latest.")
	add(&f.File.DNSRecursors, "recursor", "Address of an upstream DNS server. Can be specified multiple times.")
	add(&f.File.RejoinAfterLeave, "rejoin", "Ignores a previous leave and attempts to rejoin the cluster.")
	add(&f.File.RetryJoinIntervalLAN, "retry-interval", "Time to wait between join attempts.")
	add(&f.File.RetryJoinIntervalWAN, "retry-interval-wan", "Time to wait between join -wan attempts.")
	add(&f.File.RetryJoinLAN, "retry-join", "Address of an agent to join at start time with retries enabled. Can be specified multiple times.")
	add(&f.File.RetryJoinWAN, "retry-join-wan", "Address of an agent to join -wan at start time with retries enabled. Can be specified multiple times.")
	add(&f.File.RetryJoinMaxAttemptsLAN, "retry-max", "Maximum number of join attempts. Defaults to 0, which will retry indefinitely.")
	add(&f.File.RetryJoinMaxAttemptsWAN, "retry-max-wan", "Maximum number of join -wan attempts. Defaults to 0, which will retry indefinitely.")
	add(&f.File.SerfBindAddrLAN, "serf-lan-bind", "Address to bind Serf LAN listeners to.")
	add(&f.File.SerfBindAddrWAN, "serf-wan-bind", "Address to bind Serf WAN listeners to.")
	add(&f.File.ServerMode, "server", "Switches agent to server mode.")
	add(&f.File.EnableSyslog, "syslog", "Enables logging to syslog.")
	add(&f.File.EnableUI, "ui", "Enables the built-in static web UI server.")
	add(&f.File.UIDir, "ui-dir", "Path to directory containing the web UI resources.")

	// deprecated flags orderd by flag name
	add(&f.DeprecatedAtlasInfrastructure, "atlas", "(deprecated) Sets the Atlas infrastructure name, enables SCADA.")
	add(&f.DeprecatedAtlasEndpoint, "atlas-endpoint", "(deprecated) The address of the endpoint for Atlas integration.")
	add(&f.DeprecatedAtlasJoin, "atlas-join", "(deprecated) Enables auto-joining the Atlas cluster.")
	add(&f.DeprecatedAtlasToken, "atlas-token", "(deprecated) Provides the Atlas API token.")
	add(&f.DeprecatedDatacenter, "dc", "(deprecated) Datacenter of the agent (use 'datacenter' instead).")
	add(&f.File.DeprecatedRetryJoinAzure.TagName, "retry-join-azure-tag-name", "Azure tag name to filter on for server discovery.")
	add(&f.File.DeprecatedRetryJoinAzure.TagValue, "retry-join-azure-tag-value", "Azure tag value to filter on for server discovery.")
	add(&f.File.DeprecatedRetryJoinEC2.Region, "retry-join-ec2-region", "EC2 Region to discover servers in.")
	add(&f.File.DeprecatedRetryJoinEC2.TagKey, "retry-join-ec2-tag-key", "EC2 tag key to filter on for server discovery.")
	add(&f.File.DeprecatedRetryJoinEC2.TagValue, "retry-join-ec2-tag-value", "EC2 tag value to filter on for server discovery.")
	add(&f.File.DeprecatedRetryJoinGCE.CredentialsFile, "retry-join-gce-credentials-file", "Path to credentials JSON file to use with Google Compute Engine.")
	add(&f.File.DeprecatedRetryJoinGCE.ProjectName, "retry-join-gce-project-name", "Google Compute Engine project to discover servers in.")
	add(&f.File.DeprecatedRetryJoinGCE.TagValue, "retry-join-gce-tag-value", "Google Compute Engine tag value to filter on for server discovery.")
	add(&f.File.DeprecatedRetryJoinGCE.ZonePattern, "retry-join-gce-zone-pattern", "Google Compute Engine region or zone to discover servers in (regex pattern).")
}

// Config is the runtime configuration.
type Config struct {
	// simple values

	Bootstrap           bool
	CheckUpdateInterval time.Duration
	Datacenter          string

	// address values

	BindAddrs    []string
	JoinAddrsLAN []string

	// server endpoint values

	DNSPort     int
	DNSAddrsTCP []string
	DNSAddrsUDP []string

	// other values

	NodeMeta map[string]string
}

// NewConfig creates the runtime configuration from a configuration
// file. It performs all the necessary syntactic and semantic validation
// so that the resulting runtime configuration is usable.
func NewConfig(f ConfigFile) (c Config, err error) {
	boolVal := func(b *bool) bool {
		if err != nil || b == nil {
			return false
		}
		return *b
	}

	durationVal := func(s *string) (d time.Duration) {
		if err != nil || s == nil {
			return 0
		}
		d, err = time.ParseDuration(*s)
		return
	}

	intVal := func(n *int) int {
		if err != nil || n == nil {
			return 0
		}
		return *n
	}

	stringVal := func(s *string) string {
		if err != nil || s == nil {
			return ""
		}
		return *s
	}

	addrVal := func(s *string) string {
		addr := stringVal(s)
		if addr == "" {
			return "0.0.0.0"
		}
		return addr
	}

	joinHostPort := func(host string, port int) string {
		if host == "0.0.0.0" {
			host = ""
		}
		return net.JoinHostPort(host, strconv.Itoa(port))
	}

	c.Bootstrap = boolVal(f.Bootstrap)
	c.CheckUpdateInterval = durationVal(f.CheckUpdateInterval)
	c.Datacenter = stringVal(f.Datacenter)
	c.JoinAddrsLAN = f.JoinAddrsLAN
	c.NodeMeta = f.NodeMeta

	// if no bind address is given but ports are specified then we bail.
	// this only affects tests since in prod this gets merged with the
	// default config which always has a bind address.
	if f.BindAddr == nil && !reflect.DeepEqual(f.Ports, Ports{}) {
		return Config{}, fmt.Errorf("no bind address specified")
	}

	if f.BindAddr != nil {
		c.BindAddrs = []string{addrVal(f.BindAddr)}
	}

	if f.Ports.DNS != nil {
		c.DNSPort = intVal(f.Ports.DNS)
		for _, addr := range c.BindAddrs {
			c.DNSAddrsTCP = append(c.DNSAddrsTCP, joinHostPort(addr, c.DNSPort))
			c.DNSAddrsUDP = append(c.DNSAddrsUDP, joinHostPort(addr, c.DNSPort))
		}
	}

	return
}
