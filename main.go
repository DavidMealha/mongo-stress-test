package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
	"net/http"
	"bytes"
	"github.com/DavidMealha/mongo-stress-test/users"
	"gopkg.in/mgo.v2"
	"io/ioutil"
)

var (
	wg sync.WaitGroup
	letters = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func startClient(clientNr int) {
	for i := 0; i < 10; i++ {
		rand := rand.Intn(2)
		if rand == 0 {
			//fmt.Printf("Do write in Client %v =>", clientNr)
			//fmt.Print(time.Now())
			//fmt.Println()
			insertUser()
		} else {
			//fmt.Printf("Do read in Client %v =>", clientNr)
			//fmt.Print(time.Now())
			//fmt.Println()
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
		fmt.Println(err, body)
	}
	fmt.Println("INSERT =>", str, " - AT => ", time.Now().UnixNano())
	//fmt.Println("response Status:", resp.Status)
}

func getRandomString(size int) string {
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

func getAllRecords(dbAddress string) []users.User {
	session, err := mgo.Dial(dbAddress)
    if err != nil {
    	panic(err)
    }
    fmt.Println("Established connection to => ", dbAddress)

    defer session.Close()

    c := session.DB("users").C("customers")

    var results []users.User
    err = c.Find(nil).All(&results)

    if err != nil {
    	panic(err)
    } else {
		return results
	}
}

func verifyOrder() {
	fmt.Println("going to verify records")
	cloudRecords := getAllRecords("localhost:27018,localhost:27019,localhost:27020")
	edgeRecords	 := getAllRecords("localhost:27021")

	fmt.Println("Cloud length => ", len(cloudRecords))
	fmt.Println("Edge length => ", len(edgeRecords))
	

	//check if both lists have the same length
	//compare the value of each list position to see if they match
}

func main() {
	wg.Add(10)
	for i := 0; i < 5; i++ {
		go startClient(i)
	}
	fmt.Println("Main: Waiting for workers to finish")
	wg.Wait()

	verifyOrder();
    fmt.Println("Finished.")
}
