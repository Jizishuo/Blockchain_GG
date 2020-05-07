package main

import "flag"

func main() {
	cf := flag.String("c", "", "congig file")
	pprofPort := flag.Int("pprof", 0, "pprof prot, used by developers")
	flag.Parse()

	conf, err := parseConfig(*cf)
}
