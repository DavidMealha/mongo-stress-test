package main

import (
	"flag"
	"fmt"
	"time"
	"math/rand"
)

func init() {

}

func startClient() {
	for i := 0; i < 100; i++ {
		rand := rand.Intn()
		if rand == 0 {
			//perform write
			fmt.Println("Do write")
		} else {
			//perform read
			fmt.Println("Do read")
		}
	}
}

func main() {

	for i := 0; i < 10; i++ {
		go startClient()
	}

	fmt.Println("What am i doing????")
}