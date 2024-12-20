package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/go-logfmt/logfmt"
)

var keyColor = color.New(color.FgRed).SprintFunc()
var valueColor = color.New(color.FgGreen).SprintFunc()
var cFlag = flag.Bool("c", false, "colorize output")
var jFlag = flag.Bool("j", false, "output json")

func main() {
	flag.Parse()
	var input *os.File
	fi, _ := os.Stdin.Stat()
	if (fi.Mode() & os.ModeCharDevice) == 0 {
		input = os.Stdin
	} else if len(flag.Args()) >= 1 {
		arg := flag.Args()
		f, err := os.Open(arg[0])
		if errors.Is(err, fs.ErrNotExist) {
			fmt.Printf("file '%s' does not exist.\n", arg[0])
			os.Exit(1)
		} else {
			input = f
		}
	} else {
		fmt.Println("Usage:")
		fmt.Println("lfpp is a tool to pretty print logfmt from stdin or from a file")
		fmt.Println("lfpp [-cj] [file]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	scanner := bufio.NewScanner(input)
	for scanner.Scan() {
		if *jFlag {
			fmt.Println(encodeJson(parseInputIntoMap(scanner.Text())))
		} else if *cFlag {
			fmt.Println(colorKV(parseInputIntoStringSlice(scanner.Text())))
		} else {
			fmt.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func parseInputIntoStringSlice(input string) [][]string {
	d := logfmt.NewDecoder(strings.NewReader(input))
	var out [][]string
	for d.ScanRecord() {
		for d.ScanKeyval() {
			out = append(out, []string{string(d.Key()), string(d.Value())})
		}
	}
	if d.Err() != nil {
		panic(d.Err())
	}
	return out
}

func parseInputIntoMap(input string) map[string]string {
	d := logfmt.NewDecoder(strings.NewReader(input))
	var outMap = make(map[string]string)
	for d.ScanRecord() {
		for d.ScanKeyval() {
			outMap[string(d.Key())] = string(d.Value())
		}
	}
	if d.Err() != nil {
		panic(d.Err())
	}
	return outMap
}

func colorKV(input [][]string) string {
	var outSlice []string
	for _, v := range input {
		kv := fmt.Sprintf("%s=%s", keyColor(v[0]), valueColor(v[1]))
		outSlice = append(outSlice, kv)
	}

	return strings.Join(outSlice, " ")
}

func encodeJson(input map[string]string) string {
	j, err := json.Marshal(input)
	if err != nil {
		panic(err)
	}
	return string(j)
}
