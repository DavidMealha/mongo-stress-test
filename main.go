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
  "gopkg.in/mgo.v2/bson"
)

var (
	wg sync.WaitGroup
	letters = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
)

const (
  DATABASE_ADDRESS  = "localhost:27021"
  DATABASE_NAME     = "users"
  COLLECTION_NAME   = "customers"
  PROXY_ADDRESS     = "http://replication-manager:8126"
  WRITE_RATE        = 50
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func startClient(clientNr int) {
	for i := 0; i < 10; i++ {
		rand := rand.Intn(100)
		if rand < WRITE_RATE {
			fmt.Printf("Do write in Client %v =>", clientNr)
			insertUser()
		} else {
			fmt.Printf("Do read in Client %v =>", clientNr)
      readUser()
		}
		time.Sleep(200 * time.Millisecond)
	}
	defer wg.Done()
}

func insertUser() {
  url := "http://replication-manager:8126"
  str := `{"operationType":"INSERT",` + 
         `"fullDocument":{"name":"` + getRandomString(12) + 
         `","username":"` + getRandomString(14) +
         `"},"ns":{"coll":"` + COLLECTION_NAME +
         `","db":"` + DATABASE_NAME +
         `"},"documentKey":{"_id":"` + getRandomString(6) + `"}}`

  // str := `{"username":"` + getRandomString(8) + 
  //         `","password":"` + getRandomString(12) + 
  //         `","email":"` + getRandomString(10) + 
  //         `","firstName":"` + getRandomString(6) + 
  //         `","lastName":"` + getRandomString(8) + `"}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  //body, err := ioutil.ReadAll(resp.Body)
  //if err != nil {
  //	fmt.Println(err)
  //}
  fmt.Println("INSERT =>", str, " - AT => ", time.Now().UnixNano())
}

func readUser() {
  session, err := mgo.Dial(DATABASE_ADDRESS)

  if err != nil {
    panic(err)
  }
  fmt.Println("Established connection to => ", DATABASE_ADDRESS)

  defer session.Close()

  c := session.DB(DATABASE_NAME).C(COLLECTION_NAME)

  result := User{}
  err = c.FindId(bson.M{ "_id": bson.ObjectIdHex("") }).One(&result)

  if err != nil {
    panic(err)
  } else {
    return &result
  }
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
	// /customers/{id}
	
	//resp, err := http.Get("localhost:8080/customers")
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Println("Response Status:", resp.Status())
}

func getAllRecords(dbAddress string, collection string) []users.User {
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
  cloudRecords := getAllRecords("localhost:27018,localhost:27019,localhost:27020", "customers")
  edgeRecords	 := getAllRecords("localhost:27021", "customers")

  fmt.Println("Cloud length => ", len(cloudRecords))
  fmt.Println("Edge length => ", len(edgeRecords))

  arrayLength := len(edgeRecords)
  positiveMatch := 0
  negativeMatch := 0

  for i := 0; i < arrayLength; i++ {
    if (cloudRecords[i].Username == edgeRecords[i].Username) {
      positiveMatch += 1
    } else {
      negativeMatch += 1
      fmt.Printf("Cloud => %v | Edge => %v", string(cloudRecords[i].Username), edgeRecords[i].Username)
      fmt.Println(" | Negative Match on line => ", i)
    }
  }

  fmt.Println("Positive matches => ", positiveMatch)
  fmt.Println("Negative matches => ", negativeMatch)
}

func main() {
  wg.Add(5)
  for i := 0; i < 5; i++ {
    go startClient(i)
  }
  wg.Wait()

  fmt.Println("Waiting 60 seconds before checking order.")
  //time.Sleep(60000 * time.Millisecond)
  // verifyOrder();

  fmt.Println("Finished.")
}
