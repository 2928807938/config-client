module config-client/config/domain

go 1.24.11

require config-client/share v0.0.0

replace (
	config-client/bom => ../../bom
	config-client/share => ../../share
)
