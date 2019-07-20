package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"

	"github.com/mitchellh/go-homedir"
)

func getConfigPath() (string, error) {
	if cfg := os.Getenv("MAILER_CONFIG"); cfg != "" {
		// config file is explicitly specified in env
		path, err := homedir.Expand(cfg)
		if err != nil {
			log.Fatal(err)
		}
		return path, err
	}

	homedir, err := homedir.Dir()
	if err != nil {
		// can't get home directory!
		log.Println("Can't get home directory:", err)
		return "", err
	}

	// look for ~/.mailerrc
	path := path.Join(homedir, ".mailerrc")

	return path, nil
}

func readAccount(f *os.File, verbose bool) (string, string, string) {
	scanner := bufio.NewScanner(f)

	if verbose {
		fmt.Println("Enter your login (gmail: it's your full address with @):")
	}
	ok := scanner.Scan()
	if !ok {
		log.Fatal("Can't read login", scanner.Err())
	}
	login := scanner.Text()

	if verbose {
		fmt.Println("Enter your password (gmail: you can get one at myaccount.google.com/apppasswords):")
	}
	ok = scanner.Scan()
	if !ok {
		log.Fatal("Can't read password", scanner.Err())
	}
	password := scanner.Text()

	if verbose {
		fmt.Println("Enter your imap server (gmail: it's imap.gmail.com:993):")
	}
	ok = scanner.Scan()
	if !ok {
		log.Fatal("Can't read host", scanner.Err())
	}
	host := scanner.Text()

	return login, password, host
}

func isAccessible(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getOrCreateAccount() (string, string, string) {
	path, err := getConfigPath()
	if err == nil && isAccessible(path) {
		// if config exists, try to read it
		f, err := os.Open(path)
		if err != nil {
			log.Println("Can't open config file:", err)
			// if it went wrong, read from stdin instead
			return readAccount(os.Stdin, true)
		}
		defer f.Close()

		// read from the config file
		return readAccount(f, false)
	} else {
		// if config doesn't exist, read it from stdin
		fmt.Print("Can't read config: ")
		if err != nil {
			fmt.Println(err)
		} else {
			_, err = os.Stat(path)
			fmt.Println(err)
		}
		fmt.Println("Creating a new one\u2026")
		login, password, host := readAccount(os.Stdin, true)

		// try to write config
		f, err := os.Create(path)
		if err != nil {
			log.Println("Can't create config file:", err)
		} else {
			defer f.Close()

			_, err = f.WriteString(login + "\n" + password + "\n" + host + "\n")
			if err == nil {
				err = f.Sync()
			}
			if err != nil {
				log.Println("Can't write config file:", err)
			} else {
				fmt.Println("Wrote config to", path)
			}
		}

		return login, password, host
	}
}
