package main

import "net/http"

// api gateway is a synchronous blocking flow until it receives the responses from the services it calls
func main() {
	http.HandleFunc("/login", login)                                         // -> auth service
	http.HandleFunc("/customer/payment/authorize", customerPaymentAuthorize) // -> auth service -> money movement service
	http.HandleFunc("/customer/payment/capture", customerPaymentCapture)     // -> auth service -> money movement service
	http.HandleFunc("/customer/ledger", customerLedger)
}

func login(w http.ResponseWriter, r *http.Request)                    {}
func customerPaymentAuthorize(w http.ResponseWriter, r *http.Request) {}
func customerPaymentCapture(w http.ResponseWriter, r *http.Request)   {}
func customerLedger(w http.ResponseWriter, r *http.Request)           {}
