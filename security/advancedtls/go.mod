module google.golang.org/grpcv54/security/advancedtls

go 1.17

require (
	github.com/hashicorp/golang-lru v0.5.4
	golang.org/x/crypto v0.5.0
	google.golang.org/grpcv54 v1.52.0
	google.golang.org/grpcv54/examples v0.0.0-20201112215255-90f1b3ee835b
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	golang.org/x/net v0.8.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/text v0.8.0 // indirect
	google.golang.org/genproto v0.0.0-20230110181048-76db0878b65f // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace google.golang.org/grpcv54 => ../../

replace google.golang.org/grpcv54/examples => ../../examples
