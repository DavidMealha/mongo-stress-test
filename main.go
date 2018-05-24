package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
	"net/http"
)

var (
	wg sync.WaitGroup
	letters := []string{"a","b","c","d","e","f","g","h","i","j","k","l","m","n","o","p","q","r","s","t","u","v","w","x","y","z"}
)

func init() {

}

func startClient(clientNr int) {
	for i := 0; i < 10; i++ {
		rand := rand.Intn(2)
		if rand == 0 {
			fmt.Printf("Do write in Client %v =>", clientNr)
			fmt.Print(time.Now())
			fmt.Println()
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
	resp, err := http.PostForm("localhost:8080/customers",
								url.Values{
									"username": getRandomString(8),
									"password": getRandomString(12),
									"email": getRandomString(10),
									"firstName": getRandomString(6),
									"lastName": getRandomString(8)
								})
	if err != nil {
		panic(err)
	}
	fmt.Println("Response Status:", resp.Status())
}

func getRandomString(size int) string{
	lettersLen := len(letters)
	var str string
	for i := 0; i < size; i++ {
		str += letters[rand.Intn(lettersLen)]
	}
	return string
}

func readUsers() {
	resp, err := http.Get("localhost:8080/customers")
	if err != nil {
		panic(err)
	}
	fmt.Println("Response Status:", resp.Status())
}

func main() {
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go startClient(i)
	}

	wg.Wait()

	fmt.Println("What am i doing????")
}
