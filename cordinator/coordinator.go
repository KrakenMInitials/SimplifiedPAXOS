package coordinator

import (
	"SimplifiedPAXOS/shared"
	"fmt"
	"net"
	"net/rpc"
	"time"
	"os"
	"math/rand"
)

func Trigger(client *rpc.Client, upper_bound int){
	var response_value shared.AcceptResponse;
	client.Call("Node.TriggerConsensus", &upper_bound, &response_value)
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

	client1, err := rpc.Dial("tcp", shared.AddressRegistry[1])
	if (err != nil){
		fmt.Println("Failed to connect to Node id 1")
	}
	defer client1.Close()
	client2, err := rpc.Dial("tcp", shared.AddressRegistry[2])
	if (err != nil){
		fmt.Println("Failed to connect to Node id 2")
	}
	defer client2.Close()

	ticker := time.NewTicker(10)	
	for range ticker.C {
		num := rand.Intn(100)
		go Trigger(client1, num)
		go Trigger(client2, num)
	}


	//spawn a bunch of simuatenous goroutines to trigger consensus to proposers
}