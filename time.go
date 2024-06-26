package main

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"time"
)

// This file is for debugging purposes only.
// Do not turn DEBUG on unless you know what you're doing.
var DEBUG = false

func TimeTrack(functionStartTime time.Time) {
	// If the DEBUG option is set to true, this will display each function and its runtime in the console.
	if !DEBUG {
		return
	}

	timeElapsed := time.Since(functionStartTime)

	// Skip this function, and fetch the PC and file for its parent.
	pc, _, _, _ := runtime.Caller(1)

	// Retrieve a function object this functions parent.
	funcObj := runtime.FuncForPC(pc)

	// Regex to extract just the function name (and not the module path).
	runtimeFunc := regexp.MustCompile(`^.*\.(.*)$`)
	name := runtimeFunc.ReplaceAllString(funcObj.Name(), "$1")

	log.Println(fmt.Sprintf("%s took %s", name, timeElapsed))
}
