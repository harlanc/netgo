package logger

import (
	"fmt"
)

//LogInfo log info
func LogInfo(msg string) {

	fmt.Println("Info:" + msg)
	//log.Printf()
}

//LogError log error
func LogError(msg string) {

	fmt.Println("Error:" + msg)
	//log.Printf()
}

//LogDebug log debug info
func LogDebug(msg string) {
	fmt.Println("Debug:" + msg)
}
