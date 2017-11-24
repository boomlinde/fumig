package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

const timeout = 10

func getLines(selector *gopherline) ([]string, error) {
	out := make([]string, 0)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", selector.Host, selector.Port), time.Second*timeout)
	if err != nil {
		return out, err
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(time.Second * timeout))
	reader := bufio.NewReader(conn)
	_, err = fmt.Fprintf(conn, "%s\r\n", selector.Path)
	if err != nil {
		return out, err
	}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				line = strings.TrimRight(line, "\r\n")
				if line != "" {
					out = append(out, line)
				}
				return out, nil
			} else {
				return out, err
			}
		}
		conn.SetDeadline(time.Now().Add(time.Second * timeout))

		line = strings.TrimRight(line, "\r\n")
		out = append(out, line)
	}
	return out, nil
}

func downloadFile(path string, selector *gopherline) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", selector.Host, selector.Port), time.Second*timeout)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "%s\r\n", selector.Path)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	for {
		conn.SetDeadline(time.Now().Add(time.Second * timeout))
		_, err := io.CopyN(file, conn, 1024)
		if err == io.EOF {
			file.Close()
			return nil
		}
		if err != nil {
			file.Close()
			os.Remove(path)
			return err
		}
	}
}
