module kubevault.dev/unsealer

go 1.12

require (
	cloud.google.com/go v0.41.0
	contrib.go.opencensus.io/exporter/ocagent v0.4.12 // indirect
	github.com/Azure/azure-sdk-for-go v31.1.0+incompatible
	github.com/Azure/go-autorest/autorest v0.7.0
	github.com/Azure/go-autorest/autorest/adal v0.4.0
	github.com/Azure/go-autorest/autorest/to v0.2.0
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/appscode/go v0.0.0-20191006073906-e3d193d493fc
	github.com/appscode/pat v0.0.0-20170521084856-48ff78925b79
	github.com/aws/aws-sdk-go v1.20.20
	github.com/ghodss/yaml v1.0.0
	github.com/go-openapi/jsonpointer v0.19.0 // indirect
	github.com/go-openapi/jsonreference v0.19.0 // indirect
	github.com/go-openapi/swag v0.19.0 // indirect
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/google/martian v2.1.1-0.20190517191504-25dcb96d9e51+incompatible // indirect
	github.com/googleapis/gax-go v2.0.2+incompatible // indirect
	github.com/hashicorp/vault/api v1.0.2
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/konsorten/go-windows-terminal-sequences v1.0.2 // indirect
	github.com/mailru/easyjson v0.0.0-20190403194419-1ea4449da983 // indirect
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/sirupsen/logrus v1.4.2
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/stretchr/testify v1.3.0
	golang.org/x/crypto v0.0.0-20190605123033-f99c8df09eb5
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.7.0
	google.golang.org/genproto v0.0.0-20190508193815-b515fa19cec8 // indirect
	k8s.io/api v0.0.0-20190503110853-61630f889b3c
	k8s.io/apimachinery v0.0.0-20190509063443-7d8f8feb49c5
	k8s.io/cli-runtime v0.0.0-20190516231937-17bc0b7fcef5 // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kube-openapi v0.0.0-20190502190224-411b2483e503 // indirect
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5 // indirect
	kmodules.xyz/client-go v0.0.0-20191006173540-91f8ee6b6b4b
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.34.0
	git.apache.org/thrift.git => github.com/apache/thrift v0.12.0
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v12.4.2+incompatible
	k8s.io/api => k8s.io/api v0.0.0-20190313235455-40a48860b5ab
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed
	k8s.io/apimachinery => github.com/kmodules/apimachinery v0.0.0-20190508045248-a52a97a7a2bf
	k8s.io/apiserver => github.com/kmodules/apiserver v0.0.0-20190811223248-5a95b2df4348
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190314002645-c892ea32361a
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190314000054-4a91899592f4
	k8s.io/klog => k8s.io/klog v0.3.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190314000639-da8327669ac5
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/metrics => k8s.io/metrics v0.0.0-20190314001731-1bd6a4002213
	k8s.io/utils => k8s.io/utils v0.0.0-20190221042446-c2654d5206da
	sigs.k8s.io/structured-merge-diff => sigs.k8s.io/structured-merge-diff v0.0.0-20190302045857-e85c7b244fd2
)
