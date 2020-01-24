module github.com/f-secure-foundry/tamago-example

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.1
	github.com/f-secure-foundry/tamago v0.0.0-20200124143608-f486bef11ceb
	github.com/golang/protobuf v1.3.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v2.19.12+incompatible // indirect
	golang.org/x/crypto v0.0.0-20200117160349-530e935923ad
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
	golang.org/x/time v0.0.0-20191024005414-555d28b269f0 // indirect
	gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c
)

replace gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
