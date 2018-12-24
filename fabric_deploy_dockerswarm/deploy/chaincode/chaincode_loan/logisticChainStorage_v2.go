package main


import (
	"fmt"
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"encoding/json"
	"time"
)


type Loandata struct {
	//channelId + loanId as key
	ChannelId  string  	`json:"channelId"`
	LoanId     string	`json:"loanId"`
	//others as value, identityNo and phoneNo as index
	IdentityNo string	`json:"identityNo, omitempty"`
	PhoneNo    string	`json:"phoneNo, omitempty"`
	ExtraData  string	`json:"extraData, omitempty"`

}

type queryByChannelIdAndLoanIdData struct {
	ChannelId string  	`json:"channelId"`
	LoanId    string	`json:"loanId"`
}

type queryData struct {
	IdentityNo string	`json:"identityNo"`
	PhoneNo    string	`json:"phoneNo"`
	ChannelId  string	`json:"channelId"`
}


const (
	identityPhoneChannelIndex = "identityPhoneChannelIndex"
	identityIndex = "identityIndex"
	phoneIndex = "phoneIndex"
	channelIdIndex = "channelIdIndex"
	channelIdentityIndex = "channelIentityIndex"
	channelPhoneNoIndex = "channelPhoneNoIndex"
	identityPhoneNoIndex = "identityPhoneNoIndex"
)

type logisticChainStorage struct {
}


var logger = shim.NewLogger("logisticStorageLogger")


// ===================================================================================
// Util Functions
// ===================================================================================

func(t *logisticChainStorage) checkItemExistence(stub shim.ChaincodeStubInterface, key string, shouldExist bool) ([] byte, string) {
	// ==== Check if loan  exists ====
	loanBytes, err := stub.GetState(key)
	if err != nil {
		return nil, "Failed to get loan: " + err.Error()
	} else if loanBytes == nil && shouldExist == true {
		fmt.Println("This loan does not exist: " + key)
		return nil, "This loan does not exist: " + key
	} else if loanBytes != nil && shouldExist == false {
		fmt.Println("This loan already exists: " + key)
		return nil, "This loan already exists: " + key
	} else {
		return loanBytes, ""
	}
}

func (t *logisticChainStorage) buildIndex(stub shim.ChaincodeStubInterface, indexName string, indexAttributes []string) string {
	indexKey, err := stub.CreateCompositeKey(indexName, indexAttributes)
	if err != nil {
		return err.Error()
	}
	//  Save index entry to state. Only the key name is needed, no need to store a duplicate copy of the loan.
	//  Note - passing a 'nil' value will effectively delete the key from state, therefore we pass null character as value
	value := []byte{0x00}
	stub.PutState(indexKey, value)
	return ""

}


func(t *logisticChainStorage) updateIndex(stub shim.ChaincodeStubInterface, indexName string, oldIndexAttributes []string, newIndexAttributes []string) string {
	indexKey, err := stub.CreateCompositeKey(indexName, oldIndexAttributes)
	if err != nil {
		return err.Error()
	}
	//  Delete index entry to state.
	err = stub.DelState(indexKey)
	if err != nil {
		return "Failed to delete state:" + err.Error()
	}
	indexKey, err = stub.CreateCompositeKey(indexName, newIndexAttributes)
	if err != nil {
		return err.Error()
	}
	value := []byte{0x00}
	stub.PutState(indexKey, value)
	return ""

}


func (t *logisticChainStorage) deleteIndex(stub shim.ChaincodeStubInterface, indexName string, indexAttributes []string) string {
	indexKey, err := stub.CreateCompositeKey(indexName, indexAttributes)

	if err != nil {
		return err.Error()
	}
	//  Delete index entry to state.
	err = stub.DelState(indexKey)
	if err != nil {
		return "Failed to delete state:" + err.Error()
	}
	return ""
}


func (t *logisticChainStorage) updateLoanItems(oldDataItem *Loandata, newDataItem *Loandata ) {
	if newDataItem.IdentityNo != "" {
		oldDataItem.IdentityNo = newDataItem.IdentityNo
	}
	if newDataItem.PhoneNo != "" {
		oldDataItem.PhoneNo = newDataItem.PhoneNo
	}
	if newDataItem.ExtraData != "" {
		oldDataItem.ExtraData = newDataItem.ExtraData
	}
}

