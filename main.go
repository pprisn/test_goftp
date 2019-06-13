//Popurey Sergey 2019
//https://github.com/pprisn/test_goftp
//Приложение выполняет отправку группы файлов по маске *.zip расположенных в каталоге локальной машины -ldir 
//на сервер указаный в параметре -serv, в каталог -remdir на удаленном сервере, 
//-username имя пользователя и -passwd пароль
package main

import (
	"flag"
	"fmt"
	"github.com/secsy/goftp"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var ldir     = flag.String("ldir",    "", `Каталог локальной машины, например -ldir C:\Elsag\RPAW\Logs\`)
var serv     = flag.String("serv",    "r00qlikviewftp.main.russianpost.ru", "Адрес сервера FTP,  например -serv `r00qlikviewftp.main.russianpost.ru")
var remdir   = flag.String("remdir",  "", "Каталог на удаленном сервере, например -remdir Datamatrix/RPAW/029/398000")
var username = flag.String("username","", "Имя учетной записи подключения к FTP серверу,  например -username mylogin")
var passwd   = flag.String("passwd",  "", "Пароль пользователя FTP сервером, например: -passwd password")
var wg sync.WaitGroup

func main() {
	var err error
        //md, mf := filepath.Split(os.Args[0])
	//fmt.Println(md)
        //fmt.Println(mf)
	flag.Parse()
	var floger *os.File

	if floger, err = os.OpenFile("sendfiles.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) ; err != nil {
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
		Logger:             floger, //os.Stderr,
	}

	ftp, err := goftp.DialConfig(config, *serv)

	defer ftp.Close()
	fmt.Println("Успешное соединение с сервером", *serv)

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
		go func(fl *os.File, mfile string) {
			if err := ftp.Store(*remdir+mfile, fl); err != nil {
				panic(err)
			}
			defer fl.Close()
			defer wg.Done()
		}(fl, mfile)
	}
	wg.Wait()

}
