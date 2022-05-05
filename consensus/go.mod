module github.com/mgaldi/blockchainweb/consensus

go 1.15

replace github.com/mgaldi/blockchainweb/keyshelper => ../keyshelper

replace github.com/mgaldi/blockchainweb/election => ../election

require (
	github.com/mgaldi/blockchainweb/election v0.0.0-00010101000000-000000000000
	github.com/mgaldi/blockchainweb/keyshelper v0.0.0-00010101000000-000000000000
)
