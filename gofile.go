package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Other constants
const (
	SERVER          = "server"
	CLIENT          = "client"
	NONE            = "none"
	CONNECTION_TYPE = "tcp"
)

// Command constants
const (
	PWD  = "pwd"
	EXIT = "exit"
	LS   = "ls"
	CAT  = "cat"
	GET  = "get"
)

type commandJson struct {
	Command string   `json:"command"`
	Args    []string `json:"args"`
	Closed  bool     `json:"closed"`
}

type resultJson struct {
	Success          bool   `json:"success"`
	Result           []byte `json:"result"`
	ErrorDescription string `json:"errorDescription,omitempty"`
	FileName         string `json:"fileName,omitempty"`
}

func readInput() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')

	if err != nil {
		fmt.Println("[-] Error. Error Reading input")
		fmt.Println(err)
		os.Exit(1)
	}

	return strings.TrimSpace(input)
}

func handleConnection(conn net.Conn) {

	// client ip address and port
	addr := conn.RemoteAddr()
	clientAddress := addr.String()
	fmt.Printf("[+] Debug. Connection Received From %s\n", clientAddress)

	var closeCount int = 0

	for {

		jsonDecoder := json.NewDecoder(conn)

		var cmd commandJson

		err := jsonDecoder.Decode(&cmd)

		if err != nil {
			fmt.Println("[-] Error reading command")
			fmt.Println(err)
			closeCount += 1
			// send error message
		}

		if closeCount == 10 {
			break
		}

		if cmd.Command == EXIT {
			break
		}

		// fmt.Println("Server", cmd, err)

		var res resultJson

		c := strings.ToLower(strings.TrimSpace(cmd.Command))

		runCommand(c, cmd.Args, &res)

		// send response back
		// res.Success = true
		// res.Result = []byte("Hello world")

		// _, jsonErr := json.Marshal(res)

		// if jsonErr != nil {
		// 	fmt.Println("[-] Error Marshalling Resposne Data")
		// }

		encoder := json.NewEncoder(conn)

		encodeErr := encoder.Encode(res)

		if encodeErr != nil {
			fmt.Println("[-] Error. json write error")
			fmt.Println(encodeErr)
			conn.Close()
		}

	}

	fmt.Printf("[+] Debug. Connection Closed For %s\n", clientAddress)

	conn.Close()

	// time.Sleep(2 * time.Second)
	// fmt.Println("Server", jsonData)
	// fmt.Println("server", string(jsonData))
	// numberOfBytesWritten, writeErr := conn.Write(jsonData)

	// if writeErr != nil {
	// 	fmt.Println("[-] Error writting. ", writeErr)
	// } else {
	// 	fmt.Printf("[+] Wrote %d bytes\n", numberOfBytesWritten)
	// }

}

func startServer(port int) {
	// fmt.Println("Server main")
	address := fmt.Sprintf("0.0.0.0:%d", port)
	server, err := net.Listen(CONNECTION_TYPE, address)
	if err != nil {
		fmt.Println("[-] Error starting server. Exitting..")
		fmt.Println(err)
		os.Exit(1)
	}
	defer server.Close()
	fmt.Printf("[!] Debug. Started Server On %s\n", address)

	for {
		conn, connErr := server.Accept()
		if connErr != nil {
			fmt.Println("[-] Error accepting connection")
			continue
		}

		go handleConnection(conn)
	}
}

func startClient(host string, port int) {
	fmt.Println("client")
	fmt.Printf("connect to %s:%d\n", host, port)

	address := fmt.Sprintf("%s:%d", host, port)

	conn, err := net.Dial(CONNECTION_TYPE, address)
	if err != nil {
		fmt.Printf("[-] Error. Cannot Connect To %s on Port %d\n", host, port)
		fmt.Println(err)
		os.Exit(1)
	}
	defer conn.Close()

	// var isClosed bool = false

	for {

		// get command from user
		fmt.Printf("Enter command => ")
		c := readInput()

		if len(c) <= 0 {
			continue
		}

		clist := strings.Split(c, " ")
		// fmt.Println(clist)

		// convert to json
		var cmd commandJson
		cmd.Command = clist[0]
		if len(clist) > 1 {
			cmd.Args = clist[1:]
		}

		if cmd.Command == EXIT {
			cmd.Closed = true
		}

		// fmt.Println("cmd in struct", cmd)

		// _, cmdErr := json.Marshal(cmd)

		// if cmdErr != nil {
		// 	fmt.Println("[!] Error. Error Converting to JSON")
		// 	fmt.Println(cmdErr)
		// 	os.Exit(1)
		// }

		// write to socket
		encoder := json.NewEncoder(conn)
		encodeErr := encoder.Encode(cmd)

		if encodeErr != nil {
			fmt.Println("[-] Error. Error writing to server")
			fmt.Println(encodeErr)
			conn.Close()
			os.Exit(1)
		}

		if cmd.Command == EXIT {
			break
		}

		// get result back
		var r resultJson

		d := json.NewDecoder(conn)

		// fmt.Println("client", d)
		rj := d.Decode(&r)

		// fmt.Println("errrrrrr", rj)

		if rj != nil {
			fmt.Println("[-] Error. Error decoding data")
			fmt.Println(rj)
		}

		// fmt.Println("response message")
		// fmt.Println(r.Success)
		if len(r.Result) > 0 && r.Success {
			if cmd.Command == GET {
				writeFileClient(&r)
			}
			fmt.Println(string(r.Result))
		} else if !r.Success {
			fmt.Println(r.ErrorDescription)
		}
		// fmt.Println([]byte(r.ErrorDescription))
	}

}

