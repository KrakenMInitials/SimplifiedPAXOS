package shared

import "net/rpc"

//deterministic id-port mappings
var AddressRegistry = map[int]string {
	0: "localhost:8000", //coordinator and distinguished learner 
	1: "localhost:8001", //proposer
	2: "localhost:8002", //proposer
	3: "localhost:8003", //proposer
	4: "localhost:8004", //proposer
	5: "localhost:8005", //acceptor
	6: "localhost:8006", //acceptor
	7: "localhost:8007", //acceptor
	8: "localhost:8008", //acceptor
	9: "localhost:8009", //acceptor
	10: "localhost:8010", //acceptor
	11: "localhost:8011", //acceptor
}

//Nodes have fixed classes assigned from CLI args
type Class int 
const (
	UNKNOWN_CLASS Class = iota
	PROPOSER_CLASS //1
	ACCEPTOR_CLASS //2
	LEARNER_CLASS //3 but unused
)

var Known_acceptors []int = []int{5,6,7,8,9,10, 11} // 6 acceptors (even, so add one more for odd)
var Known_proposers []int = []int{1,2,3,4}

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
	Agreement bool 
	// Yes if no new highest proposal num came in 
	// No if new highest proposal num recieved in between
	HighestPrpslNum int // if agreement = no
	Value int
}

type ConsensusArgs struct { //RPC args to trigger consensus from coordinator to proposers
	ConsensusRoundID int
	UpperBound int
}

type IdToRPCClientTuple struct { //tuple workaround as helper to modify PeerClient state; 
	ID int
	Client *rpc.Client
}

type ConsensusRoundToValueTuple struct {
	RoundID int
	Value int
}

// majority had to be found too commonly; made shared helper function
// takes in value : frequency map
func FindMajority[T comparable] (freq_dic map[T]int) T {
	var majorityVal T
	var majorityCount int
	for val, count := range freq_dic {
		if count > majorityCount {
			majorityVal = val
		}
	}
	return majorityVal
}