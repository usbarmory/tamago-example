module github.com/f-secure-foundry/tamago-example

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.1
	github.com/f-secure-foundry/tamago v0.0.0-20200330171548-98026a2a018f
	github.com/golang/protobuf v1.3.5 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.20.2+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200323165209-0ec3e9974c59
	golang.org/x/sys v0.0.0-20200327173247-9dae0f8f5775 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c
)

replace gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
