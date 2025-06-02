package main

type ServerKeys struct {
	Media string
	Auth  string
}

var (
	keys = ServerKeys{
		Media: "media-server",
		Auth:  "auth-server",
	}
)
