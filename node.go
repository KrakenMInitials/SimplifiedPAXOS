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
	"concurrency"
	"time"
)

type Node struct {
	ID int
	Class Class
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

	//acceptor behavior on prepare request
}

func (n *Node) ProcessAcceptRequest(arg *shared.AcceptRequest, reply *shared.AcceptResponse) error {	
	if (n.Class != shared.ACCEPTOR_CLASS) {
		return errors.New("ProcessAcceptRequest() called on non-acceptor")
	}

	proposal_num = arg.PrpslNum
	proposal_val = arg.PrpsdValue
	
	if proposal_num <= 
	
	n.Acceptors
	//acceptor behavior on accept request
} 



//RPC Callable by coordinator
// args has upper bound
func (n *Node) TriggerConsensus(args *int, reply *shared.AcceptResponse) error {
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.New("TriggerConsensus called on non-proposer")
	}
	upper_bound := *args

	Phase1:
	for i:=0;;i+={
		n := n.ID * i 
		v := rand.Intn(upper_bound)

		prepare_request := shared.PrepareRequest{PrpslNum: n, PrpsdValue: v}
		var wg concurrency.WaitGroup
		responsesCh := make(chan shared.PrepareResponse, len(n.Acceptors))
		failuresCh := make(chan shared.PrepareResponse, len(n.Acceptors))

		//spawn goroutines to contact acceptors
		for _,client := range n.Acceptors{

			wg.Add(1) //just a counter

			go func (c *rpc.Client){
				defer wg.Done()

				var prepare_response shared.PrepareResponse
				call := c.Go("Node.ProcessPrepareRequest", &prepare_request, &prepare_response) //is an async call, not a goroutine
				
				select {
				case <-call.Done:
					if (call.Error != nil){
						fmt.Println("Acceptor node failed to RPC ProcessPrepareRequest: ", err)
						failuresCh <-
						return
						//dies and decrements counter
						// return Errors.New("Acceptor node failed to RPC ProcessPrepareRequest")
					}
					responsesCh <- prepare_response
					return	
				case <-time.After(time.Second * 2):
					//dies and decrements counter
					failuresCh <-
					return
				}
			} (client)
		}

		success_count := 0
		failure_count := 0
		for {
			select {
			case <-failuresCh:
				failure_count += 1
			case <-responsesCh
				success_count += 1
			}
			

			select {
			case (failure_count == len(n.Acceptors)):

				break
			case (success_count >= len(n.Acceptors)/2):
				//success and begin phase 2	
				break Phase1
			case (failure_count + success_count == len(n.Acceptors)):
				break	
			}
		}
	
	}

	Phase2:



	


	//PAXOS logic

}

// endregion

//triggered by coordinator using http
//rpc.Register w/o (*Server).Register uses the global default server when net/rpc is imported
func main(){

	node_id, err := strconv.Atoi(os.Args[1])
	class, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Process failed to parse CLI Arguments: ", err)
		os.Exit(1)
	}
	self_node := Node{
		ID: node_id,
		Class: class,
		NodeListener: nil,
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
	
	//create connections to known acceptors
	known_acceptors := []int{3,4,5}
	for _,id := range known_acceptors {
		client, err := rpc.Dial(shared.AddressRegistry[id])
		if (err != nil) {
			fmt.Println("Failed to connect to Node ", id , ": ", err)
		}
		defer client.Close()

		self_node.Acceptors = append(self_node.Acceptors, client)
	}

	for 
	
	
	
	
	if self_node.Class == shared.PROPOSER_CLASS || self_node.Class == shared.ACCEPTOR_CLASS{
		rpc.Register(self_node) //only coordinator can RPC call / or only coordinator calls TriggerConsensus
	} else {
		fmt.Println("Node Class ID Unknown (PROPOSER - 1) (ACCEPTOR - 2)")
		return
	}

	rpc.Accept(listener) //blocks unless goroutined
}
