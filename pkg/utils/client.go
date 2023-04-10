package utils

import "net/http"

func HTTPClient() http.Client {
	c := http.Client{}
	return c
}
