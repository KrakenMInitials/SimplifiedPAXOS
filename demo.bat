START "Coordinator" cmd /k go run .\cordinator\coordinator.go
START "Proposer 1" cmd /k go run node.go 1 1
START "Proposer 2" cmd /k go run node.go 2 1
START "Proposer 3" cmd /k go run node.go 3 1
START "Proposer 4" cmd /k go run node.go 4 1
START "Acceptor 1" cmd /k go run node.go 5 2
START "Acceptor 2" cmd /k go run node.go 6 2
START "Acceptor 3" cmd /k go run node.go 7 2
START "Acceptor 4" cmd /k go run node.go 8 2
START "Acceptor 5" cmd /k go run node.go 9 2
START "Acceptor 6" cmd /k go run node.go 10 2
START "Acceptor 7" cmd /k go run node.go 11 2

@REM START "Coordinator" cmd /k go run .\cordinator\coordinator.go
@REM timeout /t 1 >nul
@REM START "Proposer 1" cmd /k go run node.go 1 1
@REM timeout /t 1 >nul
@REM START "Proposer 2" cmd /k go run node.go 2 1
@REM timeout /t 1 >nul
@REM START "Proposer 3" cmd /k go run node.go 3 1
@REM timeout /t 1 >nul
@REM START "Proposer 4" cmd /k go run node.go 4 1
@REM timeout /t 1 >nul
@REM START "Proposer 1" cmd /k go run node.go 5 2
@REM timeout /t 1 >nul
@REM START "Acceptor 2" cmd /k go run node.go 6 2
@REM timeout /t 1 >nul
@REM START "Acceptor 3" cmd /k go run node.go 7 2
@REM timeout /t 1 >nul
@REM START "Acceptor 4" cmd /k go run node.go 8 2
@REM timeout /t 1 >nul
@REM START "Acceptor 5" cmd /k go run node.go 9 2
@REM timeout /t 1 >nul
@REM START "Acceptor 6" cmd /k go run node.go 10 2
@REM timeout /t 1 >nul
@REM START "Acceptor 7" cmd /k go run node.go 11 2