func (t *logisticChainStorage) indexQueryHandler(stub shim.ChaincodeStubInterface, indexName string, queryKey []string) ([]byte, string) {

	resultsIterator, err := stub.GetStateByPartialCompositeKey(indexName, queryKey)
	if err != nil {
		return nil, err.Error()
	}
	defer resultsIterator.Close()

	var dataItem Loandata
	var returnedResponse [] Loandata

	for resultsIterator.HasNext() {

		responseRange, err := resultsIterator.Next()
		if err != nil {
			return nil, err.Error()
		}
		// get the attributes from the composite key
		indexName, compositeKeyParts, err := stub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err.Error()
		}
		returnedKey := compositeKeyParts[len(compositeKeyParts) - 1]
		logger.Infof("- found a loan from index:%s key:%s\n", indexName, returnedKey)
		//fmt.Printf("- found a loan from index:%s key:%s\n", indexName, returnedKey)

		loanBytes, err := stub.GetState(returnedKey)

		if err != nil {
			jsonResp := "{\"Error\":\"Failed to get value for key: " + returnedKey + "\"}"
			return nil, jsonResp
		}
		err = json.Unmarshal(loanBytes, &dataItem)
		logger.Infof("- found a loan %v+", dataItem)
		//fmt.Printf("- found a loan %v+", dataItem)
		if err != nil {
			return nil, "Error unmarshal returned bytes"
		}
		//add one item to the result
		returnedResponse = append(returnedResponse, dataItem)
	}


	result, err := json.Marshal(returnedResponse)

	if err != nil {
		logger.Errorf("json err:", err)
		//fmt.Println("json err:", err)
	}

	return result, ""
}



// ===================================================================================
// Main
// ===================================================================================
func main() {
	err := shim.Start(new(logisticChainStorage))
	logger.SetLevel(shim.LogInfo)
	if err != nil {
		//fmt.Printf("Error starting logisticChainStorage chaincode: %s", err)
		logger.Errorf("Error starting logisticChainStorage chaincode: %s", err)
	}
}


// ===================================================================================
// Init
// ===================================================================================
func (t *logisticChainStorage) Init(stub shim.ChaincodeStubInterface) pb.Response {
	//TODO need to add authorization and authentication on later version
	timeStamp, err := stub.GetTxTimestamp()

	if err != nil {
		logger.Errorf("get tx timestamp error")
		return shim.Error("get tx timestamp error")
	}
	tm := time.Unix(timeStamp.GetSeconds(), int64(timeStamp.GetNanos()))
	logger.Infof(tm.Format("2006-01-02 03:04:05 PM") + " ########### LogisticChainStorage chaincode Init ###########")
	//fmt.Println(tm.Format("2006-01-02 03:04:05 PM") + " ########### LogisticChainStorage chaincode Init ###########")
	return shim.Success(nil)
}

// ===================================================================================
// Invoke Entry
// ===================================================================================

func (t *logisticChainStorage) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Infof("########### LogisticChainStorage chaincode Invoke ###########")
	//fmt.Println("########### LogisticChainStorage chaincode Invoke ###########")
	function, args := stub.GetFunctionAndParameters()

	if function != "invoke" {
		logger.Errorf("Unknown function call")
		return shim.Error("Unknown function call")
	}

	if len(args) != 2 {
		logger.Errorf("Incorrect number of arguments. Expecting 2")
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	method := args[0]

	timeStamp, err := stub.GetTxTimestamp()

	if err != nil {
		logger.Errorf("get tx timestamp error")
		return shim.Error("get tx timestamp error")
	}
	tm := time.Unix(timeStamp.GetSeconds(), int64(timeStamp.GetNanos()))


	logger.Infof(tm.Format("2006-01-02 03:04:05 PM") + " invoke is running " + method)
	logger.Infof(tm.Format("2006-01-02 03:04:05 PM") + " argument is :%+v\n", args[1])
	//fmt.Println(tm.Format("2006-01-02 03:04:05 PM") + " invoke is running " + method)
	//fmt.Println(tm.Format("2006-01-02 03:04:05 PM") + " argument is :%+v\n", args[1])



	switch method {

	case "saveLoan":
		return t.saveLoan(stub, args[1])
	case "updateLoan":
		return t.updateLoan(stub, args[1])
	case "deleteLoan":
		return t.deleteLoan(stub, args[1])
	case "queryByChannelIdAndLoanId":
		return t.queryByChannelIdAndLoanId(stub, args[1])
	case "queryByKeywords":
		return t.queryByKeywords(stub, args[1])
	case "getKeyHistory":
		return t.getKeyHistory(stub, args[1])
	}
	logger.Errorf("Received unknown function invocation")
	return shim.Error("Received unknown function invocation")
}


