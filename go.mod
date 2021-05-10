module github.com/f-secure-foundry/tamago-example

go 1.16

require (
	github.com/arl/statsviz v0.4.0
	github.com/btcsuite/btcd v0.21.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/f-secure-foundry/crucible v0.0.0-20210503082702-01e44ec14e7a
	github.com/f-secure-foundry/tamago v0.0.0-20210510104352-4af0c35e76ff
	github.com/google/btree v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20210506145944-38f3c27a63bf
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
	golang.org/x/term v0.0.0-20210503060354-a79de5458b56
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba // indirect
	gvisor.dev/gvisor v0.0.0-20210507193914-e691004e0c6c
)

replace gvisor.dev/gvisor => github.com/f-secure-foundry/gvisor v0.0.0-20210201110150-c18d73317e0f
