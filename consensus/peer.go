package main

import (
	"context"
	"crypto/dsa"
	"crypto/sha512"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	election "github.com/mgaldi/blockchainweb/election"
	keys "github.com/mgaldi/blockchainweb/keyshelper"
)

const PORTS_GAP = 4
const RANGE_START = 41000
const RANGE_API_START = RANGE_START + 5000

var (
	sender, proposer int
	//noProposerCount               int
	extracted                     []Block
	pool                          []Book
	Blockchain                    []Block
	publicKeys                    []*dsa.PublicKey
	mutex                         = &sync.Mutex{}
	givenPort, i, n, assignedPort int
	stakeAmount                   int64
	otherPorts                    []int
	validatorSet                  election.ValidatorSet
	verbose                       bool
	publickey                     *dsa.PublicKey
	publicKeyPath, privateKeyPath string
)

type Message struct {
	MsgType   string
	Block     Block
	ProcessNr int `json:"processnr"`
	ToRelay   []RelayStructure
	PublicKey *dsa.PublicKey `json:"publickey"`
}
type RelayStructure struct {
	Value      Block
	Signatures []keys.BlockSignature
}
type Block struct {
	Height     int
	Timestamp  string
	ParentHash string
	RootHash   string
	Data       []Book
}
type Book struct {
	Title  string
	Author string
	Pages  int
	Year   int
}
type Validator struct {
	VP int64
	A  int64
}

