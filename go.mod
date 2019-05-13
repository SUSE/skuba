module github.com/SUSE/caaspctl

require (
	github.com/MakeNowJust/heredoc v0.0.0-20171113091838-e9091a26100e // indirect
	github.com/PuerkitoBio/purell v1.1.1 // indirect
	github.com/Sirupsen/logrus v1.0.6 // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/emicklei/go-restful v2.9.3+incompatible // indirect
	github.com/evanphx/json-patch v3.0.0+incompatible // indirect
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d // indirect
	github.com/fatih/camelcase v1.0.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-openapi/spec v0.18.0 // indirect
	github.com/go-openapi/swag v0.18.0 // indirect
	github.com/go-yaml/yaml v2.1.0+incompatible
	github.com/gogo/protobuf v1.2.1 // indirect
	github.com/google/btree v1.0.0 // indirect
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/json-iterator/go v1.1.5 // indirect
	github.com/liggitt/tabwriter v0.0.0-20181228230101-89fcab3d43de // indirect
	github.com/lithammer/dedent v1.1.0 // indirect
	github.com/mailru/easyjson v0.0.0-20190221075403-6243d8e04c3f // indirect
	github.com/mitchellh/go-wordwrap v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/pkg/errors v0.8.1
	github.com/russross/blackfriday v0.0.0-20151117072312-300106c228d5 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/sirupsen/logrus v1.3.0 // indirect
	github.com/spf13/afero v1.2.2 // indirect
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.3
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	golang.org/x/crypto v0.0.0-20190513172903-22d7a77e9e5f
	golang.org/x/oauth2 v0.0.0-20190220154721-9b3c75971fc9 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/yaml.v2 v2.2.2 // indirect
	k8s.io/api v0.0.0-20190409092523-d687e77c8ae9
	k8s.io/apiextensions-apiserver v0.0.0-20190315093550-53c4693659ed // indirect
	k8s.io/apimachinery v0.0.0-20190409092423-760d1845f48b
	k8s.io/cli-runtime v0.0.0-20190314001948-2899ed30580f // indirect
	// client-go is master branch see https://github.com/kubernetes/kubernetes/issues/76304
	k8s.io/client-go v0.0.0-20190409092706-ca8df85b1798
	k8s.io/cloud-provider v0.0.0-20190405093944-6c8b65ee8f98 // indirect
	k8s.io/cluster-bootstrap v0.0.0-20190314002537-50662da99b70
	k8s.io/klog v0.2.0
	k8s.io/kube-proxy v0.0.0-20190314002154-4d735c31b054 // indirect
	k8s.io/kubelet v0.0.0-20190314002251-f6da02f58325 // indirect
	k8s.io/kubernetes v1.14.1
	sigs.k8s.io/kustomize v2.0.3+incompatible // indirect
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

//fix for https://github.com/kubernetes-sigs/controller-runtime/issues/219
exclude github.com/Sirupsen/logrus v1.4.1

exclude github.com/Sirupsen/logrus v1.4.0

exclude github.com/Sirupsen/logrus v1.3.0

exclude github.com/Sirupsen/logrus v1.2.0

exclude github.com/Sirupsen/logrus v1.1.1

exclude github.com/Sirupsen/logrus v1.1.0

exclude github.com/renstrom/dedent v1.1.0
