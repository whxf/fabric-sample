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
	ObjectType     string  `json:"docType"` //docType is used to distinguish the various types of objects in state database
	Sender         string  `json:"sender"`
	Receiver       string  `json:"receiver"`
	TransferAmount float64 `json:"transfer_amount"`
	TransferTime   string  `json:"transfer_time"`
	TransferType   int     `json:"transfer_type"`
}

/*
 * The Init method is called when the Smart Contract "wallet" is instantiated by the blockchain network
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {

	var err error

	fmt.Println("- start init record")

	// ======== init record attribute ========
	record_sender := "test user 1"
	record_receiver := "test user 2"
	record_transfer_amount := 100.0
	record_transfer_time := time.Now().Format("2006-01-02 15:04:05") // time to string format
	record_transfer_type := 1
	// string to time format : t, _ := time.Parse("2006-01-02 15:04:05", "2014-06-15 08:37:18")

	// ======== create record object and json byte ========
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

	// ======== create composite key to record ========
	// ======== save record to state ==================
	indexName := "sender~receiver~transfer_time"
	indexKey, err := stub.CreateCompositeKey(indexName, []string{record.Sender, record.Receiver, record.TransferTime})

	if err != nil {
		return shim.Error(err.Error())
	}

	stub.PutState(indexKey, recordJSONasBytes)

	fmt.Println("- end init record")
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "queryRecord" {
		return s.queryRecord(APIstub, args)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "createRecord" {
		return s.createRecord(APIstub, args)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	recordAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(recordAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) createRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var record = Record{FromPos: args[1], ToPos: args[2], Amount: args[3], TransferTime: args[4]}

	recordAsBytes, _ := json.Marshal(record)
	APIstub.PutState(args[0], recordAsBytes)

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
