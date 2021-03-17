package emvcc

import (
	"fmt"
	"testing"
	"time"

	"github.com/hyperledger/fabric/core/ledger/internal/version"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/rwsetutil"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
)

func TestNewMVCCValidator(t *testing.T) {
	// AddNewMVCCValidator
	MVCCValidator := NewMVCCValidator()

	fmt.Println("Détection précose")
	MVCCValidator.AddKey("TX1", "key1")
	MVCCValidator.AddKey("TX2", "key2")
	MVCCValidator.AddKey("TX1", "key3")
	MVCCValidator.AddKey("TX1", "key5")
	fmt.Println(MVCCValidator)
	// res, ver := FindKey(MVCCValidator, "abc")
	// fmt.Println(res, ver)

	// AddNewRWset
	rwsetBuilder2 := rwsetutil.NewRWSetBuilder()
	rwsetBuilder2.AddToReadSet("ns1", "abc", version.NewHeight(1, 1))
	rwsetBuilder2.AddToReadSet("ns1", "def", version.NewHeight(1, 2))
	rwsetBuilder2.AddToWriteSet("ns1", "key5", []byte("value1_new"))
	rwsetBuilder2.AddToWriteSet("ns1", "key6", []byte("value2_new"))
	s := getTestPubSimulationRWSet(t, rwsetBuilder2)
	d := *s
	fmt.Print("s: ")
	fmt.Println(d)

	f := getTestPubSimulationRWSetByte(t, rwsetBuilder2)
	fmt.Println(f)
	// Check Validation

	code, err := TxValidation(MVCCValidator, getTestPubSimulationRWSet(t, rwsetBuilder2))
	fmt.Print("TX Validation Result1: ")
	fmt.Println(code, err)

	fmt.Println(MVCCValidator)
	MVCCValidator.AddToMVCCValidator(getTestPubSimulationRWSet(t, rwsetBuilder2), "TX3")

	// AddNewRWset
	rwsetBuilder1 := rwsetutil.NewRWSetBuilder()
	rwsetBuilder1.AddToReadSet("ns1", "key5", version.NewHeight(1, 1))
	rwsetBuilder1.AddToReadSet("ns1", "def", version.NewHeight(1, 2))
	// Check Validation
	code1, err1 := TxValidation(MVCCValidator, getTestPubSimulationRWSet(t, rwsetBuilder1))
	fmt.Print("TX Validation Result2: ")
	fmt.Println(code1, err1)
}

func getTestPubSimulationRWSet(t *testing.T, builders ...*rwsetutil.RWSetBuilder) *rwsetutil.TxRwSet {
	var pubRWSets []*rwsetutil.TxRwSet
	for _, b := range builders {
		s, e := b.GetTxSimulationResults()
		assert.NoError(t, e)
		sBytes, err := s.GetPubSimulationBytes()
		assert.NoError(t, err)
		pubRWSet := &rwsetutil.TxRwSet{}
		assert.NoError(t, pubRWSet.FromProtoBytes(sBytes))
		fmt.Print("1111: ")
		fmt.Println(pubRWSet.NsRwSets[0].KvRwSet.Reads[0].Key)
		pubRWSets = append(pubRWSets, pubRWSet)
	}
	return pubRWSets[0]
}

func getTestPubSimulationRWSetByte(t *testing.T, builders ...*rwsetutil.RWSetBuilder) *rwsetutil.TxRwSet {
	var pubRWSets []*rwsetutil.TxRwSet
	for _, b := range builders {
		s, e := b.GetTxSimulationResults()
		assert.NoError(t, e)
		sBytes, err := s.GetPubSimulationBytes()
		r, _ := b.GetTxReadWriteSet().ToProtoBytes()
		fmt.Print("555555: ")
		fmt.Println(r)
		assert.NoError(t, err)
		pubRWSet := &rwsetutil.TxRwSet{}
		assert.NoError(t, pubRWSet.FromProtoBytes(sBytes))
		fmt.Print("1111: ")
		fmt.Println(pubRWSet.NsRwSets[0].KvRwSet.Reads[0].Key)
		pubRWSets = append(pubRWSets, pubRWSet)
	}

	return pubRWSets[0]
}

func TestRemoveFromMVCCValidator(t *testing.T) {
	// AddNewMVCCValidator
	MVCCValidator := NewMVCCValidator()

	fmt.Println("TestRemoveFromMVCCValidator")
	MVCCValidator.AddKey("TX1", "key1")
	MVCCValidator.AddKey("TX2", "key2")
	MVCCValidator.AddKey("TX1", "key3")
	MVCCValidator.AddKey("TX1", "key5")
	fmt.Println(MVCCValidator)
	MVCCValidator.RemoveFromMVCCValidator("TX1")
	fmt.Println(MVCCValidator)
	MVCCValidator.RemoveFromMVCCValidator("TX3")
}

func TestCache(t *testing.T) {
	// AddNewMVCCValidator
	MVCCValidator := NewMVCCValidator()

	fmt.Println("TestCache")
	MVCCValidator.AddKey("TX1", "key1")
	MVCCValidator.AddKey("TX2", "key2")
	MVCCValidator.AddKey("TX1", "key3")
	MVCCValidator.AddKey("TX1", "key5")
	// fmt.Println(MVCCValidator)
	c := cache.New(5*time.Minute, 10*time.Minute)
	c.Set("MVCCValidator1", MVCCValidator, cache.NoExpiration)
	foo, found := c.Get("MVCCValidator1")
	if found {
		fmt.Println(foo)
	}
}
