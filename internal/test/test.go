package test

import (
	"fmt"
	"net/http"
)

func Test() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Get")
		fmt.Fprintf(w, "Hello, World!")
	})

	// Порт для прослушивания
	port := 8921

	fmt.Printf("Сервер запущен на порту %d...\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
	}
}
