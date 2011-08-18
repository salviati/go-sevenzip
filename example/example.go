package main

import (
	sz "sevenzip"
	"fmt"
	"flag"
	"log"
	"io/ioutil"
	"path/filepath"
	"os"
)

var filename = flag.String("f", "example.7z", "Filename")

func main() {
	flag.Parse()
	z, err := sz.Open(*filename)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Number of files in the archive: ", len(z.File))

	for _, f := range z.File {
		if f.IsDir != 0 {
			continue
		}

		fi, err := os.Stat(f.Name)
		if err == nil && uint64(fi.Size) == f.Size {
			fmt.Println("File ", f.Name, " already exists, skipping")
			continue
		}

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

		dir, _ := filepath.Split(f.Name)
		if dir != "" {
			if err = os.MkdirAll(dir, 0755); err != nil {
				log.Fatal(err)
			}
		}

		err = ioutil.WriteFile(f.Name, file, 0666)
		if err != nil {
			log.Fatal(err)
		}
	}
}
