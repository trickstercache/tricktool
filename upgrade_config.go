package main

import (
	"fmt"
	"io/ioutil"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func upgradeConfig(cmd *cobra.Command, args []string) {

	// load the toml file from disk
	b, err := ioutil.ReadFile(args[0])
	if err != nil {
		cmd.PrintErr("unable to open source file: ", err, "\n")
		return
	}

	// verify it is valid toml
	c := &config{}
	_, err = toml.Decode(string(b), c)
	if err != nil {
		cmd.PrintErr("unable to parse source file: ", err, "\n")
		return
	}

	c.makeConversions()

	b, err = yaml.Marshal(c)
	if err != nil {
		cmd.PrintErr("unable to create destination file: ", err, "\n")
		cmd.PrintErr(err)
		return
	}

	fmt.Println(string(b))

}

func (c *config) makeConversions() {
	for _, cache := range c.Caches {
		cache.Provider = cache.CacheType
		if cache.Index == nil {
			continue
		}
		cache.Index.FlushIntervalMS = cache.Index.FlushIntervalSecs * 1000
		cache.Index.ReapIntervalMS = cache.Index.ReapIntervalSecs * 1000
	}

	for _, nc := range c.NegativeCaches {
		for k, v := range nc {
			nc[k] = v * 1000
		}
	}

	for _, b := range c.Backends {
		b.Provider = b.OriginType
		b.TimeoutMS = b.TimeoutSecs * 1000
		b.KeepAliveTimeoutMS = b.KeepAliveTimeoutSecs * 1000
		b.MaxTTLMS = b.MaxTTLSecs * 1000
		b.BackfillToleranceMS = b.BackfillToleranceSecs * 1000
		b.TimeseriesTTLMS = b.TimeseriesTTLSecs * 1000
		b.FastForwardTTLMS = b.FastForwardTTLSecs * 1000

		if b.HealthCheckPath != "" || b.HealthCheckQuery != "" || b.HealthCheckVerb != "" {
			b.HealthCheck = &hc{
				Verb:    b.HealthCheckVerb,
				Query:   b.HealthCheckQuery,
				Path:    b.HealthCheckPath,
				Headers: b.HealthCheckHeaders,
			}
		}
	}

	for _, t := range c.TracingConfigs {
		t.Provider = t.TracerType
	}

}

type config struct {
	Main             mainConfig                `toml:"main,omitempty" yaml:"main,omitempty"`
	Frontend         *frontend                 `toml:"frontend,omitempty" yaml:"frontend,omitempty"`
	ReloadConfig     *reloading                `toml:"reloading,omitempty" yaml:"reloading,omitempty"`
	Backends         map[string]*backend       `toml:"origins,omitempty" yaml:"backends,omitempty"`
	Caches           map[string]*cache         `toml:"caches,omitempty" yaml:"caches,omitempty"`
	NegativeCaches   map[string]map[string]int `toml:"negative_caches,omitempty" yaml:"negative_caches,omitempty"`
	Logging          *logging                  `toml:"logging,omitempty" yaml:"logging,omitempty"`
	Metrics          *metrics                  `toml:"metrics,omitempty" yaml:"metrics,omitempty"`
	TracingConfigs   map[string]*tracing       `toml:"tracing,omitempty" yaml:"tracing,omitempty"`
	Rules            map[string]*rule          `toml:"rules,omitempty" yaml:"rules,omitempty"`
	RequestRewriters map[string]*rewriter      `toml:"request_rewriters,omitempty" yaml:"request_rewriters,omitempty"`
}

type mainConfig struct {
	InstanceID        int    `toml:"instance_id,omitempty" yaml:"instance_id,omitempty"`
	ConfigHandlerPath string `toml:"config_handler_path,omitempty" yaml:"config_handler_path,omitempty"`
	PingHandlerPath   string `toml:"ping_handler_path,omitempty" yaml:"ping_handler_path,omitempty"`
	ReloadHandlerPath string `toml:"reload_handler_path,omitempty" yaml:"reload_handler_path,omitempty"`
	HealthHandlerPath string `toml:"health_handler_path,omitempty" yaml:"health_handler_path,omitempty"`
	PprofServer       string `toml:"pprof_server,omitempty" yaml:"pprof_server,omitempty"`
	ServerName        string `toml:"server_name,omitempty" yaml:"server_name,omitempty"`
}

type frontend struct {
	ListenAddress    string `toml:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	ListenPort       int    `toml:"listen_port,omitempty" yaml:"listen_port,omitempty"`
	TLSListenAddress string `toml:"tls_listen_address,omitempty" yaml:"tls_listen_address,omitempty"`
	TLSListenPort    int    `toml:"tls_listen_port,omitempty" yaml:"tls_listen_port,omitempty"`
	ConnectionsLimit int    `toml:"connections_limit,omitempty" yaml:"connections_limit,omitempty"`
}

type cache struct {
	CacheType  string     `toml:"cache_type,omitempty" yaml:"-"`
	Provider   string     `toml:"-" yaml:"provider,omitempty"`
	Index      *index     `toml:"index,omitempty" yaml:"index,omitempty"`
	Redis      redis      `toml:"redis,omitempty" yaml:"redis,omitempty"`
	Filesystem filesystem `toml:"filesystem,omitempty" yaml:"filesystem,omitempty"`
	BBolt      bbolt      `toml:"bbolt,omitempty" yaml:"bbolt,omitempty"`
	Badger     badgerDB   `toml:"badger,omitempty" yaml:"badger,omitempty"`
}

type index struct {
	ReapIntervalSecs  int `toml:"reap_interval_secs,omitempty" yaml:"-"`
	FlushIntervalSecs int `toml:"flush_interval_secs,omitempty" yaml:"-"`
	ReapIntervalMS    int `toml:"-" yaml:"reap_interval_ms,omitempty"`
	FlushIntervalMS   int `toml:"-" yaml:"flush_interval_ms,omitempty"`

	MaxSizeBytes          int64 `toml:"max_size_bytes,omitempty" yaml:"max_size_bytes,omitempty"`
	MaxSizeBackoffBytes   int64 `toml:"max_size_backoff_bytes,omitempty" yaml:"max_size_backoff_bytes,omitempty"`
	MaxSizeObjects        int64 `toml:"max_size_objects,omitempty" yaml:"max_size_objects,omitempty"`
	MaxSizeBackoffObjects int64 `toml:"max_size_backoff_objects,omitempty" yaml:"max_size_backoff_objects,omitempty"`
}

type redis struct {
	ClientType           string   `toml:"client_type,omitempty" yaml:"client_type,omitempty"`
	Protocol             string   `toml:"protocol,omitempty" yaml:"protocol,omitempty"`
	Endpoint             string   `toml:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Endpoints            []string `toml:"endpoints,omitempty" yaml:"endpoints,omitempty"`
	Password             string   `toml:"password,omitempty" yaml:"password,omitempty"`
	SentinelMaster       string   `toml:"sentinel_master,omitempty" yaml:"sentinel_master,omitempty"`
	DB                   int      `toml:"db,omitempty" yaml:"db,omitempty"`
	MaxRetries           int      `toml:"max_retries,omitempty" yaml:"max_retries,omitempty"`
	MinRetryBackoffMS    int      `toml:"min_retry_backoff_ms,omitempty" yaml:"min_retry_backoff_ms,omitempty"`
	MaxRetryBackoffMS    int      `toml:"max_retry_backoff_ms,omitempty" yaml:"max_retry_backoff_ms,omitempty"`
	DialTimeoutMS        int      `toml:"dial_timeout_ms,omitempty" yaml:"dial_timeout_ms,omitempty"`
	ReadTimeoutMS        int      `toml:"read_timeout_ms,omitempty" yaml:"read_timeout_ms,omitempty"`
	WriteTimeoutMS       int      `toml:"write_timeout_ms,omitempty" yaml:"write_timeout_ms,omitempty"`
	PoolSize             int      `toml:"pool_size,omitempty" yaml:"pool_size,omitempty"`
	MinIdleConns         int      `toml:"min_idle_conns,omitempty" yaml:"min_idle_conns,omitempty"`
	MaxConnAgeMS         int      `toml:"max_conn_age_ms,omitempty" yaml:"max_conn_age_ms,omitempty"`
	PoolTimeoutMS        int      `toml:"pool_timeout_ms,omitempty" yaml:"pool_timeout_ms,omitempty"`
	IdleTimeoutMS        int      `toml:"idle_timeout_ms,omitempty" yaml:"idle_timeout_ms,omitempty"`
	IdleCheckFrequencyMS int      `toml:"idle_check_frequency_ms,omitempty" yaml:"idle_check_frequency_ms,omitempty"`
}

type bbolt struct {
	Filename string `toml:"filename,omitempty" yaml:"filename,omitempty"`
	Bucket   string `toml:"bucket,omitempty" yaml:"bucket,omitempty"`
}

type filesystem struct {
	CachePath string `toml:"cache_path,omitempty" yaml:"cache_path,omitempty"`
}

type badgerDB struct {
	Directory      string `toml:"directory,omitempty" yaml:"directory,omitempty"`
	ValueDirectory string `toml:"value_directory,omitempty" yaml:"value_directory,omitempty"`
}

type backend struct {
	Hosts                 []string         `toml:"hosts,omitempty" yaml:"hosts,omitempty"`
	OriginType            string           `toml:"origin_type,omitempty" yaml:"-"`
	Provider              string           `toml:"-" yaml:"provider,omitempty"`
	OriginURL             string           `toml:"origin_url,omitempty" yaml:"origin_url,omitempty"`
	TimeoutSecs           int64            `toml:"timeout_secs,omitempty" yaml:"-"`
	TimeoutMS             int64            `toml:"-" yaml:"timeout_ms,omitempty"`
	KeepAliveTimeoutSecs  int64            `toml:"keep_alive_timeout_secs,omitempty" yaml:"-"`
	KeepAliveTimeoutMS    int64            `toml:"-" yaml:"keep_alive_timeout_ms,omitempty"`
	MaxIdleConns          int              `toml:"max_idle_conns,omitempty" yaml:"max_idle_conns,omitempty"`
	CacheName             string           `toml:"cache_name,omitempty" yaml:"cache_name,omitempty"`
	CacheKeyPrefix        string           `toml:"cache_key_prefix,omitempty" yaml:"cache_key_prefix,omitempty"`
	TSRetentionFactor     int              `toml:"timeseries_retention_factor,omitempty" yaml:"timeseries_retention_factor,omitempty"`
	TSEvictionMethodN     string           `toml:"timeseries_eviction_method,omitempty" yaml:"timeseries_eviction_method,omitempty"`
	BackfillToleranceSecs int64            `toml:"backfill_tolerance_secs,omitempty" yaml:"-"`
	BackfillToleranceMS   int64            `toml:"-" yaml:"backfill_tolerance_ms,omitempty"`
	Paths                 map[string]*path `toml:"paths,omitempty" yaml:"paths,omitempty"`
	NegativeCacheName     string           `toml:"negative_cache_name,omitempty" yaml:"negative_cache_name,omitempty"`
	TimeseriesTTLSecs     int              `toml:"timeseries_ttl_secs,omitempty" yaml:"-"`
	TimeseriesTTLMS       int              `toml:"-" yaml:"timeseries_ttl_ms,omitempty"`
	FastForwardTTLSecs    int              `toml:"fastforward_ttl_secs,omitempty" yaml:"-"`
	FastForwardTTLMS      int              `toml:"-" yaml:"fastforward_ttl_ms,omitempty"`
	MaxTTLSecs            int              `toml:"max_ttl_secs,omitempty" yaml:"-"`
	MaxTTLMS              int              `toml:"-" yaml:"max_ttl_ms,omitempty"`
	RevalidationFactor    float64          `toml:"revalidation_factor,omitempty" yaml:"revalidation_factor,omitempty"`
	MaxObjectSizeBytes    int              `toml:"max_object_size_bytes,omitempty" yaml:"max_object_size_bytes,omitempty"`
	CompressableTypes     []string         `toml:"compressable_types,omitempty" yaml:"compressable_types,omitempty"`
	TracingConfigName     string           `toml:"tracing_name,omitempty" yaml:"tracing_name,omitempty"`
	RuleName              string           `toml:"rule_name,omitempty" yaml:"rule_name,omitempty"`
	ReqRewriterName       string           `toml:"req_rewriter_name,omitempty" yaml:"req_rewriter_name,omitempty"`
	TLS                   tls              `toml:"tls,omitempty" yaml:"tls,omitempty"`
	ForwardedHeaders      string           `toml:"forwarded_headers,omitempty" yaml:"forwarded_headers,omitempty"`
	IsDefault             bool             `toml:"is_default,omitempty" yaml:"is_default,omitempty"`
	FastForwardDisable    bool             `toml:"fast_forward_disable,omitempty" yaml:"fast_forward_disable,omitempty"`
	PathRoutingDisabled   bool             `toml:"path_routing_disabled,omitempty" yaml:"path_routing_disabled,omitempty"`
	RequireTLS            bool             `toml:"require_tls,omitempty" yaml:"require_tls,omitempty"`
	MpartRangesDisabled   bool             `toml:"multipart_ranges_disabled,omitempty" yaml:"multipart_ranges_disabled,omitempty"`
	DearticulateRanges    bool             `toml:"dearticulate_upstream_ranges,omitempty" yaml:"dearticulate_upstream_ranges,omitempty"`

	HealthCheck        *hc               `yaml:"healthcheck,omitempty"`
	HealthCheckPath    string            `toml:"health_check_upstream_path" yaml:"-"`
	HealthCheckVerb    string            `toml:"health_check_verb" yaml:"-"`
	HealthCheckQuery   string            `toml:"health_check_query" yaml:"-"`
	HealthCheckHeaders map[string]string `toml:"health_check_headers" yaml:"-"`
}

type hc struct {
	Verb    string            `yaml:"verb,omitempty"`
	Path    string            `yaml:"path,omitempty"`
	Query   string            `yaml:"query,omitempty"`
	Headers map[string]string `yaml:"headers,omitempty"`
}

type path struct {
	Path               string            `toml:"path,omitempty" yaml:"path,omitempty"`
	MatchTypeName      string            `toml:"match_type,omitempty" yaml:"match_type,omitempty"`
	HandlerName        string            `toml:"handler,omitempty" yaml:"handler,omitempty"`
	Methods            []string          `toml:"methods,omitempty" yaml:"methods,omitempty"`
	CacheKeyParams     []string          `toml:"cache_key_params,omitempty" yaml:"cache_key_params,omitempty"`
	CacheKeyHeaders    []string          `toml:"cache_key_headers,omitempty" yaml:"cache_key_headers,omitempty"`
	CacheKeyFormFields []string          `toml:"cache_key_form_fields,omitempty" yaml:"cache_key_form_fields,omitempty"`
	RequestHeaders     map[string]string `toml:"request_headers,omitempty" yaml:"request_headers,omitempty"`
	RequestParams      map[string]string `toml:"request_params,omitempty" yaml:"request_params,omitempty"`
	ResponseHeaders    map[string]string `toml:"response_headers,omitempty" yaml:"response_headers,omitempty"`
	ResponseCode       int               `toml:"response_code,omitempty" yaml:"response_code,omitempty"`
	ResponseBody       string            `toml:"response_body,omitempty" yaml:"response_body,omitempty"`
	CFName             string            `toml:"collapsed_forwarding,omitempty" yaml:"collapsed_forwarding,omitempty"`
	ReqRewriterName    string            `toml:"req_rewriter_name,omitempty" yaml:"req_rewriter_name,omitempty"`
	NoMetrics          bool              `toml:"no_metrics,omitempty" yaml:"no_metrics"`
}

type tls struct {
	FullChainCertPath  string   `toml:"full_chain_cert_path,omitempty" yaml:"full_chain_cert_path,omitempty"`
	PrivateKeyPath     string   `toml:"private_key_path,omitempty" yaml:"private_key_path,omitempty"`
	InsecureSkipVerify bool     `toml:"insecure_skip_verify,omitempty" yaml:"insecure_skip_verify,omitempty"`
	AuthorityPaths     []string `toml:"no_metrics,certificate_authority_paths" yaml:"certificate_authority_paths,omitempty"`
	ClientCertPath     string   `toml:"client_cert_path,omitempty" yaml:"client_cert_path,omitempty"`
	ClientKeyPath      string   `toml:"client_key_path,omitempty" yaml:"client_key_path,omitempty"`
}

type rule struct {
	NextRoute   string         `toml:"next_route,omitempty" yaml:"next_route,omitempty"`
	Ingress     string         `toml:"ingress_req_rewriter_name,omitempty" yaml:"ingress_req_rewriter_name,omitempty"`
	Egress      string         `toml:"egress_req_rewriter_name,omitempty" yaml:"egress_req_rewriter_name,omitempty"`
	NoMatch     string         `toml:"nomatch_req_rewriter_name,omitempty" yaml:"nomatch_req_rewriter_name,omitempty"`
	InputSource string         `toml:"input_source,omitempty" yaml:"input_source,omitempty"`
	InputKey    string         `toml:"input_key,omitempty" yaml:"input_key,omitempty"`
	InputType   string         `toml:"input_type,omitempty" yaml:"input_type,omitempty"`
	Encoding    string         `toml:"input_encoding,omitempty" yaml:"input_encoding,omitempty"`
	InputIndex  int            `toml:"input_index,omitempty" yaml:"input_index,omitempty"`
	Delimiter   string         `toml:"input_delimiter,omitempty" yaml:"input_delimiter,omitempty"`
	Operation   string         `toml:"operation,omitempty" yaml:"operation,omitempty"`
	OpArg       string         `toml:"operation_arg,omitempty" yaml:"operation_arg,omitempty"`
	Cases       map[string]*rc `toml:"cases,omitempty" yaml:"cases,omitempty"`
	RedirectURL string         `toml:"redirect_url,omitempty" yaml:"redirect_url,omitempty"`
	MaxExec     int            `toml:"max_rule_executions,omitempty" yaml:"max_rule_executions,omitempty"`
}

type rc struct {
	Matches         []string `toml:"matches,omitempty" yaml:"matches,omitempty"`
	ReqRewriterName string   `toml:"req_rewriter_name,omitempty" yaml:"req_rewriter_name,omitempty"`
	NextRoute       string   `toml:"next_route,omitempty" yaml:"next_route,omitempty"`
	RedirectURL     string   `toml:"redirect_url,omitempty" yaml:"redirect_url,omitempty"`
}

type rewriter struct {
	Instructions [][]string `toml:"instructions,omitempty" yaml:"instructions,omitempty"`
}

type tracing struct {
	TracerType    string            `toml:"tracer_type,omitempty" yaml:"-"`
	Provider      string            `toml:"-" yaml:"provider,omitempty"`
	ServiceName   string            `toml:"service_name,omitempty" yaml:"service_name,omitempty"`
	CollectorURL  string            `toml:"collector_url,omitempty" yaml:"collector_url,omitempty"`
	CollectorUser string            `toml:"collector_user,omitempty" yaml:"collector_user,omitempty"`
	CollectorPass string            `toml:"collector_pass,omitempty" yaml:"collector_pass,omitempty"`
	SampleRate    float64           `toml:"sample_rate,omitempty" yaml:"sample_rate,omitempty"`
	Tags          map[string]string `toml:"tags,omitempty" yaml:"tags,omitempty"`
	OmitTagsList  []string          `toml:"omit_tags,omitempty" yaml:"omit_tags,omitempty"`

	StdOutOptions stdout `toml:"stdout,omitempty" yaml:"stdout,omitempty"`
	JaegerOptions jaeger `toml:"jaeger,omitempty" yaml:"jaeger,omitempty"`
}

type jaeger struct {
	EndpointType string `toml:"endpoint_type,omitempty" yaml:"endpoint_type,omitempty"`
}

type stdout struct {
	PrettyPrint bool `toml:"pretty_print,omitempty" yaml:"pretty_print,omitempty"`
}

type metrics struct {
	ListenAddress string `toml:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	ListenPort    int    `toml:"listen_port,omitempty" yaml:"listen_port,omitempty"`
}

type reloading struct {
	ListenAddress  string `toml:"listen_address,omitempty" yaml:"listen_address,omitempty"`
	ListenPort     int    `toml:"listen_port,omitempty" yaml:"listen_port,omitempty"`
	HandlerPath    string `toml:"handler_path,omitempty" yaml:"handler_path,omitempty"`
	DrainTimeoutMS int    `toml:"drain_timeout_ms,omitempty" yaml:"drain_timeout_ms,omitempty"`
	RateLimitMS    int    `toml:"rate_limit_ms,omitempty" yaml:"rate_limit_ms,omitempty"`
}

type logging struct {
	LogFile  string `toml:"log_file,omitempty" yaml:"log_file,omitempty"`
	LogLevel string `toml:"log_level,omitempty" yaml:"log_level,omitempty"`
}
