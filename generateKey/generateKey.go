package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"github.com/liserjrqlxue/crypto/aes"
	"github.com/liserjrqlxue/simple-util"
	"os/user"
)

var (
	usr = flag.String(
		"user",
		"",
		"username, default is current user",
	)
	code1 = flag.String(
		"code1",
		"118b09d39a5d3ecd56f9bd4f351dd6d6",
		"-code1 encode -user to get code3",
	)
	code2 = flag.String(
		"code2",
		"0e0760259f0826d18eb6e22988804617",
		"code3 encode -code2 to get codeKey",
	)
)

func main() {
	flag.Parse()
	if *usr == "" {
		User, err := user.Current()
		fmt.Printf("Gid\t\t%s\n", User.Gid)
		fmt.Printf("Uid\t\t%s\n", User.Uid)
		fmt.Printf("Name\t\t%s\n", User.Name)
		fmt.Printf("HomeDir\t\t%s\n", User.HomeDir)
		fmt.Printf("Username\t%s\n", User.Username)
		simple_util.CheckErr(err)
		*usr = User.Username
	}
	fmt.Printf("Usr\t%s\n", *usr)
	code3, err := AES.Encode([]byte(*usr), []byte(*code1))
	simple_util.CheckErr(err)
	fmt.Printf("Code1\t%s\n", *code1)
	fmt.Printf("Code2\t%s\n", *code2)
	fmt.Printf("Code3\t%x\n", []byte(code3))
	md5sum := md5.Sum([]byte(code3))
	fmt.Printf("md5sum\t%x\n", md5sum)
	codeKey, err := AES.Encode([]byte(*code2), []byte(fmt.Sprintf("%x", md5sum)))
	simple_util.CheckErr(err)
	fmt.Printf("codeKey\t%x\n", codeKey)
}
