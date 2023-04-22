package provider

import (
	logger "log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/tpretz/go-zabbix-api"
)

// Provider definition
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API username",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_USER", "ZABBIX_USERNAME"}, nil),
			},
			"password": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API password",
				ValidateFunc: validation.StringIsNotWhiteSpace,
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_PASS", "ZABBIX_PASSWORD"}, nil),
			},
			"url": &schema.Schema{
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Zabbix API url",
				DefaultFunc:  schema.MultiEnvDefaultFunc([]string{"ZABBIX_URL", "ZABBIX_SERVER_URL"}, nil),
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"tls_insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "Disable TLS certificate checking (for testing use only)",
				Optional:    true,
				Default:     false,
			},
			"serialize": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Serialize API requests, if required due to API race conditions",
			},
		},
		DataSourcesMap: map[string]*schema.Resource{
			"zabbix_host":        dataHost(),
			"zabbix_application": dataApplication(),
			"zabbix_proxy":       dataProxy(),
			"zabbix_hostgroup":   dataHostgroup(),
			"zabbix_template":    dataTemplate(),
			"zabbix_httptest":    dataHttpTest(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"zabbix_trigger":       resourceTrigger(),
			"zabbix_proto_trigger": resourceProtoTrigger(),
			"zabbix_template":      resourceTemplate(),
			"zabbix_hostgroup":     resourceHostgroup(),
			"zabbix_host":          resourceHost(),
			"zabbix_application":   resourceApplication(),
			"zabbix_httptest":   resourceHttpTest(),

			"zabbix_graph":       resourceGraph(),
			"zabbix_proto_graph": resourceProtoGraph(),

			"zabbix_item_trapper":       resourceItemTrapper(),
			"zabbix_proto_item_trapper": resourceProtoItemTrapper(),
			"zabbix_lld_trapper":        resourceLLDTrapper(),

			"zabbix_item_http":       resourceItemHttp(),
			"zabbix_proto_item_http": resourceProtoItemHttp(),
			"zabbix_lld_http":        resourceLLDHttp(),

			"zabbix_item_simple":       resourceItemSimple(),
			"zabbix_proto_item_simple": resourceProtoItemSimple(),
			"zabbix_lld_simple":        resourceLLDSimple(),

			"zabbix_item_external":       resourceItemExternal(),
			"zabbix_proto_item_external": resourceProtoItemExternal(),
			"zabbix_lld_external":        resourceLLDExternal(),

			"zabbix_item_internal":       resourceItemInternal(),
			"zabbix_proto_item_internal": resourceProtoItemInternal(),
			"zabbix_lld_internal":        resourceLLDInternal(),

			"zabbix_item_snmp":       resourceItemSnmp(),
			"zabbix_proto_item_snmp": resourceProtoItemSnmp(),
			"zabbix_lld_snmp":        resourceLLDSnmp(),

			"zabbix_item_snmptrap":       resourceItemSnmpTrap(),
			"zabbix_proto_item_snmptrap": resourceProtoItemSnmpTrap(),

			"zabbix_item_agent":       resourceItemAgent(),
			"zabbix_proto_item_agent": resourceProtoItemAgent(),
			"zabbix_lld_agent":        resourceLLDAgent(),

			"zabbix_item_aggregate":       resourceItemAggregate(),
			"zabbix_proto_item_aggregate": resourceProtoItemAggregate(),

			"zabbix_item_calculated":       resourceItemCalculated(),
			"zabbix_proto_item_calculated": resourceProtoItemCalculated(),

			"zabbix_item_dependent":       resourceItemDependent(),
			"zabbix_proto_item_dependent": resourceProtoItemDependent(),
			"zabbix_lld_dependent":        resourceLLDDependent(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func getApiVersion(api *zabbix.API) (version int64, err error) {
	var vstr string
	vstr, err = api.Version()
	if err != nil {
		log.Trace("api version got error: %+v", err)
		return
	}

	parts := strings.Split(vstr, ".")

	version, err = strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return
	}
	version = version * 10000

	// do we have a minor version
	if len(parts) > 1 {
		var no int64
		no, err = strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return
		}
		version += no * 100
	}

	// do we have a patch version
	if len(parts) > 2 {
		var no int64
		no, err = strconv.ParseInt(parts[2], 10, 64)
		if err != nil {
			return
		}
		version += no
	}
	log.Trace("version is: %d", version)
	return
}

// providerConfigure configure this provider
func providerConfigure(d *schema.ResourceData) (meta interface{}, err error) {
	log.Trace("Started zabbix provider init")
	l := logger.New(stderr, "[DEBUG] ", logger.LstdFlags)

	api := zabbix.NewAPI(zabbix.Config{
		Url:         d.Get("url").(string),
		TlsNoVerify: d.Get("tls_insecure").(bool),
		Log:         l,
		Serialize:   d.Get("serialize").(bool),
	})

	var version int64
	version, err = getApiVersion(api)
	if err != nil {
		return
	}

	api.Config.Version = int(version)

	_, err = api.Login(d.Get("username").(string), d.Get("password").(string))
	meta = api
	log.Trace("Started zabbix provider got error: %+v", err)

	return
}

// tagGenerate build tag structs from terraform inputs
func tagGenerate(d *schema.ResourceData) (tags zabbix.Tags) {
	set := d.Get("tag").(*schema.Set).List()
	tags = make(zabbix.Tags, len(set))

	for i := 0; i < len(set); i++ {
		current := set[i].(map[string]interface{})
		tags[i] = zabbix.Tag{
			Tag:   current["key"].(string),
			Value: current["value"].(string),
		}
	}

	return
}

func stepGenerate(d *schema.ResourceData) (steps zabbix.Steps) {
	log.Trace("STEP!!!!!!!!!!!!!!!!!!!!LOADING STEPS")
	set := d.Get("step").([]interface{})
	//set := d.Get("step")
	//for _,item:=range set :
	


	log.Trace("SET!!!!!!!!!!!!!!!!!!!! %v",set)
	steps = make(zabbix.Steps, len(set))
  
	log.Trace("STEPS!!!!!!!!!!!!!!!!!!!! %v",steps)
	for i := 0; i < len(set); i++ {
		current := set[i].(map[string]interface{})

		log.Trace("CURRENT!!!!!!! %v", current)
		//step_no, _ := strconv.Atoi(current["no"].(string))
		headers := make(zabbix.Headers, len(current["header"].([]interface{})))
		log.Trace("HEADERS!!!!!!! %v", headers)
		for j := 0; j < len(headers); j++ {
			log.Trace("HEADER count!!!!!!! %v", j)
			log.Trace("CURRENT HEADER!!!!!!! %v", current["header"])
			headerList := current["header"].([]interface{})
			log.Trace("HEADERLIST!!!!!!! %v", headerList)	
			current_headers := headerList[j].(map[string]interface{})
			log.Trace("ACTIVE HEADER!!!!!!! %v", current_headers)
			headers[j] = zabbix.Header{
				Name:   current_headers["name"].(string),
				Value:  current_headers["value"].(string),
			}
		}

		steps[i] = zabbix.Step{
			Name:   current["name"].(string),
			Url: current["url"].(string),
			StatusCodes: current["status_codes"].(string),
			FollowRedirects: current["follow_redirects"].(string),
			No: current["no"].(string),
			Headers: headers,
		}
	}
	//log.Trace("STEP!!!!!!!!!!!!!!!!!!!!: %+v", steps)
	return
}

// flattenTags convert response to terraform input
func flattenTags(list zabbix.Tags) *schema.Set {
	set := schema.NewSet(func(i interface{}) int {
		m := i.(map[string]interface{})
		return hashcode.String(m["key"].(string) + "V" + m["value"].(string))
	}, []interface{}{})
	for i := 0; i < len(list); i++ {
		set.Add(map[string]interface{}{
			"key":   list[i].Tag,
			"value": list[i].Value,
		})
	}
	return set
}
