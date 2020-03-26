module go.borchero.com/cuckoo

go 1.13

require (
	cloud.google.com/go v0.55.0
	cloud.google.com/go/storage v1.6.0
	github.com/aws/aws-sdk-go v1.29.32
	github.com/containerd/console v0.0.0-20191219165238-8375c3424e4d
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/markbates/pkger v0.15.0
	github.com/moby/buildkit v0.7.0
	github.com/spf13/cobra v0.0.6
	github.com/xanzy/go-gitlab v0.29.0
	go.borchero.com/typewriter v0.5.5
	go.mozilla.org/sops/v3 v3.5.0
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	google.golang.org/api v0.20.0
	google.golang.org/genproto v0.0.0-20200317114155-1f3552e48f24
	gopkg.in/yaml.v2 v2.2.8
	gotest.tools v2.2.0+incompatible
	helm.sh/helm/v3 v3.1.2
	k8s.io/client-go v0.17.2
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.0.1+incompatible
	github.com/containerd/containerd => github.com/containerd/containerd v1.3.0
	github.com/docker/docker => github.com/docker/docker v1.4.2-0.20200204220554-5f6d6f3f2203
	github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305
)
