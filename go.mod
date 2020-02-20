module github.com/f-secure-foundry/tamago-example

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.1
	github.com/f-secure-foundry/tamago v0.0.0-20200218134009-a98d928a15ff
	github.com/golang/protobuf v1.3.3 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.20.1+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200219234226-1ad67e1f0ef4
	golang.org/x/sys v0.0.0-20200219091948-cb0a6d8edb6c // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gvisor.dev/gvisor v0.0.0-20200220022842-ec5630527bc4
)

replace gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
