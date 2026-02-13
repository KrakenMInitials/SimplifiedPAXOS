START "Coordinator" cmd /k go run .\cordinator\coordinator.go
timeout \t 5 >nul
START "Proposer 1" cmd /k go run node.go 1 1
timeout \t 5 >nul
START "Proposer 2" cmd /k go run node.go 2 1
timeout \t 5 >nul
START "Acceptor 1" cmd /k go run node.go 3 2
timeout \t 5 >nul
START "Acceptor 2" cmd /k go run node.go 4 2
timeout \t 5 >nul
START "Acceptor 3" cmd /k go run node.go 5 2
