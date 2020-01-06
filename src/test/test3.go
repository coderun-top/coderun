package main

import (
    "fmt"
//    "os"

    "github.com/go-ini/ini"
)

func main() {
    cfg, err := ini.Load(".agent.ini")
    if err != nil {
        fmt.Printf("Fail to read file: %v", err)
	cfg = ini.Empty()
        //os.Exit(1)
	//fmt.Println("App Mode:", cfg.Section("").Key("app_mode").String())
	//fmt.Println("Data Path:", cfg.Section("paths").Key("data").String())

	//// Let's do some candidate value limitation
	//fmt.Println("Server Protocol:",
	//    cfg.Section("server").Key("protocol").In("http", []string{"http", "https"}))
	//// Value read that is not in candidates will be discarded and fall back to given default value
	//fmt.Println("Email Protocol:",
	//    cfg.Section("server").Key("protocol").In("smtp", []string{"imap", "smtp"}))

	//// Try out auto-type conversion
	//fmt.Printf("Port Number: (%[1]T) %[1]d\n", cfg.Section("server").Key("http_port").MustInt(9999))
	//fmt.Printf("Enforce Domain: (%[1]T) %[1]v\n", cfg.Section("server").Key("enforce_domain").MustBool(false))

	 //cfg.Section("server").Key("protocol").In("http", []string{"http", "https"})
	 //cfg.Section("server").Key("protocol").In("smtp", []string{"imap", "smtp"})
	 //cfg.Section("server").Key("http_port").MustInt(9999)
	 //cfg.Section("server").Key("enforce_domain").MustBool(false)

	// Now, make some changes and save it
	cfg.Section("").Key("agent_id").SetValue("123123")
	cfg.SaveTo(".agent.ini")
    }
    fmt.Printf("agent_id:%s", cfg.Section("").Key("agent_id").String())

    // Classic read of values, default section can be represented as empty string
}
