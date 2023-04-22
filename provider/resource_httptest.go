package provider

import (
	"errors"
	//"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	//"github.com/hashicorp/terraform-plugin-sdk/helper/validation"

	"github.com/tpretz/go-zabbix-api"
)

// template resource function
func resourceHttpTest() *schema.Resource {
	return &schema.Resource{
		Create: resourceHttpTestCreate,
		Read:   resourceHttpTestRead,
		Update: resourceHttpTestUpdate,
		Delete: resourceHttpTestDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Web Scenario Display Name",
			},
			"template": &schema.Schema{
				Type:     schema.TypeString,
				Optional: false,
				Required: true,
				Description: "linked template",
			},
			"step": &schema.Schema{
				Type:        schema.TypeList,
				Description: "Step",
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Interface ID (internally generated)",
						},
						"url": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Interface DNS name",
						},
						"status_codes": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Comma delimited list of valid status codes",
						},
						"follow_redirects": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "follow redirects",
						},
						"no": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Description: "No",
						},
						"header": &schema.Schema{
							Type:        schema.TypeList,
							Description: "header",
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Header Name",
									},
									"value": &schema.Schema{
										Type:        schema.TypeString,
										Optional:    true,
										Description: "Header Value",
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func dataHttpTest() *schema.Resource {
	return &schema.Resource{
		Read: dataHttpTestRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Web Scenario Display Name",
			},
		},
	}
}

// terraform resource create handler
func resourceHttpTestCreate(d *schema.ResourceData, m interface{}) error {
	log.Trace("CREATING: BEGINING OF ALL THINGS")
	api := m.(*zabbix.API)
	log.Trace("CREATING: API CONNECTION ESTABLISHED")
	item := buildHttpTestObject(d)
	log.Trace("CREATING: BUILD HTTPTEST OBJECT")
	items := []zabbix.HttpTest{*item}
	log.Trace("CREATING: %+v", items[0])
	err := api.HttpTestCreate(items)

	if err != nil {
		return err
	}

	log.Trace("created httptest: %+v", items[0])

	d.SetId(items[0].HttpTestId)

	return resourceHttpTestRead(d, m)
}

// terraform template read handler (data source)
func dataHttpTestRead(d *schema.ResourceData, m interface{}) error {

	params := zabbix.Params{
		"filter":                map[string]interface{}{},
		"selectParentTemplates": "extend",
	}

	if v := d.Get("name").(string); v != "" {
		params["filter"].(map[string]interface{})["name"] = v
	}

	if len(params["filter"].(map[string]interface{})) < 1 {
		return errors.New("no filter parameters provided")
	}
	log.Debug("Lookup of httptest with: %#v", params)

	return httpTestRead(d, m, params)
}

// terraform template read handler (resource)
func resourceHttpTestRead(d *schema.ResourceData, m interface{}) error {
	log.Debug("Lookup of HttpTest with id %s", d.Id())

	return httpTestRead(d, m, zabbix.Params{
		"httptestids":           d.Id(),
		"selectParentTemplates": "extend",
	})
}

// generic template read function
func httpTestRead(d *schema.ResourceData, m interface{}, params zabbix.Params) error {
	api := m.(*zabbix.API)

	httptests, err := api.HttpTestGet(params)

	if err != nil {
		return err
	}

	if len(httptests) < 1 {
		d.SetId("")
		return nil
	}
	if len(httptests) > 1 {
		return errors.New("multiple httptests found")
	}
	t := httptests[0]

	log.Debug("Got httptest: %+v", t)

	d.Set("name", t.Name)
	d.Set("template", t.HostId)
	d.SetId(t.HttpTestId)

	return nil
}

// build a template object from terraform data
func buildHttpTestObject(d *schema.ResourceData) *zabbix.HttpTest {
	
	item := zabbix.HttpTest{
		Name:     d.Get("name").(string),
		HostId: 	d.Get("template").(string),
	}
	log.Trace("BUILDING!!!!")
	item.Steps = stepGenerate(d)
	log.Trace("BUILDING!!!!: %v",item.Steps)
	return &item
}

// used for updates since hostid is not required
func buildHttpTestObjectUpdate(d *schema.ResourceData) *zabbix.HttpTest {
	
	item := zabbix.HttpTest{
		Name:     d.Get("name").(string),
		//HostId: 	d.Get("template").(string),
	}
	log.Trace("UPDATE BUILDING!!!!")
	item.Steps = stepGenerate(d)
	log.Trace("UPDATING BUILDING!!!!: %v",item.Steps)
	return &item
}

// terraform update resource handler

func resourceHttpTestUpdate(d *schema.ResourceData, m interface{}) error {
	log.Trace("UPDATING: The start")
	api := m.(*zabbix.API)

	item := buildHttpTestObjectUpdate(d)
	item.HttpTestId = d.Id()

	// templates may need a bit extra effort
	if d.HasChange("template") {
		old, new := d.GetChange("template")
		diff := old.(*schema.Set).Difference(new.(*schema.Set))

		// removals, we need to unlink and clear
		if diff.Len() > 0 {
			item.TemplatesClear = buildTemplateIds(diff)
		}
	}

	items := []zabbix.HttpTest{*item}

	err := api.HttpTestUpdate(items)

	if err != nil {
		return err
	}

	return resourceHttpTestRead(d, m)

}


// terraform delete handler
func resourceHttpTestDelete(d *schema.ResourceData, m interface{}) error {
	api := m.(*zabbix.API)
	return api.HttpTestDeleteByIds([]string{d.Id()})
}
