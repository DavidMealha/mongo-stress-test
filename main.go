package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
	"net/http"
	"io/ioutil"
	"bytes"
)

var (
	wg sync.WaitGroup
	letters = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func startClient(clientNr int) {
	for i := 0; i < 100; i++ {
		rand := rand.Intn(2)
		if rand == 0 {
			fmt.Printf("Do write in Client %v =>", clientNr)
			fmt.Print(time.Now())
			fmt.Println()
			insertUser()
		} else {
			fmt.Printf("Do read in Client %v =>", clientNr)
			fmt.Print(time.Now())
			fmt.Println()
		}
		time.Sleep(1000 * time.Millisecond)
	}
	defer wg.Done()
}

func insertUser() {
	url := "http://localhost:8080/customers"
	str := `{"username":"` + getRandomString(8) + 
			 `","password":"` + getRandomString(12) + 
			 `","email":"` + getRandomString(10) + 
			 `","firstName":"` + getRandomString(6) + 
			 `","lastName":"` + getRandomString(8) + `"}`

 	fmt.Println("json =>", str)

	var jsonStr = []byte(str)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("response Status:", resp.Status)
}

func getRandomString(size int) string{
	lettersLen := len(letters)
	var str string
	for i := 0; i < size; i++ {
		rand := rand.Intn(lettersLen)
		str += letters[rand]
	}
	return str
}

func readUsers() {
	//resp, err := http.Get("localhost:8080/customers")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Response Status:", resp.Status())
}

func main() {
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go startClient(i)
	}
	wg.Wait()
}
