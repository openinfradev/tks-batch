module github.com/openinfradev/tks-batch

go 1.16

require (
	github.com/argoproj/argo-workflows/v3 v3.1.13
	github.com/golang/mock v1.6.0
	github.com/openinfradev/tks-common v0.0.0-20221122025625-be9f8957ec3c
	github.com/openinfradev/tks-proto v0.0.6-0.20220831015809-fad377174017
	github.com/stretchr/testify v1.7.0
	gorm.io/driver/postgres v1.3.1
	gorm.io/gorm v1.23.3
)

replace github.com/openinfradev/tks-batch => ./

//replace github.com/openinfradev/tks-common => ./tks-common
//replace github.com/openinfradev/tks-proto => ./tks-proto
