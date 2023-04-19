package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorBlue  = "\033[34m"
	noColor    = "\033[0m"
)

type Constructor struct {
	constructorName string
	variableNames   []string
}

func check(e error) {
	if e != nil {
		fmt.Println(e.Error(), colorRed)
	}
}

func generateFileName(filename string) (string, error) {
	x := strings.Split(filename, ".")
	if len(x) != 2 {
		return "", fmt.Errorf("NOT A VALID FILE")
	}
	return fmt.Sprintf("%s_toString.g.%s", x[0], x[1]), nil
}

func generateFile(constructors []*Constructor, path string, ch chan string) {
	filename := filepath.Base(path)

	generatedFileName, err := generateFileName(filename)
	check(err)

	err = writeToMainFile(path, constructors, generatedFileName)
	if err != nil {
		return
	}

	generatedFilePath := filepath.Join(filepath.Dir(path), generatedFileName)

	f, err := os.Create(generatedFilePath)
	check(err)

	defer f.Close()

	var body bytes.Buffer

	body.WriteString("/// ****** Generated Code ************\n\n")
	body.WriteString(fmt.Sprintf("part of \"%s\";\n\n", filename))
	// write the function for each constructor
	for _, c := range constructors {
		body.WriteString(fmt.Sprintf("/// **%s** Model **`toString()`** extended method\n", c.constructorName))
		body.WriteString(fmt.Sprintf("String _$%sToString(%s instance) {\n", c.constructorName, c.constructorName))
		body.WriteString("  return \"")
		body.WriteString(fmt.Sprintf("%s: ", c.constructorName))
		for _, name := range c.variableNames {
			body.WriteString(fmt.Sprintf("%s: ${instance.%s},", name, name))
		}
		body.WriteString("\";\n")
		body.WriteString("}\n\n")
	}

	f.Write(body.Bytes())

	ch <- generatedFilePath
}

func writeToMainFile(path string, constructors []*Constructor, generatedName string) error {
	header := fmt.Sprintf("part \"%s\";\n", generatedName)

	if len(constructors) == 0 {
		fmt.Printf("%sNo constructor in file > %s\n%s", path, colorRed, noColor)
		fmt.Println("Skipping this file...")
		return fmt.Errorf("Constructor Doesn't Exist")
	}

	file, err := os.OpenFile(path, os.O_RDWR, 0644)
	check(err)

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var line string

	// import statement and header buffer
	var importBuffer strings.Builder
	var otherBuffer strings.Builder

	ok := true
	importStarted := true

	for scanner.Scan() {
		line = scanner.Text()
		if importStarted && (strings.HasPrefix(line, "class") || strings.Contains(line, constructors[0].constructorName)) {
			importBuffer.WriteString(header + "\n")
			importStarted = false
		}
		if importStarted {
			importBuffer.WriteString(line + "\n")
		} else {
			otherBuffer.WriteString(line + "\n")
		}
		if strings.HasPrefix(line, "part") {
			ok = false
		}
	}
	check(scanner.Err())

	if ok {
		n, err := file.WriteAt([]byte(importBuffer.String()), 0)
		check(err)
		file.WriteAt([]byte(otherBuffer.String()), int64(n))
	}

	return nil
}

func fileAction(path string, ch chan string, wg *sync.WaitGroup) {
	f, err := os.ReadFile(path)
	check(err)

	fileContent := string(f)

	classRegex := regexp.MustCompile(`class\s+(\w+)\s*{([^}]+)}`)
	re := regexp.MustCompile(`this\.(\w+)`)

	classMatches := classRegex.FindAllStringSubmatch(fileContent, -1)

	var constructors []*Constructor

	for _, classMatch := range classMatches {
		className := classMatch[1]
		classBody := classMatch[2]

		var constructor Constructor

		constructor.constructorName = className

		lines := strings.Split(classBody, "\n")
		for _, line := range lines {
			// skip comment lines
			if strings.Contains(line, "//") {
				continue
			}

			x := re.FindAllStringSubmatch(line, -1)
			for _, y := range x {
				constructor.variableNames = append(constructor.variableNames, y[1])
			}
		}
		constructors = append(constructors, &constructor)
	}
	generateFile(constructors, path, ch)
	wg.Done()
}

func main() {
	args := os.Args[1:]
	channel := make(chan string)
	go func() {
		var wg sync.WaitGroup
		for _, path := range args {
			wg.Add(1)
			go fileAction(path, channel, &wg)
		}
		wg.Wait()
		close(channel)
	}()
	for msg := range channel {
		fmt.Printf("Done! File > %s%s\n%s", colorGreen, msg, noColor)
	}
	fmt.Println("All Done!", colorBlue)
}
