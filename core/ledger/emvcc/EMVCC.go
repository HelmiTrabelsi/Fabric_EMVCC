package emvcc

import (
	"fmt"

	"github.com/hyperledger/fabric-protos-go/ledger/rwset/kvrwset"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/patrickmn/go-cache"
)

var logger = flogging.MustGetLogger("validation")

type MVCCValidator map[string][]string

// NewMVCCValidator constructs a new instance of MVCCValidator
func NewMVCCValidator() MVCCValidator {
	MVCCValidator := make(map[string][]string)
	return MVCCValidator
	// return make(map[string][]string)
}

// AddNewKey to the MVCCValidator
func (MVCCValidator MVCCValidator) AddKey(TXId string, key string) {
	MVCCValidator[TXId] = append(MVCCValidator[TXId], key)
}

// FindKey in the MVCCValidator
func FindKey(MVCCValidator map[string][]string, Key string) (bool, string) {
	for _, key_list := range MVCCValidator {
		for i := range key_list {
			// fmt.Println(tx, key_value_list[i].Key)
			if Key == key_list[i] {
				return true, Key
			}
		}
	}
	return false, ""
}

// TxValidation
func TxValidation(MVCCValidator map[string][]string, kvReads []*kvrwset.KVRead) (peer.TxValidationCode, error) {

	if valid, err := validateReadSet(MVCCValidator, kvReads); !valid || err != nil {
		if err != nil {
			return peer.TxValidationCode(-1), err
		}
		return peer.TxValidationCode_MVCC_READ_CONFLICT, nil
	}

	return peer.TxValidationCode_VALID, nil
}

// validateReadSet
func validateReadSet(MVCCValidator map[string][]string, kvReads []*kvrwset.KVRead) (bool, error) {
	for _, kvRead := range kvReads {
		if valid, err := validateKVRead(MVCCValidator, kvRead); !valid || err != nil {
			return valid, err
		}
	}
	return true, nil
}

// validateKVRead performs mvcc check for a key read during transaction simulation.
func validateKVRead(MVCCValidator map[string][]string, kvRead *kvrwset.KVRead) (bool, error) {
	fmt.Print("Rset key: ")
	fmt.Println(kvRead.Key)
	res, _ := FindKey(MVCCValidator, kvRead.Key)
	// fmt.Println(res)
	fmt.Print("Result of search: ")
	fmt.Println(res)
	if res {

		return false, nil
	}
	return true, nil
}

func (MVCCValidator MVCCValidator) AddToMVCCValidator(kvWrites []*kvrwset.KVWrite, Txid string) {
	for _, kvWrite := range kvWrites {
		// fmt.Println(nsRWSet.KvRwSet.Writes[i].Key)
		MVCCValidator.AddKey(Txid, kvWrite.Key)
	}
}

func (MVCCValidator MVCCValidator) RemoveFromMVCCValidator(Txid string) {
	if _, ok := MVCCValidator[Txid]; ok {
		delete(MVCCValidator, Txid)
		fmt.Println(" Tx" + Txid + " Removed from MVCCValidator")
	} else {
		fmt.Println("Tx" + Txid + "Not found in the MVCCValidator")
	}
}

func CreateCach() *cache.Cache {
	c := cache.New(cache.NoExpiration, cache.NoExpiration)
	return c
}

func SetMVCCValidatorInCach(cachewithID *cache.Cache, MVCCValidator MVCCValidator) bool {
	c := cachewithID
	c.Set("MVCCValidator", MVCCValidator, cache.NoExpiration)
	return true
}

/*func GetMVCCValidatorFromCach(cachewithID *cache.Cache) (MVCCValidator, error) {

	MVCCValidatorVIde := NewMVCCValidator()
	return MVCCValidatorVIde, nil

}*/

func GetMVCCValidatorFromCach(cachewithID *cache.Cache) MVCCValidator {
	MVCCValidator1, found := cachewithID.Get("MVCCValidator")
	if !found {
		// fmt.Println("MVCC Validator Not Found")
		MVCCValidatorVIde := NewMVCCValidator()
		SetMVCCValidatorInCach(cachewithID, MVCCValidatorVIde)
		return MVCCValidatorVIde
	}
	return MVCCValidator1.(MVCCValidator)
}

// fmt.Println("Here")

// fmt.Println(reflect.TypeOf(MVCCValidator1))
