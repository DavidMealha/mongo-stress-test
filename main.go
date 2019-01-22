package main

import (
  "fmt"
  "flag"
  "time"
  "math/rand"
  "sync"
  "net/http"
  "bytes"
  "io/ioutil"
  "github.com/DavidMealha/mongo-stress-test/users"
  "gopkg.in/mgo.v2"
  "gopkg.in/mgo.v2/bson"
  "encoding/hex"
  "strconv"
)

var (
  wg sync.WaitGroup
  letters = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
  writeLatencies []int
  readLatencies []int
  ids = []string{"537f700b537461b70c5f0000"}
  writeRate int
)

const (
  DATABASE_ADDRESS  = "localhost:27017" // Cloud or Edge database address
  DATABASE_NAME     = "users"
  COLLECTION_NAME   = "customers"
  PROXY_ADDRESS     = "http://localhost:8126" // Cloud or Edge middleware address
  SERVICE_ADDRESS   = "http://localhost:8080/"
  //SERVICE_ADDRESS   = "http://3.120.161.145:8080/"
  //SERVICE_ADDRESS   = "http://52.213.179.93:8080/"
  WRITE_RATE        = 50
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

func startClient(client *http.Client, clientNr int, nrOperations int) {
  for i := 0; i < nrOperations; i++ {
    randRate := rand.Intn(100)

    if randRate < writeRate {
      // insertUser(client)
      randOp := rand.Intn(4)
      writeOperationToWrapper(client, randOp)
    } else {
      readUser()
      //readUserFromDatabase()
    }
  }
  defer wg.Done()
}

func insertUser(client *http.Client) {
  start := time.Now()
  //fmt.Printf("Client %v sent request at %v\n", clientNr, start)
  str := `{"username":"` + getRandomString(8) + 
         `","password":"` + getRandomString(12) + 
         `","email":"` + getRandomString(10) + 
         `","firstName":"` + getRandomString(6) + 
         `","lastName":"` + getRandomString(8) + `"}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", SERVICE_ADDRESS + "register", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Connection", "keep-alive")

  //client := &http.Client{}
  resp, err := client.Do(req)
  //fmt.Printf("Service response to client %v sent request at %v\n", clientNr, start)

  if err != nil {
    fmt.Println(err)
  }
  parsedBody,err := ioutil.ReadAll(resp.Body)
  defer resp.Body.Close()

  elapsed := int(time.Since(start) / time.Microsecond)
  writeLatencies = append(writeLatencies, elapsed)
  body := string(parsedBody)
  fmt.Println("Insert Response", body)
}

func readUser() {
  start := time.Now()
  randomId := ids[rand.Intn(len(ids))]
  resp, err := http.Get(SERVICE_ADDRESS + "customers/" + randomId)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  elapsed := int(time.Since(start) / time.Microsecond)

  readLatencies = append(readLatencies, elapsed)
  parsedBody, err := ioutil.ReadAll(resp.Body)
  body := string(parsedBody)
  fmt.Println("Response:", body)
}

func randomHex(n int) string {
  bytes := make([]byte, n)
  if _, err := rand.Read(bytes); err != nil {
    return "537f700b537461b70c5f0000"
  }
  return hex.EncodeToString(bytes)
}

func writeOperationToWrapper(client *http.Client, opType int) {
  start := time.Now()

  var payload string
  if opType == 0 { // insert
    payload = getInsertOperationInJson()
  } else if opType == 1 { // update
    payload = getUpdateOperationInJson()
  } else if opType == 2 { // replace
    payload = getReplaceOperationInJson()
  } else { // delete
    payload = getDeleteOperationInJson()
  }

  var jsonStr = []byte(payload)

  req, err := http.NewRequest("POST", PROXY_ADDRESS, bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Connection", "keep-alive")

  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
    return
  }
  parsedBody,err := ioutil.ReadAll(resp.Body)
  defer resp.Body.Close()

  elapsed := int(time.Since(start) / time.Microsecond)
  writeLatencies = append(writeLatencies, elapsed)
  
  body := string(parsedBody)
  fmt.Println("Operation =>", opType, "Response", body)
}

func getInsertOperationInJson() string {
  newId := randomHex(12)
  ids = append(ids, newId)

  return `{"operationType":"INSERT",` + 
        `"fullDocument":{"username":"` + getRandomString(12) + 
        `","password":"` + getRandomString(12) +
        `","email":"` + getRandomString(12) +
        `","firstName":"` + getRandomString(12) +
        `","lastName":"` + getRandomString(12) +
        `","origin_created_at":"` + strconv.FormatInt(time.Now().UnixNano()/int64(time.Microsecond),10) + 
        `"},"ns":{"coll":"` + COLLECTION_NAME +
        `","db":"` + DATABASE_NAME +
        `"},"documentKey":{"_id":"` + newId + `"}}`
}

func getUpdateOperationInJson() string {
  randomId := ids[rand.Intn(len(ids))]
  return `{"operationType":"UPDATE",` + 
        `"updatedFields":{"username":"` + getRandomString(12) + 
        `","email":"` + getRandomString(12) +
        `","firstName":"` + getRandomString(12) +        
        `"},"removedFields":["lastName"],` + 
        `"ns":{"coll":"` + COLLECTION_NAME +
        `","db":"` + DATABASE_NAME +
        `"},"documentKey":{"_id":"` + randomId + `"}}`
}

func getReplaceOperationInJson() string {
  randomId := ids[rand.Intn(len(ids))]
  return `{"operationType":"REPLACE",` + 
        `"fullDocument":{"username":"` + getRandomString(12) + 
        `","password":"` + getRandomString(12) +
        `","email":"` + getRandomString(12) +
        `","firstName":"` + getRandomString(12) +
        `","lastName":"` + getRandomString(12) +
        `"},"ns":{"coll":"` + COLLECTION_NAME +
        `","db":"` + DATABASE_NAME +
        `"},"documentKey":{"_id":"` + randomId + `"}}`
}

func getDeleteOperationInJson() string {
  randomId := ids[rand.Intn(len(ids))]
  return `{"operationType":"DELETE",` + 
        `"ns":{"coll":"` + COLLECTION_NAME +
        `","db":"` + DATABASE_NAME +
        `"},"documentKey":{"_id":"` + randomId + `"}}`
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
  randomId := ids[rand.Intn(len(ids))]
  err = c.FindId(bson.ObjectIdHex(randomId)).One(&result)

  if err != nil {
    fmt.Println("error retrieving record =>", err)
  } else {
    fmt.Println("result =>", &result.UserID, " - ", result.FirstName)
  }
  elapsed := int(time.Since(start) / time.Microsecond)
  readLatencies = append(readLatencies, elapsed)
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

func printStats(elapsed time.Duration, nrOperations int, nrClients int) {
  var writeRate = float64(writeRate) / 100;
  var readRate = 1 - writeRate;

  var avgWriteLatency = getAverage(writeLatencies) / 1000;
  var avgReadLatency = getAverage(readLatencies) / 1000;

  // var totalThroughput = float64(nrOperations)/float64(elapsed/1000)

  fmt.Println("=======================================");
  fmt.Println("============= STATISTICS ==============");
  fmt.Println("=======================================");
  fmt.Println("Write Rate =>\t", writeRate);
  fmt.Println("Read Rate =>\t", readRate);
  fmt.Println("=======================================");
  fmt.Println("Average Write Latency =>", avgWriteLatency, "ms")
  fmt.Println("Average Read Latency =>\t", avgReadLatency, "ms");
  fmt.Println("=======================================");
  fmt.Println("Took ", elapsed, " ms to perform ", (nrOperations * nrClients), " operations.")
  // fmt.Println("Total throughput", totalThroughput, " ops/sec")
  fmt.Println("=======================================");
}

func getAverage(array []int) int{
  var sum int;
  for i := 0; i < len(array); i++ {
    sum += array[i];
  }
  return sum / len(array)
}

func main() {
  //start time
  start := time.Now()

  var nrClients int;
  var nrOperations int;

  flag.IntVar(&nrClients, "clients", 5, "Number of clients");
  flag.IntVar(&nrOperations, "operations", 50, "Number of operations");
  flag.IntVar(&writeRate, "write-rate", 10, "Percentage of write operations");
  flag.Parse();

  wg.Add(nrClients)
  for i := 0; i < nrClients; i++ {
    client := &http.Client{}
    go startClient(client, i, nrOperations)
  }
  wg.Wait()

  elapsed := time.Since(start)
  
  printStats(elapsed, nrOperations, nrClients);
}