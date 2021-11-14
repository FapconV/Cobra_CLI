module NetSim/cmd

go 1.16

require (
	NetSim/packet v0.0.0-00010101000000-000000000000
	NetSim/telemetry v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.5.2
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	google.golang.org/protobuf v1.27.1
)

replace NetSim/telemetry => ../telemetry

replace NetSim/packet => ../packet
