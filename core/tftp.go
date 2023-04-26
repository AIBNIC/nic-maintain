package core

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/pin/tftp/v3"
)

var tftpFolder string
var s *tftp.Server

// readHandler is called when client starts file download from server
func readHandler(filename string, rf io.ReaderFrom) error {
	file, err := os.Open(tftpFolder + filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := rf.ReadFrom(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("tftp server: %d bytes sent\n", n)
	return nil
}

// writeHandler is called when client starts file upload to server
func writeHandler(filename string, wt io.WriterTo) error {
	file, err := os.OpenFile(tftpFolder+filename, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	n, err := wt.WriteTo(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return err
	}
	fmt.Printf("tftp server: %d bytes received\n", n)
	return nil
}

func Start_tftp_process() {
	fmt.Println("请允许防火墙通过，否则将不能使用备份功能")

	s = tftp.NewServer(readHandler, writeHandler)

	go func() {
		err := s.ListenAndServe(":69") // blocks until s.Shutdown() is called
		if err != nil {
			fmt.Printf("tftp server: %v\n错误！69端口被占用\n", err)
		}
	}()
}

func Stop_tftp_process() {
	s.Shutdown()
}

func get_host_ip() (string, error) {
	addrs, err := net.InterfaceAddrs()

	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		}
		if ip == nil || !ip.IsPrivate() {
			continue
		}
		return ip.String(), nil
	}

	return "", err
}

func Tftp_server() (string, error) {
	time_folor := time.Now().Format("2006010215")
	tftpFolder = "./Econnect_box/" + time_folor + "/"

	_, err := os.Stat(tftpFolder)
	if err != nil && os.IsNotExist(err) {
		os.Mkdir(tftpFolder, os.ModePerm)
	}

	return get_host_ip()
}
