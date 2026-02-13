package shared

import "net/rpc"

//deterministic id-port mappings
var AddressRegistry = map[int]string {
	0: "localhost:8000", //coordinator and distinguished learner 
	1: "localhost:8001", //proposer
	2: "localhost:8002", //proposer
	3: "localhost:8003", //acc
	4: "localhost:8004", //acc
	5: "localhost:8005", //acc 
}

//Class was only used for debugging and conceptual work
//All nodes know behaviour of proposer, acceptor, learner
//I change my mind, its complex to assume a single node changes behaviour depending on message type recieved
//Nodes have fixed classes assigned from CLI args
type Class int 
const (
	UNKNOWN_CLASS Class = iota
	PROPOSER_CLASS //1
	ACCEPTOR_CLASS //2
	LEARNER_CLASS //3 but unused
)

var Known_acceptors []int = []int{3,4,5}
var Known_proposers []int = []int{1,2}

type PrepareRequest struct {
	ConsensusRoundID int
	PrpslNum int //Proposal Number
	PrpsdValue int //Value
} 

type PrepareResponse struct {
	ConsensusRoundID int
	Agreement bool 
	// Yes if accept request highest prpslnum regardless of existing value in acceptor; proposed value or not handled proposer side
	// No if not highest proposal num recieved or against any promises;
	HighestPrpslNum int
	ExistingVal int //Highest known value by acceptor
}

type AcceptRequest struct {
	ConsensusRoundID int
	PrpslNum int 
	PrpsdValue int
}

type AcceptResponse struct {
	ConsensusRoundID int
	FinalValue int
}

type ConsensusArgs struct { //RPC args to trigger consensus from coordinator to proposers
	ConsensusRoundID int
	UpperBound int
}

type IdToRPCClientTuple struct { //tuple workaround as helper to modify PeerClient state; 
	ID int
	Client *rpc.Client
}