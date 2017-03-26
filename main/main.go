package main

import "simplefs/api"

func main() {
	s := api.NewServer(8008, "")
	s.Run()
}
