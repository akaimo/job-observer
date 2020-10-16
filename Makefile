.PHONY: codegen
codegen:
	${GOPATH}/pkg/mod/k8s.io/code-generator@v0.17.0/generate-groups.sh \
		"deepcopy,client,informer,lister" \
		github.com/akaimo/job-observer/pkg/client \
		github.com/akaimo/job-observer/pkg/apis \
		cleaner:v1alpha1 \
		--go-header-file  hack/boilerplate.go.txt

.PHONY: run
run:
	go run ./cmd/controller/main.go -kubeconfig=${HOME}/.kube/config

REGISTRY := ghcr.io/akaimo/job-observer
VERSION  := 0.1.1

.PHONY: build-image
build-image:
	docker build -t $(REGISTRY):$(VERSION) .

.PHONY: push-image
push-image: build-image
	docker push $(REGISTRY):$(VERSION)

.PHONY: test
test:
	go test -v ./...

.PHONY: bundle
bundle:
	helm template helm/ --namespace job-observer > bundle.yaml
