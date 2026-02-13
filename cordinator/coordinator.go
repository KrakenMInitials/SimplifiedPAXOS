package main

import (
	"SimplifiedPAXOS/shared"
	"fmt"
	"net"
	"net/rpc"
	"time"
	"os"
	"math/rand"
	"sync"
)

func Trigger(client *rpc.Client, arg *shared.ConsensusArgs){
	var response_value shared.AcceptResponse;
	err := client.Call("Node.TriggerConsensus", arg, &response_value)
	if (err != nil){
		fmt.Println("Trigger Error: ", err)
		return	
	}
}

func main(){
	//assumes nodes up and running from .bat
	//set up a listener for acceptors to reach out
	listener, err := net.Listen("tcp", shared.AddressRegistry[0])
	if (err != nil){
		fmt.Println("Node net.Listener failed: ", err)
		//node turned stupid, reset node?
		os.Exit(1)
	}
	defer listener.Close()

	var peers []int = shared.Known_proposers

	//create connections to known acceptors concurrently
	var wg sync.WaitGroup
	responsesCh := make(chan *rpc.Client, len(peers))

	for _,id := range peers {
		wg.Add(1) 
		go func(){
			defer wg.Done()

			var client *rpc.Client
			for {
				client, err = rpc.Dial("tcp", shared.AddressRegistry[id])
				if (err != nil) {
					fmt.Println("Failed to reach Node ", id , ": ", err, ". Retrying...")
					time.Sleep(1 * time.Second) //3 sec retry timer
				} else {
					fmt.Println("Connected to Node ", id)
					break}
			} 
			
			responsesCh<-client
		}()
	}

	wg.Wait()
	close(responsesCh)

	var peer_connections []*rpc.Client

	for x := range responsesCh {
		peer_connections = append(peer_connections, x)
	}

	ticker := time.NewTicker(5 * time.Second) //time-based and doesn't wait for nodes to finish processing last consensus -out of scope
	consensus_round := 0
	for range ticker.C {
		num := rand.Intn(100)
		fmt.Println("Asked for number below ", num)
		args := &shared.ConsensusArgs{ConsensusRoundID: consensus_round, UpperBound : num}
		for _,client := range peer_connections {
			go Trigger(client, args)
		}
		consensus_round += 1
	}


	//spawn a bunch of simuatenous goroutines to trigger consensus to proposers
}