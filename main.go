//open r00qlikviewftp.main.russianpost.ru
//user RPAWlog zR8TDzD5zb
//binary
//cd Datamatrix
//cd RPAW
//cd 029/398000
//lcd C:\Elsag\RPAW\Logs
//mput *.zip
//LITERAL PASV
//prompt
//BY

package main

import (
	//"crypto/sha256"
	//    "crypto/tls"
	"fmt"
	"path/filepath"
	//    "io"
	"flag"
	"os"
	"sync"
	"time"
//        "github.com/dutchcoders/goftp"
        "github.com/secsy/goftp"
	//    "encoding/hex"
	//"gopkg.in/dutchcoders/goftp.v1"
        //"github.com/VincenzoLaSpesa/goftp"

)

var ldir = flag.String("ldir", `C:\Elsag\RPAW\Logs\`, "Loclal Dir `C:/Elsag/RPAW/Logs`")
var serv = flag.String("serv", "r00qlikviewftp.main.russianpost.ru", "Server FTP `r00qlikviewftp.main.russianpost.ru`")
var remdir = flag.String("remdir", "Datamatrix/RPAW/029/", "Remote dir `Datamatrix/RPAW/029/398000`")
var username = flag.String("username", "", "User name `RPAWlog`")
var passwd = flag.String("passwd", "", "Passwd")
var wg sync.WaitGroup

func main() {
	var err error
//	var ftp *goftp.FTP

	flag.Parse()
	var floger *os.File
	if floger, err = os.Create("send_to_elsag_log.txt"); err != nil {
		panic(err)
	}
	defer floger.Close()

config := goftp.Config{
    User:               *username,
    Password:           *passwd,
    ConnectionsPerHost: 1,
    Timeout:            300 * time.Second,
    IPv6Lookup:         false,
    ActiveTransfers:    false,
    Logger:             floger,//os.Stderr,
}

        ftp, err := goftp.DialConfig(config, *serv)


	defer ftp.Close()
	fmt.Println("Successfully connected to", *serv)

        // Массив для хранения списка файлов
	fileList := []string{}
	var mdir, mfile string
	err = filepath.Walk(*ldir, func(path string, f os.FileInfo, err error) error {

		// проверим чтобы список формировался из файлов расположенных в заявленном каталоге
		mdir, mfile = filepath.Split(path)
		// fmt.Println(mfile)
		l, _ := filepath.Match(*ldir+"*.zip", path)
		if (mdir == *ldir) && (l == true) {
			fileList = append(fileList, path)
		}
		return nil
	})

	for _, file := range fileList {
		fmt.Println(file)
		var fl *os.File
		if fl, err = os.Open(file); err != nil {
			panic(err)
		}
		mdir, mfile = filepath.Split(file)
		wg.Add(1)
                go func(fl *os.File, mfile string){
		if err := ftp.Store(*remdir+mfile, fl); err != nil {
			panic(err)
		}
		defer fl.Close()
		defer wg.Done()
		}(fl, mfile)
//                wg.Wait()
	}
        wg.Wait()

}

/////////////
