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
  writeLatencies []string
  readLatencies []string
)

const (
  DATABASE_ADDRESS  = "localhost:27019"
  DATABASE_NAME     = "users"
  COLLECTION_NAME   = "customers"
  PROXY_ADDRESS     = "http://localhost:8127"
  SERVICE_ADDRESS   = "http://localhost:8080/"
  WRITE_RATE        = 50
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

func startClient(clientNr int) {
  for i := 0; i < 10; i++ {
    rand := rand.Intn(100)
    if rand < WRITE_RATE {
      fmt.Printf("Do write in Client %v \n", clientNr)
      insertUser()
    } else {
      fmt.Printf("Do read in Client %v \n", clientNr)
      readUser()
    }
    time.Sleep(200 * time.Millisecond)
  }
  defer wg.Done()
}

func insertUser() {
  start := time.Now()

  str := `{"username":"` + getRandomString(8) + 
         `","password":"` + getRandomString(12) + 
         `","email":"` + getRandomString(10) + 
         `","firstName":"` + getRandomString(6) + 
         `","lastName":"` + getRandomString(8) + `"}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", SERVICE_ADDRESS + "customers", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  elapsed := time.Since(start)
  readLatencies = append(readLatencies, elapsed.String())
}

func readUser() {
  start := time.Now()
  resp, err := http.Get(SERVICE_ADDRESS + "customers/" + getRandomString(10))
  
  if err != nil {
    fmt.Println(err)
  }

  fmt.Println("Response:", resp.body())
  elapsed := time.Since(start)
  readLatencies = append(readLatencies, elapsed.String())
}

func insertUserFromWrapper() {
  start := time.Now()
  url := PROXY_ADDRESS
  str := `{"operationType":"INSERT",` + 
         `"fullDocument":{"name":"` + getRandomString(12) + 
         `","username":"` + getRandomString(20) +
         `"},"ns":{"coll":"` + COLLECTION_NAME +
         `","db":"` + DATABASE_NAME +
         `"},"documentKey":{"_id":"` + getRandomString(6) + `"}}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  //fmt.Println("INSERT =>", str, " - AT => ", time.Now().UnixNano())
  elapsed := time.Since(start)
  writeLatencies = append(writeLatencies, elapsed.String())
  //fmt.Println("Took ", elapsed, " ms.")
}

func readUserFromDatabase() {
  start := time.Now()
  session, err := mgo.Dial(DATABASE_ADDRESS)

  if err != nil {
    panic(err)
  }

  defer session.Close()

  c := session.DB(DATABASE_NAME).C(COLLECTION_NAME)

  var result users.User
  err = c.FindId(bson.M{ "_id": bson.ObjectIdHex("537f700b537461b70c5f0000") }).One(&result)

  if err != nil {
    fmt.Println("error retrieving record =>", err)
  } else {
    //fmt.Println("result =>", &result)
  }
  elapsed := time.Since(start)
  readLatencies = append(readLatencies, elapsed.String())
  //fmt.Println("Took ", elapsed, " ms.")
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

func getRandomUUID() string {
  return "537f700b537461b70c5f0000";
}

func main() {
  wg.Add(5)
  for i := 0; i < 5; i++ {
    go startClient(i)
  }
  wg.Wait()

  fmt.Printf("Write latencies %\n", writeLatencies)
  fmt.Printf("Read latencias %\n", readLatencies)

  fmt.Println("Finished.")
}