func createBlock(data []Book) (block Block) {

	prev := Blockchain[len(Blockchain)-1]

	block.Height = prev.Height + 1
	block.Timestamp = time.Now().String()
	block.Data = data
	block.ParentHash = prev.RootHash
	block.RootHash = makeHash(block)

	return block
}
func makeHash(block Block) string {
	var titles []string
	var authors []string
	for _, b := range block.Data {
		titles = append(titles, b.Title)
		authors = append(authors, b.Author)
	}
	sort.Strings(titles)
	sort.Strings(authors)
	toHash := strings.Join(titles, "") + strings.Join(authors, "")
	record := fmt.Sprint(block.Height) + block.Timestamp + toHash + block.ParentHash
	h := sha512.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func handleConnection(c net.Conn) {

	var received Message
	dec := gob.NewDecoder(c)
	if err := dec.Decode(&received); err != nil {
		return
	}
	log.Println("Received: ", received)

	if received.MsgType == "p" || received.MsgType == "c" {
		mutex.Lock()
		log.Println("Received Message", received.MsgType, "from", received.ProcessNr, "Proposer for this round", proposer)

		if received.MsgType == "p" && proposer != received.ProcessNr && proposer != i {
			// if !blockVerify(received.Block, Blockchain[len(Blockchain)-1]) {
			// 	fmt.Println("Block proposed from wrong proposer not valid")
			// 	mutex.Unlock()
			// 	return
			// }
			// noProposerCount++
			// if noProposerCount == n-1 {
			// 	fmt.Println("NO PROPOSER FOUND")
			// 	proposer = validatorSet.ProposerElection()
			// 	noProposerCount = 0
			// 	newBlock := createBlock(received.Block.Data)
			// 	var signature []keys.BlockSignature
			// 	signature = append(signature, keys.SignBlockMessage(i, []byte(newBlock.RootHash)))

			// 	var relay []RelayStructure
			// 	relay = append(relay, RelayStructure{Value: newBlock, Signatures: signature})

			// 	msg := Message{MsgType: "p", Block: newBlock, ProcessNr: i, ToRelay: relay, PublicKey: publickey}
			// 	if proposer == i {

			// 		sender = 1
			// 		extracted = append(extracted, newBlock)

			// 	}
			// 	go organizeMessages(msg)
			// }
			mutex.Unlock()
			log.Println("Returning from check for wrong proposer")
			return
		}

		log.Println("Working on block from", received.ProcessNr, "Message is", received.MsgType, ". Proposer is", proposer)
		if received.MsgType == "p" {
			if !checkProposedBlock(received.Block) {
				mutex.Unlock()
				return
			}
		}
		//noProposerCount = 0
		var relay []RelayStructure

		log.Println("Checking blocks in relay")
		for _, s := range received.ToRelay {
			if verifyMsg(s, received.ProcessNr) {
				relay = append(relay, s)
				if len(extracted) == 0 {
					extracted = append(extracted, s.Value)
				} else {
					for _, v := range extracted {
						if s.Value.RootHash != v.RootHash {
							extracted = append(extracted, s.Value)
						}
					}
				}

			}
		}
		log.Println("Checked blocks received in relay")
		if len(relay) == 0 {
			log.Println("Nothing to send")
			mutex.Unlock()
			return
		}
		deliver := false
		f := n / 3
		if n%3 != 0 {
			f++
		}

		for _, m := range relay {
			if len(m.Signatures) != f+1 {
				break
			}
			deliver = true
		}
		if deliver {
			if len(extracted) == 1 {
				log.Println("DELIVER")
				log.Println("Verifying for extracted ", extracted[0])
				if blockVerify(extracted[0], Blockchain[len(Blockchain)-1]) {
					Blockchain = append(Blockchain, extracted[0])
				}
			} else {
				log.Println("SEND FAILURE")
			}
			extracted = nil
			mutex.Unlock()
			return
		}
		for k := range relay {
			relay[k].Signatures = append(relay[k].Signatures, keys.SignBlockMessage(i, ([]byte(relay[k].Value.RootHash))))
		}
		if sender == 0 {
			log.Println("Relaying the message for next round")
			msg := Message{MsgType: "c", Block: received.Block, ProcessNr: i, ToRelay: relay, PublicKey: publickey}
			organizeMessages(msg)
		}
		mutex.Unlock()

	}

	c.Close()
}
func checkProposedBlock(proposedBlock Block) bool {
	var proposedHashes []string
	for _, b := range proposedBlock.Data {
		record := b.Title + b.Author + fmt.Sprint(b.Pages) + fmt.Sprint(b.Year)
		h := sha512.New()
		h.Write([]byte(record))
		hashed := h.Sum(nil)
		proposedHashes = append(proposedHashes, hex.EncodeToString(hashed))
	}
	var poolHashes []string
	for _, b := range pool {
		record := b.Title + b.Author + fmt.Sprint(b.Pages) + fmt.Sprint(b.Year)
		h := sha512.New()
		h.Write([]byte(record))
		hashed := h.Sum(nil)
		poolHashes = append(poolHashes, hex.EncodeToString(hashed))
	}
	totalMatches := 0
	var indexMatches []int
	for _, x := range proposedHashes {
		for j, y := range poolHashes {
			if x == y {
				totalMatches++
				indexMatches = append(indexMatches, j)
			}
		}
	}
	occured := map[int]bool{}
	uniqueIndexes := []int{}
	for e := range indexMatches {
		if !occured[indexMatches[e]] {
			occured[indexMatches[e]] = true
			uniqueIndexes = append(uniqueIndexes, indexMatches[e])
		}
	}

	if totalMatches == len(proposedBlock.Data) {
		if len(uniqueIndexes) != len(pool) {
		uniqueIndexesLoop:
			for _, x := range uniqueIndexes {
				if len(pool) == 0 {
					pool = nil
					break uniqueIndexesLoop
				}
				pool[x] = pool[len(pool)-1]
				pool = pool[:len(pool)-1]
			}
		} else {
			pool = nil
		}
		return true
	}
	return false
}
func verifyMsg(msgSig RelayStructure, p int) bool {
	f := n / 3
	if n%3 != 0 {
		f++
	}
	if len(msgSig.Signatures) > f+1 {
		return false
	}
	if msgSig.Signatures[len(msgSig.Signatures)-1].ProcessNr != p {
		return false
	}
	keysCopy := make([]*dsa.PublicKey, n)
	copy(keysCopy, publicKeys)
	for j, v := range msgSig.Signatures {
		if v.ProcessNr == i && len(msgSig.Signatures) != f+1 && j != len(msgSig.Signatures)-1 {
			return false
		}
		if keysCopy[v.ProcessNr] == nil {
			return false
		}
		if !dsa.Verify(publicKeys[v.ProcessNr], []byte(makeHash(msgSig.Value)), v.R, v.S) {
			return false
		}
		keysCopy[v.ProcessNr] = nil
	}
	return true
}

func handleListenRound(l net.Listener) {

	for {

		c, err := l.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConnection(c)

	}

}
func blockVerify(block, prevBlock Block) bool {
	if (prevBlock.Height + 1) != block.Height {
		return false
	}
	if prevBlock.RootHash != block.ParentHash {
		return false
	}

	if makeHash(block) != block.RootHash {
		return false
	}
	return true

}
func listen(port int) {

	var err error
	var l net.Listener

	for {
		address := fmt.Sprintf("localhost:%d", port)
		l, err = net.Listen("tcp4", address)
		if err != nil {
			log.Printf("Couldn't listen on port %d\n", port)
			time.Sleep(5 * time.Second)

		} else {
			log.Printf("Listening on port %d\n ", port)
			break
		}
	}

	// defer l.Close()
	go handleListenRound(l)
}
func send(wg *sync.WaitGroup, msg Message, port int) {

	defer wg.Done()
	tries := 0
outLoop:
	for {
		var d net.Dialer
		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		peer := fmt.Sprintf("localhost:%d", port)
		conn, err := d.DialContext(ctx, "tcp", peer)
		if err != nil {
			if tries == 3 {
				break outLoop
			}
			log.Printf("Connection error while trying to send to port %d\n", port)
			log.Printf("Trying again in 5 seconds\n")
			time.Sleep(2 * time.Second)
			tries++
			continue
		}
		defer conn.Close()

		enc := gob.NewEncoder(conn)
		if err = enc.Encode(msg); err != nil {
			log.Fatal(err)
		}
		// log.Printf("Propose %d to %d, %v\n", msg.Value, port, msg.Signature[i])
		break
	}
}
func organizeMessages(msg Message) {

	var wg sync.WaitGroup
	wg.Add(n - 1)
	for _, port := range otherPorts {
		go send(&wg, msg, port)
	}

	wg.Wait()
}

func findOtherPorts(currentPort int) []int {

	var otherPorts []int
	for j := RANGE_START; j < (n*PORTS_GAP)+RANGE_START; j += PORTS_GAP {
		if j == currentPort {
			continue
		}
		otherPorts = append(otherPorts, j)
	}
	return otherPorts
}

func initDirectory() {
	if err := os.Mkdir("./keys", 0755); err != nil {
		log.Fatal(err)
	}
	keys.GenKey(privateKeyPath, publicKeyPath)
	publickey = keys.OpenPublic()
}

func init() {
	flag.IntVar(&givenPort, "p", -1, "port number")
	flag.IntVar(&i, "i", -1, "index of current process")
	flag.IntVar(&n, "n", -1, "total number of processes")
	flag.StringVar(&privateKeyPath, "key", "", "Private key for validator")
	flag.StringVar(&publicKeyPath, "public", "", "Public key for validator")
	flag.Int64Var(&stakeAmount, "s", -1, "Amount to stake")
	flag.BoolVar(&verbose, "v", false, "verbose")
	flag.Parse()
}

func init() {

	if len(privateKeyPath) == 0 || len(publicKeyPath) == 0 {
		privateKeyPath = fmt.Sprintf("./keys/private-key-%d.key", i)
		publicKeyPath = fmt.Sprintf("./keys/public-key-%d.crt", i)
	}
	keys.StorePaths(privateKeyPath, publicKeyPath)
	if _, err := os.Stat("keys"); err != nil {
		initDirectory()
		return
	}

	if _, err := os.Stat(privateKeyPath); err != nil {
		keys.GenKey(privateKeyPath, publicKeyPath)
	}

	publickey = keys.OpenPublic()
}

func main() {

	if i == -1 || n == -1 || stakeAmount == -1 {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if !verbose {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	if givenPort == -1 {
		assignedPort = (i * PORTS_GAP) + RANGE_START
		fmt.Printf("No port number provided. Port %d will be assigned to peer %d\n", assignedPort, i)
	} else {
		assignedPort = (i * PORTS_GAP) + RANGE_START
		if assignedPort != givenPort {
			fmt.Printf("Port number %d rassigned to usable range of ports. %d -> %d\n", givenPort, givenPort, assignedPort)

		}

	}

	sender = 0
	//noProposerCount = 0
	publicKeys = make([]*dsa.PublicKey, n)
	publicKeys[i] = publickey
	election.PostAccount(i, stakeAmount)
	zeroBlock := Block{}
	zeroBlock = Block{0, time.Now().String(), "", makeHash(zeroBlock), []Book{{Title: "Genesis Book", Author: "Marco Galdi", Pages: 999, Year: 2021}}}
	Blockchain = append(Blockchain, zeroBlock)
	rand.Seed(time.Now().UnixNano())
	go listen(assignedPort)
	otherPorts = findOtherPorts(assignedPort)
	startHttpServer()

	select {}
}

func startHttpServer() {
	go func() {
		createBlockHandler := func(w http.ResponseWriter, r *http.Request) {
			if r.Method == "POST" {
				if validatorSet.P == 0 {
					validatorSet.Init(n, publicKeys)
					proposer = validatorSet.ProposerElection()
				}
				decoder := json.NewDecoder(r.Body)
				var receivedData Book
				err := decoder.Decode(&receivedData)
				if err != nil {
					panic(err)
				}

				if validatorSet.P == 0 {
					validatorSet.Init(n, publicKeys)
					proposer = validatorSet.ProposerElection()

				}
				pool = append(pool, receivedData)
				if len(pool) != 2 {
					return
				}
				//CREATE PROPOSING BLOCK
				newBlock := createBlock(pool)
				if proposer != i {
					sender = 0
				} else {
					sender = 1
					extracted = append(extracted, newBlock)
				}
				// START CONSENSUS
				var signature []keys.BlockSignature
				signature = append(signature, keys.SignBlockMessage(i, []byte(newBlock.RootHash)))

				var relay []RelayStructure
				relay = append(relay, RelayStructure{Value: newBlock, Signatures: signature})

				msg := Message{MsgType: "p", Block: newBlock, ProcessNr: i, ToRelay: relay, PublicKey: publickey}

				organizeMessages(msg)
				io.WriteString(w, "OK")
			}

		}

		listBlockHandler := func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			json.NewEncoder(w).Encode(&Blockchain)
		}

		http.HandleFunc("/block/new", createBlockHandler)
		http.HandleFunc("/list", listBlockHandler)

		http.ListenAndServe(fmt.Sprintf(":%d", assignedPort+(RANGE_API_START-RANGE_START)), nil)
	}()
}
