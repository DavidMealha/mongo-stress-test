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
  ids = []string{"537f700b537461b70c5f0000","537f700b537461b70c5f0001","537f700b537461b70c5f0002","537f700b537461b70c5f0003"}
)

const (
  DATABASE_ADDRESS  = "localhost:27019"
  DATABASE_NAME     = "users"
  COLLECTION_NAME   = "customers"
  PROXY_ADDRESS     = "http://localhost:8126"
  //SERVICE_ADDRESS   = "http://localhost:8080/"
  //SERVICE_ADDRESS   = "http://3.120.161.145:8080/"
  //SERVICE_ADDRESS   = "http://52.213.179.93:8080/"
  WRITE_RATE        = 10
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

func startClient(client *http.Client, clientNr int, nrOperations int) {
  for i := 0; i < nrOperations; i++ {
    rand := rand.Intn(100)

    if rand < WRITE_RATE {
      // insertUser(client)
  
      randOp := rand / 25

      if randOp == 1 {
        insertUserFromWrapper(client)
      } else if randOp == 2 {
        fmt.Println("Update operation")
        //updateUserFromWrapper(client)
      } else if randOp == 3 {
        fmt.Println("Replace operation")
        // replaceUserFromWrapper(client)
      } else {
        fmt.Println("Delete operation")
        // deleteUserFromWrapper(client)
      }
    } else {
      readUserFromDatabase()
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
  resp, err := http.Get(SERVICE_ADDRESS + "customers/" + randomHex(12))

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

func insertUserFromWrapper(client *http.Client) {
  start := time.Now()
  url := PROXY_ADDRESS
  str := `{"operationType":"INSERT",` + 
          `"fullDocument":{"username":"` + getRandomString(12) + 
          `","password":"` + getRandomString(12) +
          `","email":"` + getRandomString(12) +
          `","firstName":"` + getRandomString(12) +
          `","lastName":"` + getRandomString(12) +
          `","origin_created_at":"` + strconv.FormatInt(time.Now().UnixNano()/int64(time.Microsecond),10) + 
          `"},"ns":{"coll":"` + COLLECTION_NAME +
          `","db":"` + DATABASE_NAME +
          `"},"documentKey":{"_id":"` + randomHex(12) + `"}}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")
  req.Header.Set("Connection", "keep-alive")

  resp, err := client.Do(req)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()
  elapsed := int(time.Since(start) / time.Millisecond)
  writeLatencies = append(writeLatencies, elapsed)
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
  elapsed := int(time.Since(start) / time.Millisecond)
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

func printStats() {
  const writeRate = float64(WRITE_RATE) / 100;
  const readRate = 1 - writeRate;

  var avgWriteLatency = getAverage(writeLatencies);
  var avgReadLatency = getAverage(readLatencies);

  var throughputWrite = 1000 / avgWriteLatency;
  var throughputRead = 1000 / avgReadLatency;

  // sum = (writeRate * throughputWrite) + (readRate * throughputRead)
  var sum = (writeRate * float64(throughputWrite)) + (readRate * float64(throughputRead))
  //var sum = (writeRate * float64(throughputWrite))

  fmt.Println("=======================================");
  fmt.Println("============= STATISTICS ==============");
  fmt.Println("=======================================");
  fmt.Println("Write Rate =>\t", writeRate);
  fmt.Println("Read Rate =>\t", readRate);
  fmt.Println("=======================================");
  fmt.Println("Average Write Latency =>", avgWriteLatency)
  fmt.Println("Average Read Latency =>\t", avgReadLatency);
  fmt.Println("=======================================");
  fmt.Println("Throughput Write =>\t", throughputWrite);
  fmt.Println("Throughput Read =>\t", throughputRead);
  fmt.Println("=======================================");
  fmt.Println("Total Throughput per Second =>\t", sum);
  fmt.Println("Total Throughput per Minute =>\t", sum * 60);
  fmt.Println("Total Throughput per Hour =>\t", sum * 60 * 60);
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
  flag.Parse();

  wg.Add(nrClients)
  for i := 0; i < nrClients; i++ {
    client := &http.Client{}
    go startClient(client, i, nrOperations)
  }
  wg.Wait()

  elapsed := time.Since(start)
  fmt.Println("Took ", elapsed, " ms to perform 2500 operations.")

  printStats();
}

