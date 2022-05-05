package election

import (
	"bytes"
	"crypto/dsa"
	"crypto/sha512"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	keys "github.com/mgaldi/blockchainweb/keyshelper"
)

type ValidatorSet struct {
	ValidatorsList []Validator
	P              int64
	Avg            int64
}
type Validator struct {
	VP int64
	A  int64
}
type ValidatorMessage struct {
	ProcessNr   int
	StakeAmount int64
	Signature   keys.Signature
	Public      *dsa.PublicKey
}

func PostAccount(i int, amount int64) {
	h := sha512.New()
	message := fmt.Sprintf("%d%d", i, amount)
	h.Write([]byte(message))
	hashed := h.Sum(nil)
	signature := keys.SignMessage(hashed)
	validator := ValidatorMessage{ProcessNr: i, StakeAmount: amount, Signature: signature, Public: keys.GetPublic()}
	b := new(bytes.Buffer)
	err := json.NewEncoder(b).Encode(validator)
	if err != nil {
		fmt.Println(err)
	}
	_, err = http.Post("http://localhost:8080/validators", "application/json", b)
	if err != nil {
		fmt.Println(err)
	}
}
func findMAx(vSet []Validator) int {
	max := vSet[0].A
	maxIndex := 0
	for i, v := range vSet {
		if i == 0 {
			continue
		}
		if max < v.A {
			max = v.A
			maxIndex = i
		}
	}
	return maxIndex
}
func findMin(vSet []Validator) int {
	max := vSet[0].A
	minIndex := 0
	for i, v := range vSet {
		if i == 0 {
			continue
		}
		if max > v.A {
			max = v.A
			minIndex = i
		}
	}
	return minIndex
}
func (validatorSet *ValidatorSet) ProposerElection() int {

	// scale the priority values when new validators are added
	diff := findMAx(validatorSet.ValidatorsList) - findMin(validatorSet.ValidatorsList)
	threshold := int64(2) * validatorSet.P
	var scale int64
	if int64(diff) > threshold {
		scale = int64(diff) / threshold
		for i := range validatorSet.ValidatorsList {
			validatorSet.ValidatorsList[i].A = validatorSet.ValidatorsList[i].A / scale
		}
	}

	//When a validator is removed reorganize priorities with the new set
	avg := int64(0)
	for _, v := range validatorSet.ValidatorsList {
		avg += v.A
	}
	avg = avg / int64(len(validatorSet.ValidatorsList))
	for i := range validatorSet.ValidatorsList {
		validatorSet.ValidatorsList[i].A -= avg

	}

	//Calculate priority for validators
	for i := range validatorSet.ValidatorsList {
		validatorSet.ValidatorsList[i].A += validatorSet.ValidatorsList[i].VP

	}
	validatorSet.Avg = avg
	proposer := findMAx(validatorSet.ValidatorsList)
	validatorSet.ValidatorsList[proposer].A -= validatorSet.P
	return proposer
}

func getJson(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &target)
}
func (validatorSet *ValidatorSet) Init(n int, pkeys []*dsa.PublicKey) {

	vl := new([]ValidatorMessage)
	if err := getJson("http://localhost:8080/validators", &vl); err != nil {
		fmt.Println(err)
	}
	for j, v := range *vl {
		pkeys[v.ProcessNr] = v.Public
		h := sha512.New()
		message := fmt.Sprintf("%d%d", v.ProcessNr, v.StakeAmount)
		h.Write([]byte(message))
		hashed := h.Sum(nil)
		if !dsa.Verify(pkeys[v.ProcessNr], hashed, v.Signature.R, v.Signature.S) {
			(*vl)[j] = (*vl)[len((*vl))-1]
			(*vl) = (*vl)[:len((*vl))-1]
		}
	}

	validatorSet.P = 0
	validatorSet.Avg = 0
	validatorSet.ValidatorsList = make([]Validator, n)
	total := int64(0)
	for _, v := range *vl {
		total += v.StakeAmount
	}
	for _, v := range *vl {
		validatorSet.ValidatorsList[v.ProcessNr].VP = v.StakeAmount * 100 / total
		validatorSet.ValidatorsList[v.ProcessNr].A = 0
		validatorSet.P += validatorSet.ValidatorsList[v.ProcessNr].VP
	}

}
