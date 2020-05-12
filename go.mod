module github.com/f-secure-foundry/tamago-example

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/f-secure-foundry/tamago v0.0.0-20200512091917-158eff057e28
	github.com/golang/protobuf v1.4.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.20.4+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200510223506-06a226fb4e37
	golang.org/x/sys v0.0.0-20200511232937-7e40ca221e25 // indirect
	golang.org/x/time v0.0.0-20200416051211-89c76fbcd5d1 // indirect
	gvisor.dev/gvisor v0.0.0-20200429031301-ce19497c1c08
)

replace github.com/f-secure-foundry/tamago v0.0.0-20200512091917-158eff057e28 => /mnt/git/public/tamago
replace gvisor.dev/gvisor v0.0.0-20200429031301-ce19497c1c08 => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
