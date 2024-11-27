app ?= azure-nodepool-metrics
version ?= $$(cat VERSION)

lint:
	@echo "Linting ${app}"
	@docker run --rm -v $$(pwd)/${app}:/app -w /app golangci/golangci-lint:v1.62.0 golangci-lint run -v --timeout 600s

build:
	docker build . -t ${app}:${version}

run-local: build
	docker run -p 8002:8002 \
		-v ~/.kube:/root/.kube \
		-v ~/.azure:/root/.azure \
		-e AZURE_TENANT_ID \
		-e AZURE_SUBSCRIPTION_ID \
		-e AZURE_CLIENT_ID \
		-e AZURE_CLIENT_SECRET \
		${app}:${version}
