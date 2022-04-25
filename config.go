package main

import (
	"fmt"
	"log"

	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type config struct {
	username  string
	local     bool
	port      int
	notify    bool
	forceHost bool
}

func newConfig() config {
	flag.StringP("username", "u", "noone", "user name")
	flag.BoolP("local", "l", false, "whether to search for a running server in localhost")
	flag.BoolP("notify", "n", true, "whether to send system notifications upon message receivals. Notifications have a cooldown time.")
	flag.BoolP("force-host", "f", false, "start as host without scanning for peers")
	flag.IntP("port", "p", 6776, "port ")
	var cfgPath = flag.StringP("config", "c", "", "path to config file")

	flag.Parse()
	viper.BindPFlags(flag.CommandLine)
	if *cfgPath != "" {
		viper.SetConfigFile(*cfgPath)
		if err := viper.ReadInConfig(); err != nil {
			fmt.Println(err)
			log.Fatalf("failed to load config file %v", *cfgPath)

		}
	}

	return config{
		username:  viper.GetString("username"),
		local:     viper.GetBool("local"),
		notify:    viper.GetBool("notify"),
		port:      viper.GetInt("port"),
		forceHost: viper.GetBool("force-host"),
	}
}
