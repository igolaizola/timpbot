package main

import (
	"flag"
	"log"
	"time"

	bot "github.com/igolaizola/timpbot"
)

func main() {
	emailFlag := flag.String("email", "", "user email")
	passFlag := flag.String("pass", "", "user password")
	centerFlag := flag.String("center", "", "numeric id of the center")
	activityFlag := flag.String("activity", "", "numeric id of the activity")
	dateFlag := flag.String("date", "", "date of the reservation (yyyy-mm-dd)")
	hourFlag := flag.String("hour", "", "hour of the reservation (hh:mm)")

	flag.Parse()
	if *emailFlag == "" {
		log.Fatal("email not provided")
	}
	if *passFlag == "" {
		log.Fatal("pass not provided")
	}
	if *centerFlag == "" {
		log.Fatal("center not provided")
	}
	if *activityFlag == "" {
		log.Fatal("activity not provided")
	}
	if *dateFlag == "" {
		log.Fatal("date not provided")
	}
	if *hourFlag == "" {
		log.Fatal("hour not provided")
	}

	for {
		err := bot.Book(*emailFlag, *passFlag, *centerFlag, *activityFlag, *dateFlag, *hourFlag)
		if err == nil {
			log.Printf("%s %s %s DONE!\n", *emailFlag, *dateFlag, *hourFlag)
			return
		}
		log.Println(err)
		<-time.After(5 * time.Second)
	}
}
