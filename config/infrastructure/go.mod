module config-client/config/infrastructure

go 1.24.11

require (
	config-client/config/domain v0.0.0
	gorm.io/gorm v1.25.12
)

require (
	config-client/share v0.0.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	config-client/bom => ../../bom
	config-client/config/domain => ../domain
	config-client/share => ../../share
)
