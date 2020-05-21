module github.com/f-secure-foundry/tamago-example

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/f-secure-foundry/tamago v0.0.0-20200520124612-ac702d587a63
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.20.4+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/sys v0.0.0-20200519105757-fe76b779f299 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gvisor.dev/gvisor v0.0.0-20200521055007-059879e14301
)

replace gvisor.dev/gvisor v0.0.0-20200521055007-059879e14301 => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
