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

type Coordinator struct {
	FinalValue int
}

func (c *Coordinator) ProcessFinalValue(arg int){

}

func trigger(client *rpc.Client, arg *shared.ConsensusArgs){
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

	the_coordinator := &Coordinator{FinalValue: 0}
	rpc.Register(the_coordinator)

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
		fmt.Println("Round", consensus_round, ": Asked for number below ", num)
		args := &shared.ConsensusArgs{ConsensusRoundID: consensus_round, UpperBound : num}
		//RPC Call for random value in upper_bound
		for _,client := range peer_connections {
			go trigger(client, args)
		}
		consensus_round += 1
	}

	// TODO 2/13/2026
	// Add RPC callable for coordinator that acceptors can call to send in consented values
		// Create neccessary utilities and helpers 

	// Should acceptor RPC call coordinator OR coordinator RPC call acceptor
	// if coordinator RPC calls, it can control when new consensus rounds begin
	 

	//spawn a bunch of simuatenous goroutines to trigger consensus to proposers
}