func writeFileClient(res *resultJson) {
	wFilePtr, wFilePtrErr := os.OpenFile(res.FileName, os.O_RDWR|os.O_CREATE, 0644)

	if wFilePtrErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = wFilePtrErr.Error()
		return
	}

	_, writeErr := wFilePtr.Write(res.Result)

	if writeErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = writeErr.Error()
		return
	}
}

func usage() {
	fmt.Println("Usage (Server): ./gofile -stype server -lport 8443")
	fmt.Println("Usage (Client): ./gofile -stype client -shost 192.168.1.1 -sport 8443")
}

func lsUsage() string {
	var str string = ""
	str = str + "ls <One optional directory name>\t(If directory name not specificed, considers current working directory)"
	return str
}

func catUsage() string {
	return "cat <Single File name (Required)>"
}

func runCommand(c string, args []string, res *resultJson) {
	if c == PWD {
		// get current working directory
		cwd, cwdErr := os.Getwd()

		if cwdErr != nil {
			res.Success = false
			res.ErrorDescription = cwdErr.Error()
			res.Result = nil
		} else {
			res.Success = true
			res.ErrorDescription = ""
			res.Result = []byte(cwd)
		}
	} else if c == LS {

		if len(args) > 1 {
			res.Success = false
			res.ErrorDescription = lsUsage()
			res.Result = nil
			return
		}

		if len(args) == 1 {
			getFilesBasedOnDirectoryName(args[0], res)
		} else {

			// list files in current working directory
			currentWorkingDirectory, currentWorkingDirectoryErr := os.Getwd()

			if currentWorkingDirectoryErr != nil {
				res.Success = false
				res.Result = nil
				res.ErrorDescription = currentWorkingDirectoryErr.Error()
				return
			} else {
				getFilesBasedOnDirectoryName(currentWorkingDirectory, res)
			}
		}

	} else if c == CAT {
		readFileByFileName(c, args, res)
	} else if c == GET {
		// changes in client side
		readFileByFileName(c, args, res)
	} else {
		res.Success = true
		res.Result = []byte("Hello, World!")
	}
}

func readFileByFileName(c string, args []string, res *resultJson) {

	if len(args) != 1 {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = catUsage()
		return
	}

	filePtr, filePtrErr := os.Open(args[0])

	if filePtrErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = filePtrErr.Error()
		return
	}

	fileInfo, fileInfoErr := filePtr.Stat()

	if fileInfoErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = fileInfoErr.Error()
		return
	}

	buffer := make([]byte, fileInfo.Size())

	_, fileReadErr := filePtr.Read(buffer)

	if fileReadErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = fileReadErr.Error()
		return
	}

	res.Success = true
	res.Result = buffer
	res.ErrorDescription = ""
	if c == GET {
		res.FileName = args[0]
	}
}

func getFilesBasedOnDirectoryName(directoryName string, res *resultJson) {
	dirPtr, dirPtrErr := os.Open(directoryName)

	if dirPtrErr != nil {
		res.Success = false
		res.ErrorDescription = dirPtrErr.Error()
		res.Result = nil
		return
	}

	dirEntries, dirEntriesErr := dirPtr.ReadDir(0)

	if dirEntriesErr != nil {
		res.Success = false
		res.Result = nil
		res.ErrorDescription = dirEntriesErr.Error()
		return
	}

	var result string = ""

	for _, dirEntry := range dirEntries {
		fileInfo, _ := dirEntry.Info()
		if fileInfo.IsDir() {
			result = result + "(d)\t" + strconv.Itoa(int(fileInfo.Size())) + "\t\t" + fileInfo.ModTime().Truncate(time.Second).String() + "\t" + fileInfo.Name() + "\n"
		} else {
			result = result + "(f)\t" + strconv.Itoa(int(fileInfo.Size())) + "\t\t" + fileInfo.ModTime().Truncate(time.Second).String() + "\t" + fileInfo.Name() + "\n"
		}
	}

	res.Success = true
	res.Result = []byte(result)
	res.ErrorDescription = ""
}

func main() {

	stype := flag.String("stype", "none", "Server/Client")
	lport := flag.Int("lport", -1, "Port for Server to Listen")
	shost := flag.String("shost", "none", "IP Address of the Server to Connect To")
	sport := flag.Int("sport", -1, "Server Port To Connect")
	flag.Parse()

	if *stype == SERVER {
		if *lport == -1 {
			fmt.Println("[-] Error. Specify Port Number for Server to Listen on")
			usage()
			os.Exit(1)
		}
		startServer(*lport)
	} else if *stype == CLIENT {
		if *shost == NONE {
			fmt.Println("[-] Error. Need Server Address To Connect")
			usage()
			os.Exit(1)
		}
		if *sport == -1 {
			fmt.Println("[-] Error. Need Server Port To Connect")
			usage()
			os.Exit(1)
		}
		startClient(*shost, *sport)
	} else {
		usage()
	}
}