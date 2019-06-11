package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"
)

type configfile struct {
	Offset      int64 `json:"offset"`
	Oldfilesize int64 `json:"oldfilesize"`
}

type resultcount struct {
	Twohundred  int `json:"200"`
	Fivehundred int `json:"500"`
}

var statuscode = make(map[string]int)
var statuscount resultcount

func checkfile(filename string) bool {
	if _, err := os.Stat(filename); err == nil {
		return true

	} else if os.IsNotExist(err) {
		// path/to/whatever does *not* exist
		return false
	} else {
		// Schrodinger: file may or may not exist. See err for details.
		fmt.Println(" Unknown error occured", err)
		return false
		// Therefore, do *NOT* use !os.IsNotExist(err) to test for file existence
	}

}

func createFile(filename string) {
	newFile, err := os.Create(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer newFile.Close()
}

/*
* funtion: getresultstat
* Input: path to result.json
* Output: Last Result json (type: resultcount)
*
 */
func getresultstat(filename string) resultcount {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	var obj resultcount
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println("error:", err)
	}
	return obj
}

/*
 * funtion: getlaststatanubhavsingh6663
 * input:  Path to confanubhavsingh6663ontains info about last offset and file size
 * output: Content of tanubhavsingh6663ion file, (type: configfile)
 */
func getlaststat(filename string) configfile {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Print(err)
	}
	var obj configfile
	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println(err)
	}
	return obj
}

func newfileinfo(filepath string) int64 {
	stat, err := os.Stat(filepath)
	if err != nil {
		fmt.Println(err)
	}
	filesize := stat.Size()
	return filesize
}
func writejsondata(filepath string, jsondata []byte) {
	fmt.Println("Writing data to", filepath, string(jsondata))
	_ = ioutil.WriteFile(filepath, jsondata, 0644)

}
func check(e error) {
	if e != nil {
		log.Fatal(e)

	}
}

func resultupdate(logLine string) {
	fmt.Println(logLine)
	response_status_reg := regexp.MustCompile(`( [0-9]{1,3} )`)
	response_status_list := response_status_reg.FindAllString(logLine, -1)
	for _, element := range response_status_list {
		element = strings.TrimSpace(element)
		if element == "200" {
			statuscount.Twohundred++
		}
		if element == "500" {
			statuscount.Fivehundred++
		}
	}
	fmt.Println("Result count:", statuscount)

}
func chunkread(filepath string, BufferSize int64, offset int64) {
	//const BufferSize = Buffer
	file, err := os.Open(filepath)
	check(err)
	defer file.Close()
	position, err := file.Seek(offset, 0)
	fmt.Println("Reading after ", position)
	check(err)
	buffer := make([]byte, BufferSize)

	for {
		bytesread, err := file.Read(buffer)
		if err != nil {
			if err != io.EOF {
				fmt.Println(err)
			}
			break
		}
		// the globla variable statuscount
		resultupdate(string(buffer[:bytesread]))
	}
}

func updatedata(configfilepath string, resultfile string, datafile string) {
	laststat := getlaststat(configfilepath)
	offset := laststat.Offset
	oldfilesize := laststat.Oldfilesize
	fmt.Println("Old file size", oldfilesize, "\nOld offset:", offset)
	resultstat := getresultstat(resultfile)
	print("Old result stat")
	fmt.Println(resultstat)

	newfilesize := newfileinfo(datafile)
	fmt.Println("New file size:", newfilesize)
	if oldfilesize != newfilesize {
		diff := newfilesize - oldfilesize
		fmt.Println("Data to read", diff)

		chunkread(datafile, diff, offset)
		confval := configfile{
			Offset:      oldfilesize + diff,
			Oldfilesize: newfilesize,
		}
		//d, err := json.Marshal(&init_val)
		d, err := json.Marshal(&confval)
		if err != nil {
			log.Fatalf("\njson.MarshalIndent failed with '%s'\n", err)
		}
		//updateconf(filepath)
		writejsondata(configfilepath, d)

		//Update result file
		resultval := resultcount{
			Twohundred:  resultstat.Twohundred + statuscount.Twohundred,
			Fivehundred: resultstat.Fivehundred + statuscount.Fivehundred,
		}
		fmt.Println("Updated result 2:", resultval)
		res, err := json.Marshal(&resultval)
		if err != nil {
			log.Fatalf("\njson.MarshalIndent failed with '%s'\n", err)
		}
		writejsondata(resultfile, res)
	} else {
		fmt.Print("No change detected")
	}
}

func checkconf(configfilepath string, resultfile string, datafile string) {
	// Check if the configfile is present
	exists := checkfile(configfilepath)
	if !exists {
		// Create conf.json if not present
		createFile(configfilepath)
		confval := configfile{
			Offset:      0,
			Oldfilesize: 0,
		}
		d, err := json.Marshal(&confval)
		if err != nil {
			log.Fatalf("\njson.MarshalIndent failed with '%s'\n", err)
		}
		// Write to the conf file
		writejsondata(configfilepath, d)

		// Create result file
		createFile(resultfile)
		resultval := resultcount{
			Twohundred:  0,
			Fivehundred: 0,
		}

		res, err := json.Marshal(&resultval)
		if err != nil {
			log.Fatalf("\njson.MarshalIndent failed with '%s'\n", err)
		}
		// Initialize the result file
		writejsondata(resultfile, res)
	}
}
func main() {
	configfilepath := "config.json"
	resultfile := "results.json"
	datafile := "sample.txt"

	for {
		go checkconf(configfilepath, resultfile, datafile)
		go updatedata(configfilepath, resultfile, datafile)
		time.Sleep(30 * time.Second)
	}
}

/**
 Cases:
	What if the server is crashed,restarted,deleted the log file,overwritten.
	What of log file is not present?
	What if the log file is overwritten
	What is log file is rolled out?

  **/
