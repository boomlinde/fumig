package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

var ParseError error = errors.New("Line can not be parsed")
var UriError error = errors.New("The URI can not be decoded")
var displaytypes = map[rune]string{
	'0': "(TXT)", '1': "(DIR)", 's': "(SND)", 'g': "(GIF)",
	'I': "(PIC)", '9': "(BIN)", '5': "(ARC)", 'h': "(HTM)",
	'7': "(ISS)", 'i': "     ",
}

type gopherline struct {
	Ftype rune
	Text  string
	Path  string
	Host  string
	Port  int
}

const (
	textIndex = iota
	pathIndex
	hostIndex
	portIndex
)

func (l *gopherline) copy() *gopherline {
	return &gopherline{
		l.Ftype,
		l.Text,
		l.Path,
		l.Host,
		l.Port,
	}
}

func (l *gopherline) ToUri() string {
	uri := url.URL{
		Scheme: "gopher",
		Host:   fmt.Sprintf("%s:%d", l.Host, l.Port),
		Path:   string(l.Ftype) + l.Path,
	}
	return uri.String()
}

func (l *gopherline) FromUri(uri string) error {
	if !strings.HasPrefix(uri, "gopher://") {
		uri = "gopher://" + uri
	}
	urlp, err := url.Parse(uri)
	if err != nil {
		return UriError
	}
	if urlp.Scheme != "gopher" {
		return UriError
	}
	hostPort := strings.Split(urlp.Host, ":")
	if len(hostPort) == 1 {
		l.Host = hostPort[0]
		l.Port = 70
	} else if len(hostPort) == 2 {
		l.Host = hostPort[0]
		port, err := strconv.ParseInt(hostPort[1], 10, 16)
		if err != nil {
			return UriError
		}
		l.Port = int(port)
	}
	if len(urlp.Path) < 2 {
		l.Ftype = '1'
		l.Path = ""
	} else {
		l.Ftype = rune(urlp.Path[1])
		l.Path = urlp.Path[2:]
	}
	return nil
}

func (l *gopherline) Parse(line string) error {
	split := strings.Split(line, "\t")
	if len(split) < 4 || split[textIndex] == "" {
		return ParseError
	}
	l.Ftype = rune(split[textIndex][0])
	l.Text = split[textIndex][1:]
	l.Path = split[pathIndex]
	l.Host = split[hostIndex]
	port, err := strconv.ParseInt(split[portIndex], 10, 16)
	if err != nil {
		return ParseError
	}
	l.Port = int(port)

	return nil
}

func (l *gopherline) NiceType() string {
	fdisp := displaytypes[l.Ftype]
	if fdisp == "" {
		fdisp = "(BIN)"
	}
	return fdisp
}

func (l *gopherline) Pretty() string {
	return fmt.Sprintf("%s %s", l.NiceType(), l.Text)
}

func getMenu(selector *gopherline) ([]*gopherline, error) {
	s, err := getLines(selector)
	if err != nil {
		return nil, err
	}
	out := make([]*gopherline, 0, len(s))
	for _, line := range s {
		if line == "." {
			break
		}
		if line == "" {
			continue
		}
		l := gopherline{}
		err = l.Parse(line)
		if err != nil {
			continue // be lenient for now
		}
		out = append(out, &l)
	}
	if len(out) == 0 {
		out = append(out, &gopherline{Ftype: 'i'})
	}
	return out, nil
}

func getPlain(selector *gopherline) ([]string, error) {
	out, err := getLines(selector)
	if err != nil {
		return out, err
	}
	if out[len(out)-1] == "." {
		out = out[:len(out)-1]
	}
	return out, nil
}
