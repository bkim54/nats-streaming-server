package main

import (
	"github.com/bkim54/nats-streaming-server/stores"
	//"bufio"
	//"os"
	//"fmt"
	//"strings"
	//"os/signal"
	//"runtime"
	//"os/signal"
	//"runtime"
)

func main() {
	//stores.ReadFileStore("/Users/bkim/nats-streaming-server-darwin-amd64/store");
	stores.ReadFileStore("/Users/bkim/mygo/src/github.com/bkim54/nats-streaming-server/server-store");
	//
	//reader := bufio.NewReader(os.Stdin)
	//fmt.Println("Simple Shell")
	//fmt.Println("---------------------")
	//
	//for {
	//	fmt.Print("-> ")
	//	text, _ := reader.ReadString('\n')
	//	input := strings.Fields(text);
	//	if (input[0] == "stop") {
	//		os.Exit(0)
	//	}
	//	fmt.Println(text)
	//}


	//// a channel to tell it to stop
	//stopchan := make(chan os.Signal, 1)
	//signal.Notify(stopchan, os.Interrupt)
	//
	//// a channel to signal that it's stopped
	//stoppedchan := make(chan struct{})
	//
	//go func(){ // work in background
	//	// close the stoppedchan when this func
	//	// exits
	//	defer close(stoppedchan)
	//	// TODO: do setup work
	//
	//	for {
	//		select {
	//		default:
	//			reader := bufio.NewReader(os.Stdin)
	//			fmt.Println("Simple Shell")
	//			fmt.Println("---------------------")
	//
	//			for {
	//				fmt.Print("-> ")
	//				text, _ := reader.ReadString('\n')
	//				if (text == "stop\n") {
	//					os.Exit(0)
	//				}
	//				fmt.Println(text)
	//			}
	//		case <-stopchan:
	//		// stop
	//			return
	//		}
	//	}
	//}()
	//
	//runtime.Goexit()


}
