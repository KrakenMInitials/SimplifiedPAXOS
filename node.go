package main

import (
	"SimplifiedPAXOS/shared"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"errors"
	"math/rand"
	"time"
	"sync"
)

type Node struct {
	ID int
	Class shared.Class
	NodeListener net.Listener
	Acceptors []*rpc.Client

	highestPrpslNum int
	highestVal int
}



//Use my understanding of Paxos and not the midterm's 
//Will spawn different processes for node.go in a different file

// region PROPOSER BEHAVIOUR

// interpret func names as ProcessMYAcceptRequest
// acceptor nodes do not need to RPC proposers 

// NEED TO SET TIMEOUT ON RPC CALLS TO ACCEPTORS
	func (n *Node) ProcessPrepareRequest(arg *shared.PrepareRequest, reply *shared.PrepareResponse) error {
		if (n.Class != shared.ACCEPTOR_CLASS) {
			return errors.New("ProcessPrepareRequest() called on non-acceptor")
		}

		fmt.Println("Acceptor recieved #", arg.PrpslNum, ": ", arg.PrpsdValue)
		return nil
		//acceptor behavior on prepare request
	}

	// func (n *Node) ProcessAcceptRequest(arg *shared.AcceptRequest, reply *shared.AcceptResponse) error {	
	// 	if (n.Class != shared.ACCEPTOR_CLASS) {
	// 		return errors.New("ProcessAcceptRequest() called on non-acceptor")
	// 	}

	// 	proposal_num = arg.PrpslNum
	// 	proposal_val = arg.PrpsdValue
		
	// 	return 
	// 	//acceptor behavior on accept request
	// } 



//RPC Callable by coordinator
// args has upper bound
func (n *Node) TriggerConsensus(args *int, reply *shared.AcceptResponse) error {
	fmt.Println("Consensus triggered by coordinator.")
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.New("TriggerConsensus called on non-proposer")
	}
	upper_bound := *args

	Phase1:
	for i:=0;;i+=1{
		id, err := strconv.Atoi(shared.AddressRegistry[n.ID])
		num := id * i

		v := rand.Intn(upper_bound)

		prepare_request := shared.PrepareRequest{PrpslNum: num, PrpsdValue: v}
		responsesCh := make(chan shared.PrepareResponse, len(n.Acceptors))
		failuresCh := make(chan struct{}, len(n.Acceptors))


		//spawn goroutines to contact acceptors
		for _,client := range n.Acceptors{
			go func (c *rpc.Client){

				var prepare_response shared.PrepareResponse
				call := c.Go("Node.ProcessPrepareRequest", &prepare_request, &prepare_response, make(chan *rpc.Call, 1)) //is an async call, not a goroutine
				
				select {
				case <-call.Done:
					if (call.Error != nil){
						fmt.Println("Acceptor node failed to RPC ProcessPrepareRequest: ", err)
						failuresCh<-struct{}{}
						//dies and decrements counter
						// return Errors.New("Acceptor node failed to RPC ProcessPrepareRequest")
					}
					responsesCh <- prepare_response
				case <-time.After(time.Second * 2):
					//dies and decrements counter
					failuresCh<-struct{}{}
				}
			} (client)
		}

		success_count := 0
		failure_count := 0
		for {
			select {
			case <-failuresCh:
				failure_count += 1
			case <-responsesCh:
				success_count += 1
			}
			

			if (failure_count == len(n.Acceptors)) {break}
			if (success_count >= len(n.Acceptors)/2) {
				//begin phase 2	
				break Phase1
			}
			if (failure_count + success_count == len(n.Acceptors)){break}
		}
	
	}

	// Phase2:
	return nil
}

// endregion

//triggered by coordinator using http
//rpc.Register w/o (*Server).Register uses the global default server when net/rpc is imported
func main(){
	node_id, err := strconv.Atoi(os.Args[1])
	if (node_id <= 0){
		fmt.Println("Node ID cannot be negative")
		os.Exit(1)
	}
	class, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Process failed to parse CLI Arguments: ", err)
		os.Exit(1)
	}
	self_node := &Node{
		ID: node_id,
		Class: shared.Class(class),
		NodeListener: nil,
	}
	
	// Validate node Class
	if self_node.Class == shared.PROPOSER_CLASS || self_node.Class == shared.ACCEPTOR_CLASS{
		rpc.Register(self_node) //only coordinator can RPC call / or only coordinator calls TriggerConsensus
	} else {
		fmt.Println("Node Class ID Unknown (PROPOSER - 1) (ACCEPTOR - 2)")
		return
	}

	//creates Node global listener
	listener, err := net.Listen("tcp", shared.AddressRegistry[self_node.ID]) //open listener on port 
	if err != nil {
		fmt.Println("Node net.Listener failed: ", err)
		//node turned stupid, reset node?
		os.Exit(1)
	}
	defer listener.Close()
	self_node.NodeListener = listener
	
	//find static peers from shared.go
	var peers []int
	switch self_node.Class {
	case shared.PROPOSER_CLASS:
		peers = shared.Known_acceptors
	case shared.ACCEPTOR_CLASS:
		peers = shared.Known_proposers
	}

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
					time.Sleep(3 * time.Second) //3 sec retry timer
				} else {break}
			} 
			
			responsesCh<-client
		}()
	}

	wg.Wait()
	close(responsesCh)
	for x := range responsesCh{
		self_node.Acceptors = append(self_node.Acceptors, x)
	}
	
	go func (){
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			fmt.Println("Heartbeat: This node is still active...")
		}
	}()
	rpc.Accept(listener) //blocks unless goroutined
	//kickstarts the RPC callables
}
