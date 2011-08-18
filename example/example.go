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
	
	fmt.Println("Number of files in the archive: ", len(z.File))

	for i, f := range z.File {
		if i==1 {continue}
		fmt.Println("Extracting ", f.Name)
		if f.IsDir != 0 { continue }
		/*r, err := f.Open()
		if err != nil {
			log.Fatal(err)
		}

		//file := f.ReadAll()
		file := make([]byte, f.Size)
		_, err = r.Read(file)
		if err != nil {
			log.Fatal(err)
		}*/
		file, err := z.ExtractUnsafe(i)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(fmt.Sprint("/tmp/hello",i), file, 0666)
		fmt.Println("done")
	}
}
