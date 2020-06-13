.PHONY: codegen
codegen:
	${GOPATH}/pkg/mod/k8s.io/code-generator@v0.17.0/generate-groups.sh \
		"deepcopy,client,informer,lister" \
		github.com/akaimo/job-observer/pkg/client \
		github.com/akaimo/job-observer/pkg/apis \
		cleaner:v1alpha1 \
		--go-header-file  hack/boilerplate.go.txt

