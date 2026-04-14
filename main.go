package main

import (
	"fmt"
	"net/http"
)

func main() {
	fmt.Println("Servidor WhatsApp API iniciado en el puerto 8080")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "WhatsApp API funcionando")
	})
	http.ListenAndServe(":8080", nil)
}
