package util

import "log"

func ExitOnErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
