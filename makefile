BINARY_NAME=terraform-provider-zabbix
BINARY_VERSION=0.15.1

build:
	go build -o ${BINARY_NAME} main.go
	mkdir -p ~/.terraform.d/plugins/terraform.local/local/zabbix/${BINARY_VERSION}/linux_amd64
	cp ${BINARY_NAME} ~/.terraform.d/plugins/terraform.local/local/zabbix/${BINARY_VERSION}/linux_amd64/${BINARY_NAME}_v${BINARY_VERSION}
	rm /home/fawal/zabbix/terraform/template/.terraform.lock.hcl
run:
	go build -o ${BINARY_NAME} main.go
	./${BINARY_NAME}
 
clean:
	go clean
	rm ${BINARY_NAME}
