
# This assumes a build in the .devcontainer Dockerfile environment
FLUX_SCHED_ROOT ?= /opt/flux-sched
INSTALL_PREFIX ?= /usr
COMMONENVVAR=GOOS=$(shell uname -s | tr A-Z a-z)

# This is what worked
# GOOS=linux CGO_CFLAGS="-I/home/flux-sched/resource/reapi/bindings/c" CGO_LDFLAGS="-L/usr/lib -L/home/flux-sched/resource -lresource -L/home/flux-sched/resource/libjobspec -ljobspec_conv -L/home/flux-sched/resource/reapi/bindings -lreapi_cli -lflux-idset -lstdc++ -lczmq -ljansson -lhwloc -lboost_system -lflux-hostlist -lboost_graph -lyaml-cpp" go build -ldflags '-w' -o bin/server cmd/main.go
BUILDENVVAR=CGO_CFLAGS="-I${FLUX_SCHED_ROOT}/resource/reapi/bindings/c" CGO_LDFLAGS="-L${INSTALL_PREFIX}/lib -L${FLUX_SCHED_ROOT}/resource -lfluxion-resource -L${FLUX_SCHED_ROOT}/resource/libjobspec -ljobspec_conv -L/${FLUX_SCHED_ROOT}/resource/reapi/bindings -lreapi_cli -lflux-idset -lstdc++ -lczmq -ljansson -lhwloc -lboost_system -lflux-hostlist -lboost_graph -lyaml-cpp"
RELEASE_VERSION?=v$(shell date +%Y%m%d)-$(shell git describe --tags --match "v*")


.PHONY: all
all: flex

.PHONY: flex
flex: 
	# This needs to match the flux-sched install and latest commit, for now we are using a fork of milroy's branch
	# that has a go.mod updated to match the org name
	# go get -u github.com/researchapps/flux-sched/resource/reapi/bindings/go/src/fluxcli@0f33b17f6e792c14a262609d71f4ea5f32cb3ebb
	go mod tidy
	mkdir -p ./bin
	$(COMMONENVVAR) $(BUILDENVVAR) go build -ldflags '-w' -o bin/aws-topology src/cmd/main.go

.PHONY: clean
clean:
	rm -rf ./bin/server