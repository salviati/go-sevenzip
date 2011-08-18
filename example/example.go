package main

import (
	sz "sevenzip"
	"fmt"
	"flag"
	"log"
	"io/ioutil"
)

var filename = flag.String("f", "example.7z", "Filename")

func main() {
	flag.Parse()
	z, err := sz.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range z.File {
		fmt.Println("Extracting ", f.Name)
		r, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		//file := f.ReadAll()
		file := make([]byte, f.Size)
		_, err = r.Read(file)
		if err != nil {
			log.Fatal(err)
		}

		ioutil.WriteFile(f.Name, file, 0666)
	}
}
