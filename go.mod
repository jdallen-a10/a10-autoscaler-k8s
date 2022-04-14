module main

go 1.17

replace k8sgo => ./k8s-go

replace a10/axapi => ./a10-golang-axapi

require (
	a10/axapi v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/tidwall/gjson v1.8.1
	gopkg.in/yaml.v2 v2.4.0
	k8sgo v0.0.0-00010101000000-000000000000
)

require (
	github.com/tidwall/match v1.0.3 // indirect
	github.com/tidwall/pretty v1.1.0 // indirect
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654 // indirect
)
