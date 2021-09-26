package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"time"
)

type Node struct {
	Company   string      `json:"company"`
	Operation Operation   `json:"Operation"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

type Operation struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	Id        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

func Comp_and_Id_IsValid(a Node) (interface{}, bool) {
	if a.Company == "" {
		return nil, false // company name is absent
	}

	var id interface{}
	if a.Id == nil && a.Operation.Id == nil {
		return nil, false // id is absent
	} else {
		if a.Id != nil {
			id = a.Id
		} else {
			id = a.Operation.Id
		}
		_, ok1 := id.(float64)
		_, ok2 := id.(string)
		if ok1 || ok2 {
			return id, true // return id if id is valid
		} else {
			return nil, false // id is invalid
		}
	}
}
func IsConvertable(a string) (int, bool) {
	floatVal, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return 0, false
	}
	round := math.Ceil(floatVal)
	if floatVal != round {
		return 0, false // не конвертируется без потерь в целочисленное значение
	}
	return int(round), true
}

func ValueIsValid(a Node) (int, bool) {
	var val interface{}
	if a.Value == nil && a.Operation.Value == nil {
		return 0, false // value is absent
	}
	if a.Value != nil {
		val = a.Value
	} else {
		val = a.Operation.Value
	}
	floatVal, ok1 := val.(float64)
	strVal, ok2 := val.(string)
	if !ok1 && !ok2 {
		return 0, false // не является ни числом ни строкой
	}
	if ok2 { // если строка, пытаемся конвертировать в число
		res, ok := IsConvertable(strVal)
		if !ok {
			return 0, false // не конвертируется без потерь в целочисленное значение или не является числом
		}
		return res, true
	}
	return int(floatVal), true
}

func TimeIsValid(a Node) bool {
	var createdAt string
	retval := false
	if a.CreatedAt == "" && a.Operation.CreatedAt == "" {
		return false
	}
	if a.CreatedAt != "" {
		createdAt = a.CreatedAt
	} else {
		createdAt = a.Operation.CreatedAt
	}
	_, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return retval
	}
	return true
}

func TypeIsValid(a Node) (int, bool) {
	var val string
	if a.Type == "" && a.Operation.Type == "" {
		return 0, false // Value is absent
	}
	if a.Type != "" {
		val = a.Type
	} else {
		val = a.Operation.Type
	}
	if val == "+" || val == "income" {
		return 1, true
	} else if val == "-" || val == "outcome" {
		return -1, true
	}
	return 0, false
}

type OutNode struct {
	Company string        `json:"company"`
	ValidOp int           `json:"valid_operations_count"`
	Balance int           `json:"balance"`
	Id      []interface{} `json:"invalid_operations"`
}

func main() {
	var input []Node
	nodeMap := map[string]OutNode{}

	f2, _ := os.Open("billing.json")
	data2, _ := ioutil.ReadAll(f2)
	err := json.Unmarshal(data2, &input)
	if err != nil {
		fmt.Println(err)
	}
	_ = f2.Close()

	for _, val := range input {
		id, ok := Comp_and_Id_IsValid(val)
		if !ok || !TimeIsValid(val) { //  Пропускаем элемент, если отсутствуют:
			continue //  время, название компании или Id
		}
		value, okVal := ValueIsValid(val)

		operationType, okType := TypeIsValid(val)

		_, exists := nodeMap[val.Company]
		if exists {
			curr := nodeMap[val.Company]
			if !okVal || !okType {
				curr.Id = append(curr.Id, id)
				nodeMap[val.Company] = curr
			} else {
				curr.ValidOp += 1
				curr.Balance += (operationType) * value
				nodeMap[val.Company] = curr
			}
		} else {
			nodeMap[val.Company] = OutNode{Company: val.Company, ValidOp: 1, Balance: (operationType) * value}
		}
	}

	var nodeSlice []OutNode
	for _, v := range nodeMap {
		nodeSlice = append(nodeSlice, v)
	}

	f, _ := os.Create("out.json") // create file
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	_ = enc.Encode(nodeSlice)
}
