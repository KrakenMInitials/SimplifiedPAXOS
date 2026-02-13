package shared


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

// var Known_acceptors []int = []int{3,4,5}
// var Known_proposers []int = []int{1,2}
var Known_acceptors []int = []int{3}
var Known_proposers []int = []int{1}


type PrepareRequest struct {
	PrpslNum int //Proposal Number
	PrpsdValue int //Value
} 

type PrepareResponse struct {
	Agreement bool 
	// Yes if accept request highest prpslnum;
	// Ignored if not highest proposal num recieved or against any promises;
	// No if highest proposal num but highest value exists
	HighestVal int //Highest known value by acceptor
}

type AcceptRequest struct {
	PrpslNum int 
	PrpsdValue int
}

type AcceptResponse struct {
	FinalValue int
}