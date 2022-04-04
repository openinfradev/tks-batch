module github.com/openinfradev/tks-batch

go 1.16

require (
	github.com/argoproj/argo-workflows/v3 v3.1.13
	github.com/google/uuid v1.1.2
	github.com/openinfradev/tks-common v0.0.0-20220321044608-105302d33457
	github.com/openinfradev/tks-proto v0.0.6-0.20220324075944-e471af2c8c49
	gorm.io/driver/postgres v1.3.1
	gorm.io/gorm v1.23.3
)

replace github.com/openinfradev/tks-batch => ./

replace github.com/openinfradev/tks-common => ../tks-common
replace github.com/openinfradev/tks-proto => ../tks-proto
