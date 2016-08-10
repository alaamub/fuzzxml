package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
)

/*
define variables for our xml fuzzer
*/
var (
	xmlfuzzer   = "/opt/xmlfuzzer/xmlfuzzer"
	radamsa     = "/usr/bin/radamsa"
	prefix      = "fuzz_"
	suffix      = ".xml"
	xmlArgs     = []string{"-xsd", "/opt/xmlfuzzer/OfficeOpenXML-XMLSchema/vml-main.xsd", "-root-elem", "document", "-max-elem", "10"}
	radamsaArgs = []string{"--seed", "12"}
	/* flags for command line */
	directory          = "./"
	version            = "1.0"
	ver                = flag.Bool("v", false, "Show Version.")
	xmlfuzzerBinaryArg = flag.String("xf", "", "Use xmlfuzzer tool - this flag does not require a seed")
	radamsaBinaryArg   = flag.String("ra", "", "Use Radamsa fuzzer tool.")
	seedArg            = flag.String("seed", "", "pass seed for the fuzzer by default will use /root/work/seed/xml-seed.xml .")
)

/*
define usage function
*/
func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tfuzzxml [-seed] path_to_seed.xml [-ra|-xf] binary_to_be_fuzzed\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

/*
* check error and panic
 */
func printError(err error) {
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("==> Error: %s\n", err.Error()))
	}
}

/*
* print the output
 */
func printOutput(outs []byte) {
	if len(outs) > 0 {
		fmt.Printf("==> Output: %s\n", string(outs))
	}
}

/*
* generate xml files from xmlfuzz
 */
func generateXMLMutations(cmd string, wg *sync.WaitGroup, xmlArgs []string) <-chan string {
	stringChan := make(chan string)
	go func() {
		duration := 10 * time.Millisecond
		time.Sleep(duration)
		command := exec.Command(cmd, xmlArgs...)
		var out bytes.Buffer
		command.Stdout = &out
		err := command.Run()
		printError(err)
		// generating random name for the file
		randBytes := make([]byte, 16)
		rand.Read(randBytes)
		xmlDumpFile := filepath.Join(os.TempDir(), prefix+hex.EncodeToString(randBytes)+suffix)
		fileHandle, err := os.Create(xmlDumpFile)
		// print error if any
		printError(err)
		writer := bufio.NewWriter(fileHandle)
		// close file handle
		defer fileHandle.Close()
		fmt.Fprintln(writer, out.String())
		writer.Flush()
		stringChan <- xmlDumpFile
		close(stringChan)
		wg.Done()
		fmt.Println("Generating XML is successfully completed, Duration:", duration)
	}()
	return stringChan
}

/*
* Fuzz the actual binary with the output generated by executeCommand()
 */
func fuzzBinary(receivedXMLFile <-chan string, wg *sync.WaitGroup, binary string) {
	go func() {
		duration := 10 * time.Millisecond
		time.Sleep(duration)
		for xmlFile := range receivedXMLFile {
			// initiate waitStatus for the syscall
			var waitStatus syscall.WaitStatus
			command := exec.Command(binary, xmlFile)
			var out bytes.Buffer
			command.Stdout = &out
			// run the command
			err := command.Run()
			// print error if any
			printError(err)
			color.Yellow("Now fuzzing %s using %s as input", binary, xmlFile)
			// check the status Exit codes, if crash happens it will exit and keep the file in /tmp
			if exitError, ok := err.(*exec.ExitError); ok {
				waitStatus = exitError.Sys().(syscall.WaitStatus)
				printOutput([]byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
				os.Exit(1)
			} else {
				// just relax and fuzzing will continue
				color.Red("Relax nothing yet ...")
				printOutput([]byte(fmt.Sprintf("%d", waitStatus.ExitStatus())))
				defer os.Remove(xmlFile)
			}
		}
		wg.Done()
	}()
}

func main() {
	// parse the flags for usage
	flag.Parse()
	flag.Usage = usage
	var wg sync.WaitGroup
	// print version if the flag -v is provided
	if *ver {
		fmt.Println(version)
		os.Exit(1)
	}

	if *xmlfuzzerBinaryArg != "" {
		for {
			generatedXML := generateXMLMutations(xmlfuzzer, &wg, xmlArgs)
			wg.Add(1)
			fuzzBinary(generatedXML, &wg, *xmlfuzzerBinaryArg)
			wg.Wait()
			fmt.Println("Done")
		}
	} else if *radamsaBinaryArg != "" && *seedArg != "" {
		seed := append(radamsaArgs, *seedArg)
		for {
			wg.Add(1)
			generatedXML := generateXMLMutations(radamsa, &wg, seed)
			wg.Add(1)
			fuzzBinary(generatedXML, &wg, *radamsaBinaryArg)
			wg.Wait()
			fmt.Println("Done")
		}
	} else {
		usage()
		os.Exit(1)
	}
}
