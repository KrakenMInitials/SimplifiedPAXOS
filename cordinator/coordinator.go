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
	ValuesQueue chan shared.ConsensusRoundToValueTuple //unbuffered to make blocking wait
}

func (c *Coordinator) ProcessFinalValue(arg *shared.ConsensusRoundToValueTuple, reply *int) error {
	fmt.Println("An acceptor sent", arg.Value, "for", arg.RoundID)
	*reply = 1
	c.ValuesQueue<- *arg 
	return nil
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

	the_coordinator := &Coordinator{
		ValuesQueue: make(chan shared.ConsensusRoundToValueTuple),
	}
	rpc.Register(the_coordinator)
	go rpc.Accept(listener) //create server

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
	consensus_round := 0 //could be modified to be a state property
	for range ticker.C {
		num := rand.Intn(100)
		fmt.Println("Round", consensus_round, ": Asked for number below ", num)
		args := &shared.ConsensusArgs{ConsensusRoundID: consensus_round, UpperBound : num}
		// RPC Call for random value in upper_bound
		for _,client := range peer_connections {
			go trigger(client, args) //spawn goroutines to trigger consensus to proposers
		}

		// 

		values_set := make(map[int]int) //frequency dictionary of values
		success_count := 0
		for {
			value := <- the_coordinator.ValuesQueue
			if value.RoundID < consensus_round {
				continue
			}

			values_set[value.Value] = values_set[value.Value] + 1
			success_count += 1
			if success_count > len(shared.Known_acceptors)/2 { //distinguished learner might not be supposed to continue on majority hit
				break
			}
		}
		
		fmt.Println("Round", consensus_round, ": Consensus reached on :", shared.FindMajority(values_set))

		time.Sleep(1 * time.Second)
		consensus_round += 1
	}


	// TODO 2/13/2026
	// Add RPC callable for coordinator that acceptors can call to send in consented values

	// Should acceptor RPC call coordinator OR coordinator RPC call acceptor
	// if coordinator RPC calls, it can control when new consensus rounds begin
	 

}