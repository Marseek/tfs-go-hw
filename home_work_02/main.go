package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
)

// Node - Структура для Анмаршаллинга
type Node struct {
	Company   string      `json:"company"`
	Operation Operation   `json:"Operation"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	ID        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

type Operation struct {
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
	ID        interface{} `json:"id"`
	CreatedAt string      `json:"created_at"`
}

// CompAndIDIsValid - Проверка наличия полей Company и ID
func CompAndIDIsValid(a Node) (interface{}, bool) {
	if a.Company == "" {
		return nil, false // company name is absent
	}

	var id interface{}
	if a.ID == nil && a.Operation.ID == nil {
		return nil, false // id is absent
	}
	if a.ID != nil {
		id = a.ID
	} else {
		id = a.Operation.ID
	}
	_, ok1 := id.(float64)
	_, ok2 := id.(string)
	if ok1 || ok2 {
		return id, true // return id if id is valid
	}
	return nil, false // id is invalid
}

// IsConvertable - конвертируется ли строка без потерь в целочисленное значение
func IsConvertable(a string) (int, bool) {
	floatVal, err := strconv.ParseFloat(a, 64)
	if err != nil {
		return 0, false
	}
	round := math.Ceil(floatVal)
	if floatVal != round {
		return 0, false
	}
	return int(round), true
}

// ValueIsValid - Проверка валидности поля Value
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
	// если строка, пытаемся конвертировать в число
	if ok2 {
		res, ok := IsConvertable(strVal)
		if !ok {
			return 0, false // не конвертируется без потерь в целочисленное значение или не является числом
		}
		return res, true
	}
	// если число, проверяем на наличие дробной части
	round := math.Ceil(floatVal)
	if floatVal != round {
		return 0, false
	}
	return int(floatVal), true
}

// TimeIsValid - Проверка наличия и валидности поля CreatedAt
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

// TypeIsValid - Проверка валидности поля Value
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
	}
	if val == "-" || val == "outcome" {
		return -1, true
	}
	return 0, false
}

// OutNode - Структура для вывода результата в JSON
type OutNode struct {
	Company string        `json:"company"`
	ValidOp int           `json:"valid_operations_count"`
	Balance int           `json:"balance"`
	ID      []interface{} `json:"invalid_operations"`
}

var filePath string

// Получение пути и названия файла
func init() {
	var filePath2 = flag.String("file", "", "path to file")
	flag.Parse()
	if *filePath2 != "" {
		filePath = *filePath2
		return
	}

	s, ok := os.LookupEnv("FILE")
	if ok {
		filePath = s
		return
	}

	fmt.Println("Enter path to file:")
	_, _ = fmt.Scan(&filePath)
}

func main() {
	var input []Node

	// Словарь для хранения промежуточных результатов (для удобного поиска по полю Company)
	nodeMap := map[string]OutNode{}

	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}
	data2, _ := ioutil.ReadAll(f)
	err = json.Unmarshal(data2, &input)
	if err != nil {
		fmt.Println("Произошла ошибка при Анмаршаллинге")
		panic(err)
	}
	_ = f.Close()

	// проход по считанным из файла данным
	for _, val := range input {
		id, ok := CompAndIDIsValid(val)
		if !ok || !TimeIsValid(val) { //  Пропускаем элемент, если отсутствуют:
			continue //  время, название компании или ID
		}
		value, okVal := ValueIsValid(val)         // получение данных из поля Value
		operationType, okType := TypeIsValid(val) // получение данных из поля Type

		_, exists := nodeMap[val.Company]
		if exists {
			curr := nodeMap[val.Company]
			if !okVal || !okType { // Если операция невалидна
				curr.ID = append(curr.ID, id) // Добавление ID в поле невалидных операций
				nodeMap[val.Company] = curr
			} else { // Если валидна - изменение данных балланса и счетчика операций
				curr.ValidOp++
				curr.Balance += (operationType) * value
				nodeMap[val.Company] = curr
			}
		} else { // Если компании с таким названием не существовало
			if !okVal || !okType {
				nodeMap[val.Company] = OutNode{Company: val.Company, ID: []interface{}{id}}
			} else {
				nodeMap[val.Company] = OutNode{Company: val.Company, ValidOp: 1, Balance: (operationType) * value}
			}
		}
	}
	var nodeSlice []OutNode // Копируем полученные данные в структуру для Анмаршаллинга
	for _, v := range nodeMap {
		nodeSlice = append(nodeSlice, v)
	}

	sort.SliceStable(nodeSlice, func(i, j int) bool { // Сортировка слайса структур по полю Company
		return nodeSlice[i].Company < nodeSlice[j].Company
	})

	for _, v := range nodeSlice {
		sort.Slice(v.ID, func(i, j int) bool { // Сортировка слайса с невалидными операциями
			id1, ok1 := v.ID[i].(string)
			id2, ok2 := v.ID[j].(string)
			if !ok1 {
				id1 = fmt.Sprintf("%f", v.ID[i].(float64))
			}
			if !ok2 {
				id2 = fmt.Sprintf("%f", v.ID[j].(float64))
			}
			return id1 < id2
		})
	}

	f2, _ := os.Create("out.json") // create file
	defer f2.Close()

	enc := json.NewEncoder(f2)
	enc.SetIndent("", "\t")
	_ = enc.Encode(nodeSlice)
}

