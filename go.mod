module github.com/tpretz/terraform-provider-zabbix

go 1.11

replace github.com/tpretz/go-zabbix-api => /mnt/c/gopath/src/github.com/tpretz/go-zabbix-api

require (
	github.com/hashicorp/terraform v0.12.23
	github.com/hashicorp/terraform-plugin-sdk v1.7.0
	github.com/tpretz/go-zabbix-api v0.14.0
)

//replace github.com/tpretz/go-zabbix-api => ../go-zabbix-api
