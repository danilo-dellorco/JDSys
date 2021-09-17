package main

import (
	"os"
	"fmt"
	"../main/services"
	"net/rpc"
	"encoding/json"
	"log"
	"net/http"
)

func main() {
	//go ListenHealthPing()
	if len(os.Args)!=1{
		fmt.Printf("Usage: go run server.go\n")
	}
	fmt.Printf("Server Waiting For Connection\n")

	service := services.InitializeService()
	rpc.Register(service)
	rpc.HandleHTTP()
	services.ListenHttpConnection()
}



func ListenHealthPing() {
	fmt.Printf("Starting Helth Service\n")
	handler := http.HandlerFunc(handleRequest)
	http.Handle("/example", handler)
	http.ListenAndServe(":80", nil)
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	resp := make(map[string]string)
	resp["message"] = "Status OK"
	jsonResp, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("Error happened in JSON marshal. Err: %s", err)
	}
	w.Write(jsonResp)
	return
}
