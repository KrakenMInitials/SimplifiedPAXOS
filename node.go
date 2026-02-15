package main

import (
	"SimplifiedPAXOS/shared"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

type Node struct { //stores Node state
	ID int
	Class shared.Class
	NodeListener net.Listener
	PeerClients map[int]*rpc.Client //stores id to rpc.client instances for known peers

	// dynamic states
	proposedPrpslNum int
	highestPrpslNum int
	knownVal *int //has to be pointer to be nil-checkable
	complete bool //current implementation only uses for acceptors to not repeat accepts for same round
	
	ConsensusRound int
}


//Use my understanding of Paxos and not the midterm's 
//Will spawn different processes for node.go in a different file

// region PROPOSER BEHAVIOUR

// interpret func names as ProcessMYAcceptRequest
// acceptor nodes do not need to RPC proposers 

// RPC Callable by proposers on acceptor
// NEED TO SET TIMEOUT ON RPC CALLS TO ACCEPTORS
func (n *Node) ProcessPrepareRequest(args *shared.PrepareRequest, reply *shared.PrepareResponse) error { //phase 1
	if (n.Class != shared.ACCEPTOR_CLASS) {
		return errors.New("ProcessPrepareRequest() called on non-acceptor")
	}
	fmt.Println("Round", args.ConsensusRoundID, ": Acceptor recieved PrepareRequest #", args.PrpslNum, ": ", args.PrpsdValue)
	
	//reset node state
	if args.ConsensusRoundID > n.ConsensusRound {
		fmt.Println("New consensus round detected.")
		n.ConsensusRound = args.ConsensusRoundID
		n.proposedPrpslNum = 0
		n.highestPrpslNum = 0 //proposers use this as highest known proposal number; acceptors use it as benchmark to accept/reject proposals
		n.knownVal = nil 
		n.complete = false
	}
	
	if (n.highestPrpslNum != 0) && (args.PrpslNum <= n.highestPrpslNum){ //true reject
	//handles edge case dupe proposal numbers but not thoroughly
		reply.Agreement = false
		reply.HighestPrpslNum = n.highestPrpslNum
		reply.ExistingVal = args.PrpsdValue
		return nil
	}
	if (n.knownVal == nil) {
		//modify states
		n.knownVal = &(args.PrpsdValue)
		n.highestPrpslNum = args.PrpslNum

		reply.Agreement = true
		reply.HighestPrpslNum = args.PrpslNum
		reply.ExistingVal = args.PrpsdValue
	} else if args.PrpslNum > n.highestPrpslNum {
		//modify states
		n.highestPrpslNum = args.PrpslNum

		reply.Agreement = true
		reply.HighestPrpslNum = args.PrpslNum
		reply.ExistingVal = *(n.knownVal)
	}
	
	return nil
}

func (n *Node) ProcessAcceptRequest(args *shared.AcceptRequest, reply *shared.AcceptResponse) error {	
	if (n.Class != shared.ACCEPTOR_CLASS) {
		return errors.New("ProcessAcceptRequest() called on non-acceptor")
	}

	fmt.Println("Round", args.ConsensusRoundID, ": Acceptor recieved AcceptRequest #", args.PrpslNum, ": ", args.PrpsdValue)

	//Q: Can accept request have proposal number lower than n.highest proposal number
	//A: prolly provaeable so nahh
	if n.highestPrpslNum < args.PrpslNum {
		fmt.Println("The sky is falling : known highest", n.highestPrpslNum, "incoming #:", args.PrpslNum)
		return errors.New("The sky is falling")
	}

	if n.highestPrpslNum == args.PrpslNum { //value accepted and sent to distinguished learner
		n.complete = true

		reply.ConsensusRoundID = n.ConsensusRound
		reply.Agreement = true
		reply.HighestPrpslNum = n.highestPrpslNum
		reply.Value = args.PrpsdValue


		go func (){ //goroutine because coordinator uses unbuffered channel and will blocking wait
			//problem if goroutine fails 
			client := n.PeerClients[0]
			arg := &shared.ConsensusRoundToValueTuple{RoundID: n.ConsensusRound, Value: args.PrpsdValue,}
			if err := client.Call("Coordinator.ProcessFinalValue", arg, nil) ; err != nil { 
				fmt.Println("Round", n.ConsensusRound, ": failed to send value to coordinator.")
			}
			fmt.Println("Round", args.ConsensusRoundID, ": Acceptor sent final value and completed.")
		}()
		
		return nil
	} else {
		reply.ConsensusRoundID = n.ConsensusRound
		reply.Agreement = false
		reply.HighestPrpslNum = n.highestPrpslNum
		reply.Value = *(n.knownVal)
	}
	
	return nil
	//acceptor behavior on accept request
} 


// RPC Callable by coordinator on proposers
// args has upper bound set by coordinator
func (n *Node) TriggerConsensus(args *shared.ConsensusArgs, reply *shared.AcceptResponse) error { //phase 1
	fmt.Println("Consensus triggered by coordinator.")
	if (n.Class != shared.PROPOSER_CLASS) {
		return errors.New("TriggerConsensus called on non-proposer")
	}

	//reset node state
	if args.ConsensusRoundID > n.ConsensusRound {
		n.ConsensusRound = args.ConsensusRoundID
		n.proposedPrpslNum = 0
		n.highestPrpslNum = 0
		n.knownVal = nil 
		n.complete = false //not neccessarily used by proposers
	}

	//parse args
	consensus_round := args.ConsensusRoundID
	upper_bound := args.UpperBound

	values_set := make(map[int]int) //stores <response value>:<frequency>
	responded_acceptor_ids := make(map[int]struct{}) //set of ids for acceptor that responded for use in phase 2
	var responded_acceptors_MU sync.Mutex

	RetryProposal:
	for i:=1;;i+=1{
		time.Sleep(1 * time.Second) // slowdown for monitoring purposes 

		//
		// PHASE 1
		//
		addr := shared.AddressRegistry[n.ID]

		port, err := strconv.Atoi(addr[len(addr)-4:])
		if err != nil {
			fmt.Println("String conversion err: ", err)
		}
		n.proposedPrpslNum = port * i

		if (n.proposedPrpslNum < n.highestPrpslNum){
			continue
		}

		v := rand.Intn(upper_bound)

		prepare_request := shared.PrepareRequest{ConsensusRoundID: consensus_round, PrpslNum: n.proposedPrpslNum, PrpsdValue: v}
		fmt.Println("Proposing #", n.proposedPrpslNum, " ", v)
		
		responsesCh := make(chan shared.PrepareResponse, len(n.PeerClients))
		failuresCh := make(chan struct{}, len(n.PeerClients))

		//spawn goroutines to call and collect prepare responses from acceptors 
		for peer_id,client := range n.PeerClients{
			go func (c *rpc.Client){

				var prepare_response shared.PrepareResponse
				call := c.Go("Node.ProcessPrepareRequest", &prepare_request, &prepare_response, make(chan *rpc.Call, 1)) //is an async call, not a goroutine
				
				select {
				case <-call.Done:
					if (call.Error != nil){
						fmt.Println("Acceptor node failed to RPC ProcessPrepareRequest: ", call.Error.Error())
						failuresCh<-struct{}{}
						// return Errors.New("Acceptor node failed to RPC ProcessPrepareRequest")
					}
					responsesCh <- prepare_response
					
					responded_acceptors_MU.Lock()
					responded_acceptor_ids[peer_id] = struct{}{} //add peer ID to responded acceptors
					responded_acceptors_MU.Unlock()
				case <-time.After(time.Second * 2): //timeout before RPC marked as failure by default 
					//not reached since RPC calls are never ignored unless lost (code scope doesn't inlcude loss simulation)
					failuresCh<-struct{}{}
				}
			} (client)
		}

		success_count := 0
		failure_count := 0
		
		//handles prepare responses
		for {
			select {
			case <-failuresCh:
				failure_count += 1
			case x := <-responsesCh:
				success_count += 1

				//process each prepare response
				values_set[x.ExistingVal] = values_set[x.ExistingVal] + 1 //golang set alt. using maps copies by val; cannot handle pointers properly				
				// if x.HighestPrpslNum > n.highestPrpslNum {n.highestPrpslNum = x.HighestPrpslNum} //modify node state
				n.highestPrpslNum = max(x.HighestPrpslNum, n.highestPrpslNum) //modify node state but cleaner

				fmt.Println("Round", n.ConsensusRound, ": Prepare Response recieved with value ", x.ExistingVal, "and highest proposal #", x.HighestPrpslNum)
			}
		
			//exit logic if majority responded 
			if (failure_count == len(shared.Known_acceptors)) {continue RetryProposal} //all failed -> 
			if (success_count > len(shared.Known_acceptors)/2) { //majority accepted	
				break //continue phase 2
			}
			if (failure_count + success_count == len(shared.Known_acceptors)){continue RetryProposal} //didnt hit majority
		}
		
	
		majorityVal := shared.FindMajority(values_set)
		fmt.Println("Intermediary majority value was ", majorityVal)
		n.knownVal = &majorityVal //modify node state
		
		//current implementation to collect responses continues as soon as majority of acceptors is reached

		//Q: Is it possible that a majority Value is not determinable from the responded acceptors?
		//A: I dont' think so, just based on majority rule and odd number of acceptors. 
		//Q2: But if we continue as soon as majority of acceptors is hit. 
		//    Isn't it possible that the potential prepare responses not recieved yet could overturn the
		//    majority value?
		//A2: Unexplored 
		
		//
		// PHASE 2:
		//
		accept_request := &shared.AcceptRequest{ConsensusRoundID: n.ConsensusRound, PrpslNum: n.proposedPrpslNum, PrpsdValue: *n.knownVal}
		responsesCh2 := make(chan shared.AcceptResponse, len(responded_acceptor_ids))
		failuresCh2 := make(chan struct{}, len(responded_acceptor_ids))

		//call and collect accept responses from acceptors 
		for peer_id := range responded_acceptor_ids {
			client := n.PeerClients[peer_id]
			go func (c *rpc.Client){
				var accept_response shared.AcceptResponse
				call := c.Go("Node.ProcessAcceptRequest", &accept_request, &accept_response, make(chan *rpc.Call, 1)) //is an async call, not a goroutine
				
				select {
				case <-call.Done:
					if (call.Error != nil){ //error with call
						fmt.Println("Acceptor node failed to RPC ProcessPrepareRequest: ", call.Error.Error())
						failuresCh2<-struct{}{}
						// return Errors.New("Acceptor node failed to RPC ProcessPrepareRequest")
					}
					responsesCh2 <- accept_response
					responded_acceptors_MU.Lock()
					responded_acceptor_ids[peer_id] = struct{}{} //add peer ID to responded acceptors
					responded_acceptors_MU.Unlock()
				case <-time.After(time.Second * 2): //timeout before RPC marked as failure by default 
					//not reached since RPC calls are never ignored unless lost (code scope doesn't inlcude loss simulation)
					failuresCh2<-struct{}{}
				}

			} (client)	
		}

		success_count = 0
		failure_count = 0

		//handles accept responses
		for {
			select {
			case <-failuresCh2:
				failure_count += 1
			case x := <-responsesCh2:
				if x.Agreement == true {
					success_count += 1
				}

				// if x.HighestPrpslNum > n.highestPrpslNum {n.highestPrpslNum = x.HighestPrpslNum} //modify node state
				n.highestPrpslNum = max(x.HighestPrpslNum, n.highestPrpslNum) //modify node state but cleaner

				fmt.Println("Round", n.ConsensusRound, ": Prepare Response recieved with value ", x.Value, "and highest proposal #", x.HighestPrpslNum)
			}
		
			//exit logic if majority responded
			if (failure_count == len(shared.Known_acceptors)) {continue RetryProposal}
			if (success_count > len(shared.Known_acceptors)/2) {
				break RetryProposal
			}
			if (failure_count + success_count == len(shared.Known_acceptors)){continue RetryProposal}
		}
	}

	n.complete = true
	fmt.Println("Round", n.ConsensusRound, ": Proposer finished completed for node.")
	return nil
}

// endregion

//triggered by coordinator using http
//rpc.Register w/o (*Server).Register uses the global default server when net/rpc is imported
func main(){
	//validate node ID
	node_id, err := strconv.Atoi(os.Args[1])
	if (node_id <= 0){
		fmt.Println("Node ID cannot be negative")
		os.Exit(1)
	}
	
	//validate node Class
	class, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Process failed to parse CLI Arguments: ", err)
		os.Exit(1)
	}
	if class != int(shared.PROPOSER_CLASS) && class != int(shared.ACCEPTOR_CLASS){
		fmt.Println("Node Class ID Unknown (PROPOSER - 1) (ACCEPTOR - 2)")
		return
	}

	//creates Node global listener
	listener, err := net.Listen("tcp", shared.AddressRegistry[node_id]) //open listener on port 
	if err != nil {
		fmt.Println("Node net.Listener failed: ", err)
		//node turned stupid, reset node? -out of scope
		os.Exit(1)
	}
	defer listener.Close()

	//initialize Node; build and register node with args and defaults
	self_node := &Node{
		ID: node_id,
		Class: shared.Class(class),
		NodeListener: listener,
		PeerClients: make(map[int]*rpc.Client),

		proposedPrpslNum: -1,
		highestPrpslNum: -1,
		knownVal: nil,
		complete: false,

		ConsensusRound: -1,
	}
	rpc.Register(self_node)

	//find static peers from shared.go
	var peers []int
	switch self_node.Class {
	case shared.PROPOSER_CLASS:
		peers = shared.Known_acceptors
	case shared.ACCEPTOR_CLASS:
		peers = append(shared.Known_proposers, 0)
	}

	//create RPC clients/connections concurrently and store in Node state
	var wg sync.WaitGroup
	clientsCh := make(chan shared.IdToRPCClientTuple, len(peers))

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
					break
				}
			} 
			

			clientsCh <- shared.IdToRPCClientTuple{ID: id, Client: client}
		}()
	}
	wg.Wait()
	close(clientsCh)

	for x := range clientsCh{
		self_node.PeerClients[x.ID] = x.Client //modifies node state
	}
	
	// //goroutine to monitor baseline activity of node -mainly for debugging
	// go func (){
	// 	ticker := time.NewTicker(5 * time.Second)
	// 	defer ticker.Stop()
	// 	for range ticker.C {
	// 		fmt.Println("Heartbeat: This node is still active...")
	// 	}
	// }()
	rpc.Accept(listener) //blocks unless goroutined
	//kickstarts the RPC callables
}
