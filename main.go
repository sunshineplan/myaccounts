package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/password"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/httpsvr"
	"github.com/sunshineplan/utils/metadata"
	"github.com/vharitonsky/iniflags"
	"golang.org/x/net/publicsuffix"
)

var domain, pemPath, logPath string
var maxRetry int
var meta metadata.Server
var priv *rsa.PrivateKey

var server = httpsvr.New()
var svc = service.Service{
	Name:     "MyAccounts",
	Desc:     "Instance to serve My Accounts",
	Exec:     run,
	TestExec: test,
	Options: service.Options{
		Dependencies: []string{"Wants=network-online.target", "After=network.target"},
		Environment:  map[string]string{"GIN_MODE": "release"},
	},
}

func main() {
	self, err := os.Executable()
	if err != nil {
		log.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&domain, "domain", "", "Server Domain")
	flag.IntVar(&maxRetry, "retry", 5, "Max number of retries on wrong password")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&pemPath, "pem", "", "Private Key file")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	exclude := flag.String("exclude", "", "Exclude Files")
	//flag.StringVar(&logPath, "log", joinPath(dir(self), "access.log"), "Log Path")
	flag.StringVar(&logPath, "log", "", "Log Path")
	iniflags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	iniflags.SetAllowMissingConfigFile(true)
	iniflags.SetAllowUnknownFlags(true)
	iniflags.Parse()

	domain, err = publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil {
		log.Fatal(err)
	}
	password.SetMaxAttempts(maxRetry)
	if pemPath != "" {
		b, err := os.ReadFile(pemPath)
		if err != nil {
			log.Fatal(err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			log.Fatal("no PEM data is found")
		}
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
	}
	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if service.IsWindowsService() {
		svc.Run(false)
		return
	}

	switch flag.NArg() {
	case 0:
		run()
	case 1:
		switch flag.Arg(0) {
		case "run":
			svc.Run(false)
		case "debug":
			svc.Run(true)
		case "install":
			err = svc.Install()
		case "remove":
			err = svc.Remove()
		case "start":
			err = svc.Start()
		case "stop":
			err = svc.Stop()
		case "restart":
			err = svc.Restart()
		case "update":
			err = svc.Update()
		case "add", "delete":
			log.Fatalf("%s need two arguments", flag.Arg(0))
		default:
			log.Fatalln("Unknown argument:", flag.Arg(0))
		}
	case 2:
		switch flag.Arg(0) {
		case "add":
			addUser(flag.Arg(1))
		case "delete":
			if utils.Confirm(fmt.Sprintf("Do you want to delete user %s?", flag.Arg(1)), 3) {
				deleteUser(flag.Arg(1))
			}
		default:
			log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
		}
	default:
		log.Fatalln("Unknown arguments:", strings.Join(flag.Args(), " "))
	}
	if err != nil {
		log.Fatalf("Failed to %s: %v", flag.Arg(0), err)
	}
}
