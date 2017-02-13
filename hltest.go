package main

import (
	"errors"
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"encoding/json"
	"strconv"
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
const   STATE_WAREHOUSE = "0"
const   STATE_TRUCK = "1"
const   STATE_LOCAL_DEPO = "2"
const   STATE_LOCAL_DELIVERY = "3"
const   STATE_CUSTOMER = "4"

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
	ID          string `json:"id"`
	Product     string `json:"product"`
	TempLimit   string `json:"templimit"`
	HumLimit    string `json:"humlimit"`
	State       string `json:"state"`
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
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface, function string,args []string) ([]byte, error) {
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
	} else if function == "warehouse_to_truck" {
		return t.warehouse_to_truck(stub, args)
	} else if function == "truck_to_local_depo" {
		return t.truck_to_local_depo(stub, args)
	} else if function == "local_depo_to_local_delivery" {
		return t.local_depo_to_local_delivery(stub, args)
	} else if function == "local_delivery_to_customer" {
		return t.local_delivery_to_customer(stub, args)
	}

	fmt.Println("invoke did not find func: " + function)

	return nil, errors.New("Received unknown function invocation: " + function)
}

// Query is our entry point for queries
func (t *SimpleChaincode) Query(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	fmt.Println("query is running " + function)

	// Handle different functions
	if function == "dummy_query" {
		fmt.Println("hi there " + function)
		return nil, nil
	} else if function == "read" {
		return t.read(stub, args)
	}
	fmt.Println("query did not find func: " + function)

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
	var id, product, templimit, humlimit, state, str string
	var err error

	if len(args) != 4 {
		return nil, errors.New("Incorrect number of arguments. Expecting 4")
	}

	//input sanitation
	fmt.Println("- start create baggage")
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
	id = args[0]
	product = args[1]
	templimit = args[2]
	humlimit = args[3]
	state = "0"

	//check if baggage already exists
	baggageAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get baggage name")
	}
	res := Baggage{}
	json.Unmarshal(baggageAsBytes, &res)
	if res.ID == id {
		fmt.Println("This baggage arleady exists: " + id)
		fmt.Println(res);
		//all stop a baggage by this ID exists
		return nil, errors.New("This baggage arleady exists")
	}

	//build the baggage json string manually
	str = `{"id": "` + id + `", "product": "` + product + `", "templimit": "` + templimit + `", "humlimit": "` + humlimit + `", "state": "` + state + `"}`
	//store baggage with ID as key
	err = stub.PutState(id, []byte(str))
	if err != nil {
		return nil, err
	}

	//get the baggage index
	baggagesAsBytes, err := stub.GetState(BAGGAGE_INDEX_STR)
	if err != nil {
		return nil, errors.New("Failed to get baggage index")
	}
	var baggageIndex []string
	// un stringify it aka JSON.parse()
	json.Unmarshal(baggagesAsBytes, &baggageIndex)

	// add marble name to index list
	baggageIndex = append(baggageIndex, id)
	fmt.Println("! baggage index: ", baggageIndex)
	jsonAsBytes, _ := json.Marshal(baggageIndex)
	// store name of baggage
	err = stub.PutState(BAGGAGE_INDEX_STR, jsonAsBytes)

	fmt.Println("- end init baggage")
	return []byte(args[0]), nil
}


//=============================================================================================================================
//	 Delete Baggage - IDを指定して荷物を削除する
//=============================================================================================================================
func (t *SimpleChaincode) delete_baggage(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var err error

	// 引数の数をチェック
	if len(args) != 1 {
		return nil, errors.New("Incorrect number of arguments. Expecting 1")
	}
	// 引数が空でないことをチェック
	if len(args[0]) <= 0 {
		return nil, errors.New("1st argument must be a non-empty string")
	}

	err = stub.DelState(args[0])
	if err != nil {
		return nil, errors.New("Fail to Delete Baggage")
	}

	return nil, nil
}
//=============================================================================================================================
//	 Clear Baggage - Blockchain上に保持されている全ての荷物情報を削除する
//=============================================================================================================================
func (t *SimpleChaincode) clear_baggage(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {

}


// ============================================================================================================================
// Warehouse to Truck - 倉庫からトラックに荷物を引き渡す
// ============================================================================================================================
func (t *SimpleChaincode) warehouse_to_truck(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	args = append(args, STATE_WAREHOUSE)
	args = append(args, STATE_TRUCK)
  return t.change_state(stub, args)
}

// ============================================================================================================================
// Truck to Local Depo - トラックから地元の倉庫に荷物を引き渡す
// ============================================================================================================================
func (t *SimpleChaincode) truck_to_local_depo(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	args = append(args, STATE_TRUCK)
	args = append(args, STATE_LOCAL_DEPO)
  return t.change_state(stub, args)
}

// ============================================================================================================================
// Local Depo to Local Delivery - 地元の倉庫から地元の配送業者に荷物を引き渡す
// ============================================================================================================================
func (t *SimpleChaincode) local_depo_to_local_delivery(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	args = append(args, STATE_LOCAL_DEPO)
	args = append(args, STATE_LOCAL_DELIVERY)
  return t.change_state(stub, args)
}

// ============================================================================================================================
// Local Delivery to Customer - 地元の配送業者から顧客に荷物を引き渡す
// ============================================================================================================================
func (t *SimpleChaincode) local_delivery_to_customer(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	args = append(args, STATE_LOCAL_DELIVERY)
	args = append(args, STATE_CUSTOMER)
  return t.change_state(stub, args)
}

// ============================================================================================================================
// Change State - 荷物の状態を更新する
// ============================================================================================================================
func (t *SimpleChaincode) change_state(stub shim.ChaincodeStubInterface, args []string) ([]byte, error) {
	var id, prestate, poststate string
	// var id, poststate string
  var TempLimitVal, HumLimitVal, tempVal, humVal int
	// var tempVal, humVal int
	var err error

	if len(args) != 5 {
		return nil, errors.New("Incorrect number of arguments. Expecting 5")
	}

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
	if len(args[4]) <= 0 {
		return nil, errors.New("5th argument must be a non-empty string")
	}
	id = args[0]
	tempVal, err = strconv.Atoi(args[1])
	humVal, err = strconv.Atoi(args[2])
	prestate = args[3]
	poststate = args[4]
	// 存在チェック
	// 指定されたIDが存在しない場合にエラー
	// ======未実装======

	// baggage情報の取り出し
	baggageAsBytes, err := stub.GetState(id)
	if err != nil {
		return nil, errors.New("Failed to get baggage info")
	}
	res := Baggage{}
	json.Unmarshal(baggageAsBytes, &res)

	// 現在の状態をチェック

	if res.State != prestate {
		return nil, errors.New("This baggage can not be accepted")
	}

	// 温度と湿度のチェック
	TempLimitVal, err = strconv.Atoi(res.TempLimit)
	if err != nil {
		return nil, errors.New("Expecting integer value")
	}

	HumLimitVal, err = strconv.Atoi(res.HumLimit)
	if err != nil {
		return nil, errors.New("Expecting integer value")
	}

	if TempLimitVal < tempVal {
		return nil, errors.New("Temp Over")
	}

	if HumLimitVal < humVal {
		return nil, errors.New("Hum Over")
	}

	// 状態を更新
	res.State = poststate
	// 台帳への書き込み
	jsonAsBytes, _ := json.Marshal(res)
	err = stub.PutState(id, jsonAsBytes)

	return nil, nil

}
