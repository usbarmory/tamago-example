module github.com/f-secure-foundry/tamago-example

go 1.13

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v0.0.0-20191219182022-e17c9730c422
	github.com/f-secure-foundry/tamago v0.0.0-20200113153442-40c47e7efc2f
	github.com/mkevac/debugcharts v0.0.0-20191222103121-ae1c48aa8615
	golang.org/x/crypto v0.0.0-20200109152110-61a87790db17
	gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c
)

replace gvisor.dev/gvisor v0.0.0-20191224014503-95108940a01c => github.com/f-secure-foundry/gvisor v0.0.0-20191224100818-98827aa91607
