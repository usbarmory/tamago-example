module github.com/f-secure-foundry/tamago-example

go 1.16

require (
	github.com/StackExchange/wmi v0.0.0-20210224194228-fe8f1750fd46 // indirect
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/f-secure-foundry/crucible v0.0.0-20210503082702-01e44ec14e7a
	github.com/f-secure-foundry/tamago v0.0.0-20210507073204-a2e021e0e2f6
	github.com/go-ole/go-ole v1.2.5 // indirect
	github.com/google/btree v1.0.1 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	github.com/shirou/gopsutil v3.21.4+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.5 // indirect
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210507014357-30e306a8bba5 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	gvisor.dev/gvisor v0.0.0-20210507010342-339001204000
)

replace gvisor.dev/gvisor => github.com/f-secure-foundry/gvisor v0.0.0-20210201110150-c18d73317e0f
