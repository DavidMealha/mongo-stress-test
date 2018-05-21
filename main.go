package main

import (
	"fmt"
	"time"
	"math/rand"
	"sync"
)

var (
	wg sync.WaitGroup
)

func init() {

}

func startClient(clientNr int) {
	for i := 0; i < 10; i++ {
		rand := rand.Intn(2)
		if rand == 0 {
			fmt.Printf("Do write in Client %v =>", i)
			fmt.Print(time.Now())
			fmt.Println()
		} else {
			fmt.Printf("Do read in Client %v =>", i)
			fmt.Print(time.Now())
			fmt.Println()
		}
		time.Sleep(100 * time.Millisecond)
	}
	wg.Done()
}

func main() {
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go startClient(i)
	}

	wg.Wait()

	fmt.Println("What am i doing????")
}
