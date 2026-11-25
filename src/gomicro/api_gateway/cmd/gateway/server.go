package main

import "net/http"

// api gateway for synchronous blocking flow until it receives the responses from the services it calls
func main() {
	http.HandleFunc("/login", login)                                         // -> auth service via jwt
	http.HandleFunc("/customer/payment/authorize", customerPaymentAuthorize) // when a customer wants to make a payment to a merchant, first authorize the payment then capture it (payment/capture)
	http.HandleFunc("/customer/payment/capture", customerPaymentCapture)
	http.HandleFunc("/customer/ledger", customerLedger)
}

func login(w http.ResponseWriter, r *http.Request)                    {}
func customerPaymentAuthorize(w http.ResponseWriter, r *http.Request) {}
func customerPaymentCapture(w http.ResponseWriter, r *http.Request)   {}
func customerLedger(w http.ResponseWriter, r *http.Request)           {}
