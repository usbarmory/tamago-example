module github.com/f-secure-foundry/tamago-example

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/f-secure-foundry/tamago v0.0.0-20200623140307-cbfc9242d09e
	github.com/golang/protobuf v1.4.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.20.5+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/sys v0.0.0-20200622214017-ed371f2e16b4 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	google.golang.org/protobuf v1.24.0 // indirect
	gvisor.dev/gvisor v0.0.0-20200623043054-38d7b2fe5630
)

replace gvisor.dev/gvisor => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
