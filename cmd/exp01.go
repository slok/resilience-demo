package main

import (
	"log"
	"net/http"
)

func mainExperiment01() {
	log.Println("Experiment 01 (naked server)")
	h := getServerHandler()
	http.ListenAndServe(addr, h)
}
