package main

import (
  "fmt"
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
)

var (
  wg sync.WaitGroup
  letters = []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
  writeLatencies []int
  readLatencies []int
)

const (
  DATABASE_ADDRESS  = "localhost:27019"
  DATABASE_NAME     = "users"
  COLLECTION_NAME   = "customers"
  PROXY_ADDRESS     = "http://localhost:8127"
  SERVICE_ADDRESS   = "http://localhost:8080/"
  //SERVICE_ADDRESS   = "http://52.213.179.93:8080/"
  WRITE_RATE        = 50
)

func init() {
  rand.Seed(time.Now().UnixNano())
}

func startClient(clientNr int) {
  for i := 0; i < 5; i++ {
    rand := rand.Intn(100)
    if rand < WRITE_RATE {
      //fmt.Printf("Do write in Client %v \n", clientNr)
      insertUser(clientNr)
    } else {
      //fmt.Printf("Do read in Client %v \n", clientNr)
      readUser()
    }
    //time.Sleep(30 * time.Millisecond)
  }
  defer wg.Done()
}

func insertUser(clientNr int) {
  start := time.Now()
  fmt.Printf("Client %v sent request at %v\n", clientNr, start)
  str := `{"username":"` + getRandomString(8) + 
         `","password":"` + getRandomString(12) + 
         `","email":"` + getRandomString(10) + 
         `","firstName":"` + getRandomString(6) + 
         `","lastName":"` + getRandomString(8) + `"}`

  var jsonStr = []byte(str)

  req, err := http.NewRequest("POST", SERVICE_ADDRESS + "register", bytes.NewBuffer(jsonStr))
  req.Header.Set("Content-Type", "application/json")

  client := &http.Client{}
  resp, err := client.Do(req)
  fmt.Printf("Service response to client %v sent request at %v\n", clientNr, start)

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  elapsed := int(time.Since(start) / time.Millisecond)
  writeLatencies = append(writeLatencies, elapsed)
}

func readUser() {
  start := time.Now()
  resp, err := http.Get(SERVICE_ADDRESS + "customers/" + randomHex(12))

  if err != nil {
    fmt.Println(err)
  }
  defer resp.Body.Close()

  elapsed := int(time.Since(start) / time.Millisecond)

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
  //writeLatencies = append(writeLatencies, elapsed)
  fmt.Println("Took ", elapsed, " ms.")
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
  //readLatencies = append(readLatencies, elapsed.String())
  fmt.Println("Took ", elapsed, " ms.")
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

  //calc average write latency
  var avgWriteLatency = getAverage(writeLatencies);
  //calc average read latency
  var avgReadLatency = getAverage(readLatencies);

  //throughputWrite = 1000 / avgWriteLatency
  var throughputWrite = 1000 / avgWriteLatency;
  //throughputRead = 1000 / avgReadLatency
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

  wg.Add(2)
  for i := 0; i < 2; i++ {
    go startClient(i)
  }
  wg.Wait()

  //end time
  elapsed := time.Since(start)
  fmt.Println("Took ", elapsed, " ms to perform 2500 operations.")

  printStats();
  // fmt.Printf("Write latencies %\n", writeLatencies)
  // fmt.Printf("Read latencias %\n", readLatencies)

  //fmt.Println("Finished.")
}
