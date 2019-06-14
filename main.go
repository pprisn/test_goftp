//Popurey Sergey 2019
//https://github.com/pprisn/test_goftp
//Приложение выполняет отправку группы файлов по маске *.zip расположенных в каталоге локальной машины -ldir 
//на сервер указаный в параметре -serv, в каталог -remdir на удаленном сервере, 
//-username имя пользователя и -passwd пароль
package main

import (
	"flag"
	"github.com/secsy/goftp"
	"os"
        "io"
	"log"
	"path/filepath"
	"sync"
	"time"
)

//send_to_elsag.exe -ldir C:\Elsag\RPAW\Logs\ -larchdir C:\Elsag\RPAW\Logs\Arc -serv ftp.rostovpost.ru -remdir Datamatrix/ -username r48cl### -passwd ####

var ldir     = flag.String("ldir",    "", `Каталог локальной машины, например -ldir C:\Elsag\RPAW\Logs\`)
var larchdir = flag.String("larchdir","", `Каталог локальной машины для переноса отправленных файлов в архив, например -larchdir C:\Elsag\RPAW\Logs\Arc`)
var serv     = flag.String("serv",    "r00qlikviewftp.main.russianpost.ru", "Адрес сервера FTP,  например -serv `r00qlikviewftp.main.russianpost.ru")
var remdir   = flag.String("remdir",  "", "Каталог на удаленном сервере, например -remdir Datamatrix/RPAW/029/398000")
var username = flag.String("username","", "Имя учетной записи подключения к FTP серверу,  например -username mylogin")
var passwd   = flag.String("passwd",  "", "Пароль пользователя FTP сервером, например: -passwd password")

var wg sync.WaitGroup


// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
    in, err := os.Open(src)
    if err != nil {
        return err
    }
    defer in.Close()
    out, err := os.Create(dst)
    if err != nil {
        return err
    }
    defer func() {
        cerr := out.Close()
        if err == nil {
            err = cerr
        }
    }()
    if _, err = io.Copy(out, in); err != nil {
        return err
    }
    err = out.Sync()
    if err != nil {
    log.Println("ОШИБКА копирования ",dst)
      return err
    }
    log.Println("Файл успешно скопирован ",dst)
    return nil
}

func deleteFile(path string) {
    // delete file
    var err = os.Remove(path)
    if isError(err){
        return
    }
    log.Println("Файл удален",path)
}

func isError(err error) bool {
    if err != nil {
        log.Println(err.Error())
    }
    return (err != nil)
}


// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success. Otherise, attempt to create a hard link
// between the two files. If that fail, copy the file contents from src to dst.
// эта функция в данном приложении не используется в цепочке .
//func CopyFile(src, dst string) (err error) {
//    sfi, err := os.Stat(src)
//    if err != nil {
//        return
//    }
//    if !sfi.Mode().IsRegular() {
//        // cannot copy non-regular files (e.g., directories,
//        // symlinks, devices, etc.)
//        return fmt.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
//    }
//    dfi, err := os.Stat(dst)
//    if err != nil {
//        if !os.IsNotExist(err) {
//            return
//        }
//    } else {
//        if !(dfi.Mode().IsRegular()) {
//            return fmt.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
//        }
//        if os.SameFile(sfi, dfi) {
//            return
//        }
//    }
//    if err = os.Link(src, dst); err == nil {
//        return
//    }
//    err = copyFileContents(src, dst)
//    return
//}



func main() {
	var err error
        //md, mf := filepath.Split(os.Args[0])
	//fmt.Println(md)
        //fmt.Println(mf)
	flag.Parse()
	var floger *os.File

	if floger, err = os.OpenFile("ftpmessage.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644) ; err != nil {
		panic(err)
	}
	defer floger.Close()


//f, err := os.OpenFile("testlogfile", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
//if err != nil {
//    log.Fatalf("error opening file: %v", err)
//}
//defer f.Close()

       log.SetOutput(floger)
       t0 := time.Now()
       log.Printf("СТАРТ %v \n", t0)
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
	log.Println("Успешное соединение с сервером", *serv)

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
		log.Println("Файл в очередь на отправку ",file)
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
        ftp.Close()
        //Перенос файлов в архивный каталог т.к. отправка файлов завершена
	if *larchdir !=""  {
		//Проверим, существет ли такой каталог
		if _, err := os.Stat(*larchdir); os.IsNotExist(err) {
                  if err:= os.MkdirAll(*larchdir, os.ModePerm); err != nil {
                     log.Println("Не удается создать архивный каталог", *larchdir)
                     return
                   }
                 }
		   for _, file:=range fileList {
                      mdir, mfile = filepath.Split(file)
                       // Move file moveFileContents(src,dst)
                       if err :=copyFileContents(file, *larchdir+`\`+mfile); err !=nil {
		          log.Println("Не удается скопировать файл в архивный каталог ", *larchdir+`\`+mfile)
                       } else {
                         deleteFile(file)
		       }
		   }
                
	 }

       t1 := time.Now()
       log.Printf("Успешное завершение работы, общее время выполнения %v сек.\n", t1.Sub(t0))
}
