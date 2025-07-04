package main

import (
	"github.com/xbapps/xbvr/pkg/server"
)

var version = "0.4.35 beta1"
var commit = "HEAD"
var branch = "master"
var date = "moment ago"

func main() {
	server.StartServer(version, commit, branch, date)
}
