[Simplified PAXOS Project.pdf](https://github.com/user-attachments/files/25377183/Simplified.PAXOS.Project.pdf)
ReadMe to be directly polished up and formatted. PDF link contains design decisions made and scope of project. 

How to run demo:
* For Windows users, demo.bat runs a configuration of a coordinator/distinguished learner + 4 proposers + 7 acceptors
* For other users, run the following while in the root of the repository 
(go run node.go <node_id> <1 for proposer | 2 for acceptor>) coordinator is necessary:
* go run .\cordinator\coordinator.go                
* go run node.go 1 1
* go run node.go 2 1
* go run node.go 3 1
* go run node.go 4 1
* go run node.go 5 2
* go run node.go 6 2
* go run node.go 7 2
* go run node.go 8 2
* go run node.go 9 2
* go run node.go 10 2
* go run node.go 11 2
* OR spawn nodes as desired, but shared/shared.go needs modifications in 
  * AddressRegistry 
  * Known_acceptors
  * Known_proposers
