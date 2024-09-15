package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/eclipse/paho.mqtt.golang"

	"go.viam.com/rdk/logging"

	"github.com/erh/viamseakeeper"
)

var f mqtt.MessageHandler = func(client mqtt.Client, msg mqtt.Message) {
	fmt.Printf("TOPIC: %s\n", msg.Topic())
	fmt.Printf("MSG: %s\n", msg.Payload())
}

func main() {

	err := mainReal()
	if err != nil {
		panic(err)
	}
}

func mainReal() error {

	host := ""
	power := false
	enable := false
	onFlag := false
	offFlag := false
	numTimes := 1
	flag.StringVar(&host, "host", "", "host of the seakeeper")
	flag.BoolVar(&power, "power", false, "")
	flag.BoolVar(&enable, "enable", false, "")
	flag.BoolVar(&onFlag, "on", false, "")
	flag.BoolVar(&offFlag, "off", false, "")
	flag.IntVar(&numTimes, "n", 1, "")

	flag.Parse()

	if power || enable {
		if onFlag && offFlag {
			return fmt.Errorf("cannot turn on and off")
		}
		if !onFlag && !offFlag {
			return fmt.Errorf("need to specify on or off")
		}
	}

	logger := logging.NewDebugLogger("seakeeper")
	s, err := viamseakeeper.NewSeakeeper(host, logger)
	if err != nil {
		return err
	}
	defer s.Close(context.Background())

	err = s.Start()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	if time.Since(s.LastStatusTime()) > time.Minute {
		return fmt.Errorf("no status")
	}
	fmt.Printf("%#v\n", s.LastStatus())

	if power {
		err = s.Power(onFlag)
		if err != nil {
			return err
		}
		numTimes++
	}
	if enable {
		err = s.Enable(onFlag)
		if err != nil {
			return err
		}
		numTimes++
	}

	for i := 1; i < numTimes; i++ {
		time.Sleep(1 * time.Second)
		fmt.Printf("%v \t %#v\n", time.Since(s.LastStatusTime()), s.LastStatus())
	}

	return nil
}
