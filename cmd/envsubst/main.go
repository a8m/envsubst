// envsubst command line tool
package main

// BUG If -w fails, the restored original file may not have original file
//     attributes.

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/a8m/envsubst"
)

var (
	input   = flag.String("i", "", "")
	output  = flag.String("o", "", "")
	inplace = flag.Bool("w", false, "")
	noUnset = flag.Bool("no-unset", false, "")
	noEmpty = flag.Bool("no-empty", false, "")
)

var usage = `Usage: envsubst [options...] <input>
Options:
  -i         Specify file input, otherwise use last argument as input file. 
             If no input file is specified, read from stdin.
  -o         Specify file output. If none is specified, write to stdout.
  -w         Write result to input file (output is ignored).
  -no-unset  Fail if a variable is not set.
  -no-empty  Fail if a variable is set but empty.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(usage))
	}
	flag.Parse()
	// Reader
	var reader *bufio.Reader
	if *input != "" {
		file, err := os.Open(*input)
		if err != nil {
			usageAndExit(fmt.Sprintf("Error to open file input: %s.", *input))
		}
		defer file.Close()
		reader = bufio.NewReader(file)
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil || (stat.Mode()&os.ModeCharDevice) != 0 {
			usageAndExit("")
		}
		reader = bufio.NewReader(os.Stdin)
	}
	// Collect data
	var data string
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				data += line
				break
			}
			usageAndExit("Failed to read input.")
		}
		data += line
	}
	// Writer
	var file *os.File
	var backupFilename string
	var err error
	if *inplace {
		file, err = os.OpenFile(*input, os.O_WRONLY, 0666)
		if err != nil {
			usageAndExit(fmt.Sprintf(
				"Error while opening input file for writing: %s.",
				*input,
			))
		}
		backupFilename, err = backupFile(*input, data, 0644)
		if err != nil {
			errorAndExit(err)
		}
	} else if *output != "" {
		file, err = os.Create(*output)
		if err != nil {
			usageAndExit("Error to create the wanted output file.")
		}
	} else {
		file = os.Stdout
	}
	// Parse input string
	result, err := envsubst.StringRestricted(data, *noUnset, *noEmpty)
	if err != nil {
		errorAndExit(err)
	}
	if _, err := file.WriteString(result); err != nil {
		filename := *output
		if filename == "" {
			filename = "STDOUT"
		}
		if *inplace {
			err = os.Rename(backupFilename, *input)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Failed to recover backup file %s", backupFilename)
			}
		}
		usageAndExit(fmt.Sprintf("Error writing output to: %s.", filename))
	}

	if backupFilename != "" {
		err = os.Remove(backupFilename)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to remove backup file %s", backupFilename)
		}
	}
}

// backupFile writes data to a new file named <filename>_<number> with
// permissions perm, with <number randomly chosen such that the file name is
// unique. backupFile returns the chosen file name.
func backupFile(filename string, data string, perm os.FileMode) (string, error) {
	// create backup file
	f, err := ioutil.TempFile(
		filepath.Dir(filename),
		filepath.Base(filename)+"_",
	)
	if err != nil {
		return "", err
	}
	bakname := f.Name()

	// write data to backup file
	_, err = f.WriteString(data)
	if err1 := f.Close(); err == nil {
		err = err1
	}

	return bakname, err
}

func usageAndExit(msg string) {
	if msg != "" {
		fmt.Fprintf(os.Stderr, msg)
		fmt.Fprintf(os.Stderr, "\n\n")
	}
	flag.Usage()
	fmt.Fprintf(os.Stderr, "\n")
	os.Exit(1)
}

func errorAndExit(e error) {
	fmt.Fprintf(os.Stderr, "%v\n\n", e.Error())
	os.Exit(1)
}
