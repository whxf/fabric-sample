//为state database 添加索引的方法，注意只有数据库中有数据之后才能运行下面的命令
//curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"transfer_time\":\"desc\"}]},\"ddoc\":\"indexTimeSortDoc\", \"name\":\"indexTimeSortDesc\",\"type\":\"json\"}" http://127.0.0.1:5984/mychannel_wallet/_index

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the Record structure.  Structure tags are used by encoding/json library
type Record struct {
	ObjectType     string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Sender         string `json:"sender"`
	Receiver       string `json:"receiver"`
	TransferAmount string `json:"transfer_amount"`
	TransferTime   string `json:"transfer_time"`
	TransferType   string `json:"transfer_type"`
}

/*
 * The Init method is called when the Smart Contract "wallet" is instantiated by the blockchain network
 */
func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "wallet"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := stub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryRecord" {
		return s.queryRecord(stub, args)
	} else if function == "initLedger" {
		return s.initLedger(stub)
	} else if function == "createRecord" {
		return s.createRecord(stub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

//func (s *SmartContract) queryRecord(stub shim.ChaincodeStubInterface, args []string) sc.Response {
//	if len(args) != 1 {
//		return shim.Error("Incorrect number of arguments. Expecting 1")
//	}
//
//	var err error
//
//	phone := args[0]
//
//	indexName := "sender~receiver~time~amount~type"
//	resultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, []string{phone})
//
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	defer resultsIterator.Close()
//
//	var recordJsonList = []Record{}
//	object_type := "record"
//
//	for resultsIterator.HasNext() {
//		queryResponse, _ := resultsIterator.Next()
//		_, compositeKeyParts, err := stub.SplitCompositeKey(queryResponse.Key)
//
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//
//		// Record is a JSON object, so we write as-is
//
//		result := Record{object_type,
//			compositeKeyParts[0],
//			compositeKeyParts[1],
//			compositeKeyParts[2],
//			compositeKeyParts[3],
//			compositeKeyParts[4]}
//
//		err = json.Unmarshal(queryResponse.Value, &result)
//		if err != nil {
//			return shim.Error(err.Error())
//		}
//
//		recordJsonList = append(recordJsonList, result)
//	}
//
//	recordJSONasBytes, err := json.Marshal(recordJsonList)
//
//	return shim.Success(recordJSONasBytes)
//}

func (s *SmartContract) queryRecord(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var err error

	phone := args[0]
	sqlString := fmt.Sprintf("{\"selector\": {\"$or\": [{\"sender\": \"%s\"},{\"receiver\": \"%s\"}]}, \"sort\": [{\"transfer_time\":\"desc\"}],\"limit\" : 100 }", phone, phone)
	resultsIterator, err := stub.GetQueryResult(sqlString)

	if err != nil {
		return shim.Error(err.Error())
	}

	defer resultsIterator.Close()

	var recordJsonList = []Record{}

	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()

		if err != nil {
			return shim.Error(err.Error())
		}

		// Record is a JSON object, so we write as-is

		result := Record{}
		err = json.Unmarshal(queryResponse.Value, &result)
		if err != nil {
			return shim.Error(err.Error())
		}

		recordJsonList = append(recordJsonList, result)
	}

	recordJSONasBytes, err := json.Marshal(recordJsonList)

	return shim.Success(recordJSONasBytes)
}

func (s *SmartContract) initLedger(stub shim.ChaincodeStubInterface) sc.Response {
	fmt.Println("- start init record")

	// ======== init record attribute ========
	// ======== create record object and json byte ========
	object_type := "record"
	records := []Record{
		Record{object_type,
			"test1",
			"test2",
			"100.0",
			"2006-01-02 15:04:05",
			"1"},
		Record{object_type,
			"test3",
			"test1",
			"100.0",
			"2006-01-03 00:04:05",
			"2"},
	}

	i := 0
	for i < len(records) {
		recordJSONasBytes, err := json.Marshal(records[i])
		if err != nil {
			return shim.Error(err.Error())
		}

		// ======== create composite key to record ========
		// ======== save record to state ==================
		indexName := "sender~receiver~time~amount~type"
		indexKey, err := stub.CreateCompositeKey(indexName, []string{records[i].Sender, records[i].Receiver,
			records[i].TransferTime, records[i].TransferAmount, records[i].TransferType})
		if err != nil {
			return shim.Error(err.Error())
		}

		stub.PutState(indexKey, recordJSONasBytes)
		i = i + 1
	}

	fmt.Println("- end init record")
	return shim.Success(nil)
}

func (s *SmartContract) createRecord(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	var cstZone = time.FixedZone("CST", 8*3600)
	record_sender := args[0]
	record_receiver := args[1]
	record_transfer_amount := args[2]
	record_transfer_time := time.Now().In(cstZone).Format("2006-01-02 15:04:05") // time to string format
	record_transfer_type := args[3]
	// string to time format : t, _ := time.Parse("2006-01-02 15:04:05", "2014-06-15 08:37:18")

	object_type := "record"
	record := Record{object_type,
		record_sender,
		record_receiver,
		record_transfer_amount,
		record_transfer_time,
		record_transfer_type}

	recordJSONasBytes, err := json.Marshal(record)

	if err != nil {
		return shim.Error(err.Error())
	}

	indexName := "sender~receiver~time~amount~type"
	indexKey, err := stub.CreateCompositeKey(indexName, []string{record.Sender, record.Receiver, record.TransferTime, record.TransferAmount, record.TransferType})

	if err != nil {
		return shim.Error(err.Error())
	}

	stub.PutState(indexKey, recordJSONasBytes)

	return shim.Success(nil)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}

}
