module mlvm

go 1.20

replace github.com/ethereum/go-ethereum => github.com/ethereum-optimism/minigeth v0.0.0-20220614121031-c2b6152b4afb

replace github.com/unicorn-engine/unicorn => ../unicorn

replace mlgo => ../mlgo

require (
	mlgo v0.0.0
	github.com/ethereum/go-ethereum v1.10.8
	github.com/fatih/color v1.13.0
	github.com/unicorn-engine/unicorn v0.0.0-20211005173419-3fadb5aa5aad
)

require (
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5 // indirect
	golang.org/x/sys v0.6.0 // indirect
)

require (
	github.com/jessevdk/go-flags v1.5.0
	github.com/mattn/go-colorable v0.1.13
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db
	github.com/schollz/progressbar/v3 v3.13.1
	github.com/x448/float16 v0.8.4
	golang.org/x/exp v0.0.0-20230321023759-10a507213a29
)

require (
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	golang.org/x/sys v0.6.0 // indirect
	golang.org/x/term v0.6.0 // indirect
)