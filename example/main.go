package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"go.acpr.dev/touchportal-golang-sdk/client"
	"go.acpr.dev/touchportal-golang-sdk/plugin"
)

type settings struct {
	Host string `json:"Host"`
	Port int    `json:"Port,string"`
}

var (
	counter int = 0
)

// Implement plugin.Settings on your settings struct to receive notification that the
// settings have been updated.
func (s *settings) IsUpdated() {
	fmt.Printf("settings updated: %#v\n", s)
}

func main() {
	ctx, cnl := context.WithCancel(context.Background())
	defer shutdownHandling(ctx, cnl)

	p := plugin.NewPlugin(ctx, "gsdk")

	// register settings before calling plugin.Register so we're made aware of the
	// plugin setting immediately
	p.Settings(&settings{})

	// registers our plugin with TouchPortal. Blocks until the plugin is ready for use
	err := p.Register()
	if err != nil {
		fmt.Printf("Failed to register plugin with TouchPortal. %s", err)
	}

	// add an action handler for our "gsdk_increment_counter" action
	p.OnAction(func(event client.ActionMessage) {
		fmt.Printf("Received action: %#v\n", event)

		counter++
		err := p.UpdateState("gsdk_counter", fmt.Sprint(counter))
		if err != nil {
			fmt.Printf("Failed to update state \"gsdk_counter\" with TouchPortal. %s", err)
		}
	}, "gsdk_increment_counter")

	// if you want an easy way to wait around for the plugin to exit plugin.Done() offers
	// an unbuffered channel you can watch.
	<-p.Done()
}

func shutdownHandling(ctx context.Context, cnl context.CancelFunc) func() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		select {
		case sig := <-c:
			fmt.Printf("Got %s signal. Quitting...\n", sig)
			cnl()
		case <-ctx.Done():
		}
	}()

	return func() {
		signal.Stop(c)
		cnl()
	}
}
