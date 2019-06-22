package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

var (
	InvConfig = errors.New("invalid config")
)

type Processor struct {
	treeRoot *Vertex
}

func (p *Processor) Init(config map[string]interface{}, other ...interface{}) error {
	filename, ok := config["dictionary"].(string)
	if !ok {
		return InvConfig
	}

	p.treeRoot = &Vertex{
		root: true,
		next: make(map[byte]*Vertex, 1),
		to:   make(map[byte]*Vertex, 1),
	}

	resourcesPath := other[0]

	filePath := fmt.Sprintf("%s/%s", resourcesPath, filename)
	logrus.Info("Loading strings from ", filePath)

	inFile, err := os.Open(filePath)
	if err != nil {
		return InvConfig
	}
	defer inFile.Close()

	scanner := bufio.NewScanner(inFile)
	for scanner.Scan() {
		line := scanner.Text()
		p.treeRoot.Add([]byte(line))
	}

	return nil
}

func (p *Processor) Process(buf []byte) []byte {
	if p.treeRoot.Find(buf) {
		return buf[:0]
	}
	return buf
}

func GetProcessor() interface{} {
	return &Processor{}
}
