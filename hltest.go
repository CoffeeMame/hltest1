package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
//	"strconv"
//	"strings"
//	"regexp"
)

//==============================================================================================================================
//	 Participant types - Each participant type is mapped to an integer which we use to compare to the value stored in a
//						 user's eCert
//==============================================================================================================================
//CURRENT WORKAROUND USES ROLES CHANGE WHEN OWN USERS CAN BE CREATED SO THAT IT READ 1, 2, 3, 4, 5
const   WAREHOUSE = "warehouse"
const   TRUCK = "truck"
const   LOCAL_DEPO = "local_depo"
const   LOCAL_DELIVERY = "local_delivery"
const   CUSTOMER = "customer"


//==============================================================================================================================
//	 Status types - Asset lifecycle is broken down into 5 statuses, this is part of the business logic to determine what can
//					be done to the baggage at points in it's lifecycle
//==============================================================================================================================
const   STATE_WAREHOUSE 			=  0
const   STATE_TRUCK  			=  1
const   STATE_LOCAL_DEPO 	=  2
const   STATE_LOCAL_DELIVERY 			=  3
const   STATE_CUSTOMER 		=  4

//==============================================================================================================================
//name for the key/value that will store a list of all known baggage
//==============================================================================================================================
const BAGGAGE_INDEX_STR = "_baggageindex"

//==============================================================================================================================
//	 Structure Definitions
//==============================================================================================================================
//	Chaincode - A blank struct for use with Shim (A HyperLedger included go file used for get/put state
//				and other HyperLedger functions)
//==============================================================================================================================
type SimpleChaincode struct {
}

//==============================================================================================================================
//	Baggage - Defines the structure for a baggage object. JSON on right tells it what JSON fields to map to
//			  that element when reading a JSON object into the struct e.g. JSON make -> Struct Make.
//==============================================================================================================================
type Baggage struct {
	ID          string `json:"ID"`
	Product     string `json:"Product"`
	TempLimit   string `json:"TempLimit"`
	HumLimit    string `json:"HumLimit"`
	State       string `json:"State"`
}

//==============================================================================================================================
//	 General Functions
//==============================================================================================================================




//==============================================================================================================================
//	 Router Functions
//==============================================================================================================================
// ============================================================================================================================
// Main
// ============================================================================================================================
func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

// Init resets all the things
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}

	return nil, nil
}

// Invoke is our entry point to invoke a chaincode function
func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("invoke is running " + function)

	// Handle different functions
	if function == "create_baggage" {
		// Create new baggage
		return t.create_baggage(stub, args)
	} else if function == "read" {
		return t.read(stub, args)
	}

		/*
		return t.create_baggage(stub, args)
	} else if function == "baggage_confirmation" {
		return t.luggage_confirmation(stub)
	} else if function == "warehouse_to_truck" {
		return t.warehouse_to_truck(stub)
	} else if function == "truck_to_local_depo" {
		return t.truck_to_local_depo(stub)
	} else if function == "local_depo_to_local_delivery" {
		return t.local_depo_to_local_delivery(stub)
	} else if function == "local_delivery_to_customer"  {
		return t.local_delivery_to_customer(stub)
	}
	*/

	fmt.Println("invoke did not find func: " + function)					//error

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "dummy_query" {											//read a variable
		fmt.Println("hi there " + function)						//error
		return nil, nil;
	}
	fmt.Println("query did not find func: " + function)						//error

	return nil, errors.New("Received unknown function query: " + function)
}



// ============================================================================================================================
// Read - read a variable from chaincode state
// ============================================================================================================================
func (t *SimpleChaincode) read(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var id, jsonResp string
	var err error

	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting ID of the var to query")
	}

	id = args[0]
	valAsbytes, err := stub.GetState(id)									//get the var from chaincode state
	if err != nil {
		jsonResp = "{\"Error\":\"Failed to get state for " + id + "\"}"
		return nil, errors.New(jsonResp)
	}

	//send it onward
	return valAsbytes, nil
}


//=================================================================================================================================
//	 Create Function
//=================================================================================================================================
//	 Create Baggage - Creates the initial JSON for the baggage and then saves it to the ledger.
//=================================================================================================================================
func (t *SimpleChaincode) create_baggage(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	//   0       1       2     3
	// "asdf", "blue", "35", "bob"
	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start init marble")
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, errors.New("2nd argument must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, errors.New("3rd argument must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, errors.New("4th argument must be a non-empty string")
	}

	// Variables to define the JSON
	ID         := args[0]
	Product    := args[1]
	TempLimit  := args[2]
	HumLimit   := args[3]
	State      := "0"

	//check if baggage already exists
	baggageAsBytes, err := stub.GetState(ID)
	if err != nil {
		return nil, errors.New("Failed to get baggage name")
	}
	res := Baggage{}
	json.Unmarshal(baggageAsBytes, &res)
	if res.ID == ID{
		fmt.Println("This baggage arleady exists: " + ID)
		fmt.Println(res);
		//all stop a baggage by this ID exists
		return nil, errors.New("This baggage arleady exists")
	}

	//build the baggage json string manually
	str := `{"ID": "` + ID + `", "Product": "` + Product + `", "TempLimit": ` + TempLimit + `, "HumLimit": "` + HumLimit + `", "State": "` + State + `"}`
	//store baggage with ID as key
	err = stub.PutState(ID, []byte(str))
	if err != nil {
		return nil, err
	}

	//get the baggage index
	baggagesAsBytes, err := stub.GetState(BAGGAGE_INDEX_STR)
	if err != nil {
		return nil, errors.New("Failed to get baggage index")
	}
	var baggageIndex []string
	//un stringify it aka JSON.parse()
	json.Unmarshal(baggagesAsBytes, &baggageIndex)

	//add marble name to index list
	baggagesIndex = append(baggageIndex, ID)
	fmt.Println("! baggage index: ", baggageIndex)
	jsonAsBytes, _ := json.Marshal(baggageIndex)
	//store name of baggage
	err = stub.PutState(BAGGAGE_INDEX_STR, jsonAsBytes)

	fmt.Println("- end init baggage")
	return nil, nil
}
