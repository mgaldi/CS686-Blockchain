# Blockchain

  

The goal for the assignment is to create a website that interfaces with the underlying blockchain.
The blockchain will have a history of books that a user has inserted on the website.

The blockchain works on proof of stake. The peers join with an initial stake and based on the amount of stake, a peer will have a voting power.  The voting power is a percentage of the total stake which is the sum of all the stake from every other peer. Each peer will have a chance proportional to their stake to propose a block . The selection for the block proposer is made by a round-robin deterministic algorithm which is based on the voting power of each peer. So, every round, each peer knows who should propose the new block. A **new block** is proposed every **two new books** added through the web interface. Since this blockchain is synchronous, after the block is proposed, the other peers will validate the block by checking the rootHash, the height, the prevHash, and match the new books with their current list of pending books. Afma is then run on the new proposed bock in order to make sure that every other peer has received the right block, isn't crashed, or is not byzantine.

  
  

# Folders and Files

```
.
+-- _consensus
|   +-- peer.go
|   +-- go.mod
+-- _election
|   +-- election.go
|   +-- go.mod
+-- _keyshelper
|   +-- KeysHelper.go
|   +-- go.mod
+-- _server
|   +-- webserver.go
|  	+-- _blockchainweb
|   	+-- build
		+-- _src
```
**Folders**
**consensus**: contains `peer.go` and the folder `keys` which is created on runtime for storing the private and public keys.
**election**: contains the module election from `election.go` 
**keyshelper**: contains the module keyshelper from `keyshelper.go`
**server**: contains the file `webserver.go` and the folder `blockchainweb` with inside `build` and `src` for the view of the website.
**Files**
-**peer.go**: This file contains all the code necessary to run a peer.
-**election.go**:This file is used as a module for `peer.go`. It handles the proposers for the new 	blocks.
-**keyshelper.go**:This module helps with handling the private and public keys.
-**webserver.go**:This file runs the server for interfacing with the blockchain
-**Blocks.js**:Is located in `/server/blockchainweb/src`  and contains the code for handling the view for the website.
  
  ## peer.go API
Other than the `assignedPort` which `peer.go` will listen to, it will also listen on `assignedPort + 4000` for receiving new books and for showing the current state of the blockchain.
The peer will accept post request on `localhost:peerport+4000/block/new`  for adding new books and on `localhost:peerport+4000/list` for showing the current state of the blockchain.
 ### webserver.go requests  to peer.go
 The webserver uses the list of peers which joined with stake that can be found on `http://localhost:8080/validators`.
 When new data is submitted, the webserver will relay the message to all peers.
 In order to populate the view for the current state of the blockchain, the webserver asks to every peer on `localhost:peerport+4000/list`. If all the latest heights and the root hashes match, the view is populated.
 
## Usage
Files to run: `webserver.go` `peer.go`

There are no flags required for `webserver.go`

For `peer.go`, the program needs different flags:

-**i**: the peer number.
-**n**: the maximum number of peers.
-**p**: the port number.
-**s**: the stake amount

### Optional flags

-**v**: verbose.


## Example
**For the web interface**
`cd server && go run webserver.go`

**For the peers**
`cd consensus`

`go run peer.go -i 0 -n 4 -p 41000 -s 120`

`go run peer.go -i 1 -n 4 -p 41004 -s 400`

`go run peer.go -i 2 -n 4 -p 41008 -s 320`

`go run peer.go -i 3 -n 4 -p 41012 -s 120`

Test by visiting `http://localhost:8080/`
Insert the fields for the adding a book: `Title, Author, Pages, Year.`
Then `Submit`
The blocks are created **every two books** added. So, add at least two books to see the new state of the blockchain.
