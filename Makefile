APP_NAME ?= webhook-server
REGISTRY ?= debankur1
TAG ?= v1.54
WEBHOOK_SERVER_MANIFEST ?= manifests/webhook_server.yaml
MUTATING_WEBHOOK ?= manifests/webhooks.yaml
ENCODED_CA ?=""
WEBHOOK_SERVER_NAMESPACE ?= kube-system
CERT_DIR = certs
CERT_KEY = $(CERT_DIR)/tls.key
CERT_REQ = $(CERT_DIR)/tls.csr
CERT_EXT = $(CERT_DIR)/tls.ext
CERT_OUT = $(CERT_DIR)/tls.crt

build:
	go mod tidy
	GOOS=linux GOARCH=amd64 go build  -ldflags="-s -w" -o bin/webhook main.go

unit:

lint: ## Golang Static code analysis
	go version
	go mod tidy
	@if [ ! -x "`which golangci-lint 2>/dev/null`" ]; then \
		echo "golangci-lint is not found in PATH!!"; \
		exit 1; \
	fi
	golangci-lint --version
	@echo Running code static anaysis...
	golangci-lint run -v
	@echo ""

docker-build: build
	docker build -t $(REGISTRY)/$(APP_NAME):$(TAG) .
	docker push $(REGISTRY)/$(APP_NAME):$(TAG)

uninstall:
	kubectl delete -f $(MUTATING_WEBHOOK)
	kubectl delete -f $(WEBHOOK_SERVER_MANIFEST) -n $(WEBHOOK_SERVER_NAMESPACE)
	kubectl delete secret webhook-server-tls -n $(WEBHOOK_SERVER_NAMESPACE)

generate-tls-certs:
	#bash deploy_webhook.sh
	rm -rf certs
	mkdir -p certs
	openssl genrsa -out $(CERT_KEY) 2048
	openssl req -new -key $(CERT_KEY) -out $(CERT_REQ) -subj "/CN=$(APP_NAME).$(WEBHOOK_SERVER_NAMESPACE).svc"
	@echo "authorityKeyIdentifier=keyid,issuer" >> $(CERT_EXT)
	@echo "basicConstraints=CA:TRUE" >> $(CERT_EXT)
	@echo "keyUsage = digitalSignature, keyEncipherment" >> $(CERT_EXT)
	@echo "subjectAltName = @alt_names" >> $(CERT_EXT)
	@echo "" >> $(CERT_EXT)
	@echo "[alt_names]" >> $(CERT_EXT)
	@echo "DNS.1 = $(APP_NAME).$(WEBHOOK_SERVER_NAMESPACE).svc" >> $(CERT_EXT)
	openssl x509 -req -extfile $(CERT_EXT) -in $(CERT_REQ) -signkey $(CERT_KEY) -out $(CERT_OUT)

helm-install : generate-tls-certs docker-build
	kubectl create secret tls webhook-server-tls  --cert $(CERT_OUT) --key $(CERT_KEY) -n $(WEBHOOK_SERVER_NAMESPACE)
	@$(eval ENCODED_CA=$(shell cat certs/tls.crt | base64 | tr -d '\n' | sed 's/\"//g'))
	helm install compliance-webhook-server helm/webhook-server --set image.tag=$(TAG) --set webhook.caBundle=$(ENCODED_CA) -n kube-system
	helm install webhook-config helm/webhook --set webhook.caBundle=$(ENCODED_CA)

helm-install-bkp : generate-tls-certs docker-build
	@$(eval ENCODED_CA=$(shell cat certs/tls.crt | base64 | tr -d '\n' | sed 's/\"//g'))
	helm install compliance-webhook helm/compliance-webhook --set image.tag=$(TAG) --set secret.tls.crt="$(shell cat $(CERT_OUT) | base64 | tr -d '\n')" --set secret.tls.key="$(shell cat $(CERT_KEY) | base64 | tr -d '\n')" --set webhook.caBundle=$(ENCODED_CA) -n kube-system


helm-uninstall:
	kubectl delete secret webhook-server-tls -n $(WEBHOOK_SERVER_NAMESPACE)
	helm uninstall webhook-config
	helm uninstall compliance-webhook-server -n kube-system

install : generate-tls-certs
	kubectl create secret tls webhook-server-tls  --cert $(CERT_OUT) --key $(CERT_KEY) -n $(WEBHOOK_SERVER_NAMESPACE)
	sed -i '' 's|debankur1/webhook-server:latest|$(REGISTRY)/$(APP_NAME):$(TAG)|g' $(WEBHOOK_SERVER_MANIFEST)
	kubectl create -f $(WEBHOOK_SERVER_MANIFEST) -n $(WEBHOOK_SERVER_NAMESPACE)
	@$(eval ENCODED_CA=$(shell cat certs/tls.crt | base64 | tr -d '\n'))
	echo $(ENCODED_CA)
	sed -i '' 's|<ca_bundle>|$(ENCODED_CA)|g' $(MUTATING_WEBHOOK)
	kubectl create -f $(MUTATING_WEBHOOK)


redeploy: uninstall install