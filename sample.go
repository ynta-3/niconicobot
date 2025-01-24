package main

import (
	"encoding/json"
	"fmt"
	"log"
)

// PersonはJSONの各要素に対する構造体
type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func main() {
	jsonData := `[{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]`

	var people []Person
	if err := json.Unmarshal([]byte(jsonData), &people); err != nil {
		log.Fatal(err)
	}

	// 新しい要素をスライスに追加
	newPerson := Person{ID: 3, Name: "Charlie"}
	people = append(people, newPerson)

	// スライスの内容を確認
	for _, person := range people {
		fmt.Printf("ID: %d, Name: %s\n", person.ID, person.Name)
	}
}