// ===================================================================================
// add, update, delete
// ===================================================================================

func (t *logisticChainStorage) saveLoan(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload Loandata


	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse payload")
		return shim.Error("Error parse payload")
	}

	key := payload.ChannelId + payload.LoanId

	logger.Infof("- payload is :%+v\n", payload)

	logger.Infof("- key is :%+v\n", key)

	//fmt.Printf("- payload is :%+v\n", payload)

	//fmt.Printf("- key is :%+v\n", key)

	_, errMsg := t.checkItemExistence(stub, key, false)

	if errMsg != "" {
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}

	// === Save loan to state ===
	err = stub.PutState(key, []byte(params))
	if err != nil {
		logger.Errorf(err.Error())
		return shim.Error(err.Error())
	}

	//  ==== Index the loan to enable identityNo or phoneNo based range queries,
	//  An 'index' is a normal key/value entry in state.
	//  The key is a composite key, with the elements that you want to range query on listed first.
	//  In our case, the composite key is based on indexName~color~name.
	//  This will enable very efficient state range queries based on composite keys matching indexName~IndentityNo~PhoneNo~ChannelId*
	errMsg = t.buildIndex(stub, identityIndex, []string{payload.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + identityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, phoneIndex, []string{payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + phoneIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, channelIdIndex, []string{payload.ChannelId, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + channelIdIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, channelIdentityIndex, []string{payload.ChannelId, payload.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + channelIdentityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, channelPhoneNoIndex, []string{payload.ChannelId, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + channelPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, identityPhoneNoIndex, []string{payload.IdentityNo, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + identityPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.buildIndex(stub, identityPhoneChannelIndex, []string{payload.ChannelId, payload.IdentityNo, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Building Index: " + identityPhoneChannelIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}

	// ==== loan saved and indexed. Return success ====
	//fmt.Println("- end save loan")
	logger.Infof("- end save loan")
	return shim.Success(nil)

}



func (t *logisticChainStorage) updateLoan(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload Loandata
	var dataItem Loandata
	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse payload")
		return shim.Error("Error parse payload")
	}
	key := payload.ChannelId + payload.LoanId

	logger.Infof("- payload is :%+v\n", payload)
	//fmt.Printf("- payload is :%+v\n", payload)
	logger.Infof("- key is :%+v\n", key)
	//fmt.Printf("- key is :%+v\n", key)

	loanBytes, errMsg := t.checkItemExistence(stub, key, true)

	if errMsg != "" {
		return shim.Error(errMsg)
	}
	err = json.Unmarshal([]byte(loanBytes), &dataItem)

	if err != nil {
		logger.Errorf("Error parse dataItem")
		return shim.Error("Error parse dataItem")
	}

	//make a copy for index update
	oldData := Loandata{
		ChannelId: dataItem.ChannelId,
		LoanId: dataItem.LoanId,
		IdentityNo: dataItem.IdentityNo,
		PhoneNo: dataItem.PhoneNo,
		ExtraData: dataItem.ExtraData,
	}


	t.updateLoanItems(&dataItem, &payload)

	// === Save loan to state ===
	dataBytes, err := json.Marshal(dataItem)

	if err != nil {
		logger.Errorf("Error marshaling dataItem")
		return shim.Error("Error marshaling dataItem")
	}
	err = stub.PutState(key, dataBytes)
	if err != nil {
		logger.Error(err.Error())
		return shim.Error(err.Error())
	}
	//update  the index
	errMsg = t.updateIndex(stub, identityIndex, []string{oldData.IdentityNo, key}, []string{dataItem.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + identityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, phoneIndex, []string{oldData.PhoneNo, key}, []string{dataItem.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + phoneIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, channelIdIndex, []string{oldData.ChannelId, key}, []string{dataItem.ChannelId, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + channelIdIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, channelIdentityIndex, []string{oldData.ChannelId, oldData.IdentityNo, key}, []string{dataItem.ChannelId, dataItem.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + channelIdentityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, channelPhoneNoIndex, []string{oldData.ChannelId, oldData.PhoneNo, key}, []string{dataItem.ChannelId, dataItem.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + channelPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, identityPhoneNoIndex, []string{oldData.IdentityNo, oldData.PhoneNo, key}, []string{dataItem.IdentityNo, dataItem.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + identityPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.updateIndex(stub, identityPhoneChannelIndex, []string{oldData.ChannelId, oldData.IdentityNo, oldData.PhoneNo, key},
		                                       []string{dataItem.ChannelId, dataItem.IdentityNo,  dataItem.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error Update Index: " + identityPhoneChannelIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}

	// ==== loan updated. Return success ====
	logger.Infof("- end update loan")
	//fmt.Println("- end update loan")
	return shim.Success(nil)
}



func (t *logisticChainStorage) deleteLoan(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload Loandata
	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse payload")
		return shim.Error("Error parse payload")
	}
	key := payload.ChannelId + payload.LoanId

	logger.Infof("- payload is :%+v\n", payload)
	//fmt.Printf("- payload is :%+v\n", payload)
	logger.Infof("- key is :%+v\n", key)
	//fmt.Printf("- key is :%+v\n", key)

	_, errMsg:= t.checkItemExistence(stub, key, true)

	if errMsg != "" {
		logger.Errorf(errMsg)
		return shim.Error(errMsg)
	}
	err = stub.DelState(key) //remove the loan from the state
	if err != nil {
		logger.Errorf("Failed to delete state:" + err.Error())
		return shim.Error("Failed to delete state:" + err.Error())
	}

	// delete the index
	errMsg = t.deleteIndex(stub, identityIndex, []string{payload.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + identityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, phoneIndex, []string{payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + phoneIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, channelIdIndex, []string{payload.ChannelId, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + channelIdIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, channelIdentityIndex, []string{payload.ChannelId, payload.IdentityNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + channelIdentityIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, channelPhoneNoIndex, []string{payload.ChannelId, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + channelPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, identityPhoneNoIndex, []string{payload.IdentityNo, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + identityPhoneNoIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	errMsg = t.deleteIndex(stub, identityPhoneChannelIndex, []string{payload.ChannelId, payload.IdentityNo, payload.PhoneNo, key})
	if errMsg != "" {
		logger.Errorf("error delete Index: " + identityPhoneChannelIndex + " errMsg: " + errMsg)
		return shim.Error(errMsg)
	}
	logger.Infof("- end delete loan")
	return shim.Success(nil)

}



// ===================================================================================
// Query by Key
// ===================================================================================

func (t *logisticChainStorage) queryByChannelIdAndLoanId(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload queryByChannelIdAndLoanIdData
	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse payload")
		return shim.Error("Error parse payload")
	}
	key := payload.ChannelId + payload.LoanId

	logger.Infof("- payload is :%+v\n", payload)
	//fmt.Printf("- payload is :%+v\n", payload)
	logger.Infof("- key is :%+v\n", key)
	//fmt.Printf("- key is :%+v\n", key)

	loanBytes, err := stub.GetState(key)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get value for channelId: " + payload.ChannelId + " and loanId:" + payload.LoanId + "\"}"
		logger.Errorf(jsonResp)
		return shim.Error(jsonResp)
	} else if loanBytes == nil {
		jsonResp := "{\"Error\":\"Nil state for " + payload.ChannelId + " and loanId:" + payload.LoanId + "\"}"
		logger.Errorf(jsonResp)
		return shim.Error(jsonResp)
	}
	logger.Infof("queryResult is " + string(loanBytes))
	return shim.Success(loanBytes)

}

// ===================================================================================
// Query functions
// ===================================================================================



func (t *logisticChainStorage) queryByKeywords(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload queryData
	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse queryData")
		return shim.Error("Error parse queryData")
	}
	logger.Infof("- Query payload is :%+v\n", payload)
	//fmt.Printf("- payload is :%+v\n", payload)
	var queryKeyword []string
	if payload.ChannelId != "" {
		queryKeyword = append(queryKeyword, payload.ChannelId)
		if payload.IdentityNo != "" {
			queryKeyword = append(queryKeyword, payload.IdentityNo)
			if payload.PhoneNo != "" {
				queryKeyword = append(queryKeyword, payload.PhoneNo)
				return t.queryByIdentityNoAndPhoneNoAndChannelId(stub, queryKeyword)
			}
			return t.queryByChannelIdAndIdentity(stub, queryKeyword)
		} else if payload.PhoneNo != "" {
			queryKeyword = append(queryKeyword, payload.PhoneNo)
			return t.queryByChannelIdAndPhoneNo(stub, queryKeyword)
		}
		return t.queryByChannelId(stub, queryKeyword)
	} else if payload.IdentityNo != "" {
		queryKeyword = append(queryKeyword, payload.IdentityNo)
		if payload.PhoneNo != "" {
			queryKeyword = append(queryKeyword, payload.PhoneNo)
			return t.queryByIdentityAndPhoneNo(stub, queryKeyword)
		}
		return t.queryByIdentity(stub, queryKeyword)
	} else if payload.PhoneNo != "" {
		queryKeyword = append(queryKeyword, payload.PhoneNo)
		return t.queryByPhoneNo(stub, queryKeyword)
	}

	return shim.Success(nil)

}



func (t *logisticChainStorage) queryByIdentityNoAndPhoneNoAndChannelId(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {

	logger.Infof("- querykey is :%+v\n", queryKeyword)
	//fmt.Printf("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, identityPhoneChannelIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)

}

func (t *logisticChainStorage) queryByIdentity(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {
	logger.Infof("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, identityIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)

}

func (t *logisticChainStorage) queryByPhoneNo(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {

	logger.Infof("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, phoneIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)

}

func (t *logisticChainStorage) queryByChannelId(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {

	logger.Infof("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, channelIdIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)
}

func (t *logisticChainStorage) queryByIdentityAndPhoneNo(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {
	logger.Infof("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, identityPhoneNoIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)
}

func (t *logisticChainStorage) queryByChannelIdAndIdentity(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {
	logger.Infof("- querykey is :%+v\n", queryKeyword)
	result, errMsg := t.indexQueryHandler(stub, channelIdentityIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)
}

func (t *logisticChainStorage) queryByChannelIdAndPhoneNo(stub shim.ChaincodeStubInterface, queryKeyword []string) pb.Response {
	logger.Infof("- querykey is :%+v\n", queryKeyword)

	result, errMsg := t.indexQueryHandler(stub, channelPhoneNoIndex, queryKeyword)

	if errMsg != "" {
		logger.Errorf("query error : " + errMsg)
		return shim.Error(errMsg)
	}

	return shim.Success(result)
}



func (t *logisticChainStorage) getKeyHistory(stub shim.ChaincodeStubInterface, params string) pb.Response {
	var payload queryByChannelIdAndLoanIdData
	err := json.Unmarshal([]byte(params), &payload)

	if err != nil {
		logger.Errorf("Error parse payload")
		return shim.Error("Error parse payload")
	}

	if payload.ChannelId == "" {
		logger.Errorf("Missing ChanelId")
		return shim.Error("Missing ChannelId")
	} else if payload.LoanId == "" {
		logger.Errorf("Missing LoanId")
		return shim.Error("Missing LoanId")
	}

	key := payload.ChannelId + payload.LoanId

	logger.Infof("- payload is :%+v\n", payload)
	//fmt.Printf("- payload is :%+v\n", payload)
	logger.Infof("- key is :%+v\n", key)
	//fmt.Printf("- key is :%+v\n", key)

	keysIter, err := stub.GetHistoryForKey(key)
	if err != nil {
		logger.Errorf(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
		return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
	}
	defer keysIter.Close()
	var keys []string
	for keysIter.HasNext() {
		response, iterErr := keysIter.Next()
		if iterErr != nil {
			logger.Errorf(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
			return shim.Error(fmt.Sprintf("query operation failed. Error accessing state: %s", err))
		}
		keys = append(keys, response.TxId)
	}
	/*
	for key, txID := range keys {
		fmt.Printf("key %d contains %s\n", key, txID)
	}
	*/

	jsonKeys, err := json.Marshal(keys)
	if err != nil {
		logger.Errorf(fmt.Sprintf("query operation failed. Error marshaling JSON: %s", err))
		return shim.Error(fmt.Sprintf("query operation failed. Error marshaling JSON: %s", err))
	}

	return shim.Success(jsonKeys)
}
