package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/sunshineplan/metadata"
	"github.com/sunshineplan/password"
	"github.com/sunshineplan/service"
	"github.com/sunshineplan/utils"
	"github.com/sunshineplan/utils/flags"
	"github.com/sunshineplan/utils/httpsvr"
	"golang.org/x/net/publicsuffix"
)

var (
	priv *rsa.PrivateKey

	server = httpsvr.New()
	svc    = service.New()
	meta   metadata.Server
)

func init() {
	svc.Name = "MyAccounts"
	svc.Desc = "Instance to serve My Accounts"
	svc.Exec = run
	svc.TestExec = test
	svc.Options = service.Options{
		Dependencies: []string{"Wants=network-online.target", "After=network.target"},
		Environment:  map[string]string{"GIN_MODE": "release"},
	}
	svc.RegisterCommand("add", "add user", func(arg ...string) error {
		return addUser(arg[0])
	}, 1, true)
	svc.RegisterCommand("delete", "delete user", func(arg ...string) error {
		if utils.Confirm("Do you want to delete this user?", 3) {
			return deleteUser(arg[0])
		}
		return nil
	}, 1, true)
}

var (
	domain   = flag.String("domain", "", "Server Domain")
	maxRetry = flag.Int("retry", 5, "Max number of retries on wrong password")
	pemPath  = flag.String("pem", "", "PEM File Path")
	exclude  = flag.String("exclude", "", "Exclude Files")
	maxage   = flag.Int("maxage", 60*60*24*400, "Cookie Max-Age")
	logPath  = flag.String("log", "", "Log Path")
	// logPath = flag.String("log", filepath.Join(filepath.Dir(self), "access.log"), "Log Path")
)

func main() {
	self, err := os.Executable()
	if err != nil {
		svc.Fatalln("Failed to get self path:", err)
	}

	flag.StringVar(&meta.Addr, "server", "", "Metadata Server Address")
	flag.StringVar(&meta.Header, "header", "", "Verify Header Header Name")
	flag.StringVar(&meta.Value, "value", "", "Verify Header Value")
	flag.StringVar(&server.Unix, "unix", "", "UNIX-domain Socket")
	flag.StringVar(&server.Host, "host", "0.0.0.0", "Server Host")
	flag.StringVar(&server.Port, "port", "12345", "Server Port")
	flag.StringVar(&svc.Options.UpdateURL, "update", "", "Update URL")
	flags.SetConfigFile(filepath.Join(filepath.Dir(self), "config.ini"))
	flags.Parse()

	*domain, err = publicsuffix.EffectiveTLDPlusOne(*domain)
	if err != nil {
		svc.Fatal(err)
	}
	password.SetMaxAttempts(*maxRetry)
	if *pemPath != "" {
		b, err := os.ReadFile(*pemPath)
		if err != nil {
			svc.Fatal(err)
		}
		block, _ := pem.Decode(b)
		if block == nil {
			svc.Fatal("no PEM data is found")
		}
		priv, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			svc.Fatal(err)
		}
	}
	svc.Options.ExcludeFiles = strings.Split(*exclude, ",")

	if err := svc.ParseAndRun(flag.Args()); err != nil {
		svc.Fatal(err)
	}
}
