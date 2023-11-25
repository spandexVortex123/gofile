package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	title, err := reader.ReadString('\n')
	fmt.Println(title, err)
	fmt.Println(strings.TrimSpace(title))
	fmt.Println([]byte(strings.TrimSpace(title)))
}
