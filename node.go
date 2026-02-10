package main

import (
	"SimplifiedPAXOS/shared"
	"fmt"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"errors"
)

type Node struct {
	ID int
	Class Class
	NodeListener net.Listener
	Acceptors []*rpc.Client

}



//Use my understanding of Paxos and not the midterm's 
//Will spawn different processes for node.go in a different file
// func (n *Node) ProcessPrepareRequest(arg )  {.} //(*Node) refer to self_node

// region PROPOSER BEHAVIOUR

// interpret func names as ProcessMYAcceptRequest
// acceptor nodes do not need to RPC proposers 

// NEED TO SET TIMEOUT ON RPC CALLS TO ACCEPTORS
func (n *Node) ProcessAcceptRequest(arg *shared.AcceptRequest, reply *shared.AcceptResponse) error {	
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.New("TriggerConsensus called on non-proposer")
	}
} 
func (n *Node) ProcessPrepareRequest(arg *shared.PrepareRequest, reply *shared.PrepareResponse) error {
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.New("TriggerConsensus called on non-proposer")
	}
}


//RPC Callable by coordinator
// args has upper bound
func (n *Node) TriggerConsensus(args *int, reply *shared.AcceptResponse) error {
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.new("TriggerConsensus called on non-proposer")
	}
	upper_bound := *args
	v := rand.Intn(upper_bound)
	n := self_node.id
	

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

	listener, err := net.Listen("tcp", shared.AddressRegistry[self_node.ID]) //open listener on port 
	if err != nil {
		fmt.Println("Node net.Listener failed: ", err)
		//node turned stupid, reset node?
		os.Exit(1)
	}
	defer listener.Close()

	self_node.NodeListener = listener
	if self_node.Class == shared.PROPOSER_CLASS || self_node.Class == shared.ACCEPTOR_CLASS{
		rpc.Register(self_node) //only coordinator can RPC call / or only coordinator calls TriggerConsensus
	} else {
		fmt.Println("Node Class ID Unknown (PROPOSER - 1) (ACCEPTOR - 2)")
		return 1
	}

	rpc.Accept(listener) //blocks unless goroutined
}
