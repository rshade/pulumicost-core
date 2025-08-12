module github.com/rshade/pulumicost-core

go 1.24.5

require (
	github.com/rshade/pulumicost-spec v0.0.0-20240101000000-000000000000
	github.com/spf13/cobra v1.8.0
	google.golang.org/grpc v1.62.0
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/rshade/pulumicost-spec => ../pulumicost-spec

require (
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/sys v0.16.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240123012728-ef4313101c80 // indirect
	google.golang.org/protobuf v1.32.0 // indirect
)
