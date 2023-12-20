
import json
import argparse
import sys
import socket
import base64

# commands as constants
EXIT = "exit"
GET = "get"

class Command:
    def __init__(self, command: str, args: list, closed: bool):
        self.command = command
        self.args = args
        self.closed = closed

    def setCommand(self, command):
        self.command = command

    def setArgs(self, args):
        self.args = args

    def setClosed(self, closed):
        self.closed = closed

    def toString(self):
        return f"Command => {self.command}\nArgs => {self.args}\nClosed => {self.closed}"

    def getDict(self):
        d = dict()
        d["command"] = self.command
        d["args"] = self.args
        if self.closed is None:
            d["closed"] = True
        else:
            d["closed"] = self.closed

        return d
    
    def getCommand(self):
        return self.command


def getCommandObject(rawInput):
    rawInputSplit = rawInput.split(" ")
    if len(rawInputSplit) == 1:
        # Single command
        cmd = rawInputSplit[0]
        c = Command(cmd, None, False)
        if cmd.lower() == "exit":
            c.setClosed(True)

        return c

    else:
        # Command with args
        cmd = rawInputSplit[0]
        args = rawInputSplit[1:]
        c = Command(cmd, args, False)
        return c

if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("-shost", help="IP Address of the Server to Connect")
    parser.add_argument("-sport", help="Port to connect", type=int)
    args = parser.parse_args()
    if args.shost is None or args.sport is None:
        parser.print_help()
        sys.exit(1)

    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM, 0)
    sock.connect((args.shost, args.sport))

    while True:
        # get command object
        c = getCommandObject(input("Enter Command => "))
        if len(c.getCommand()) <= 0:
            continue

        if c.getCommand().lower() == EXIT:
            c.setClosed(True)

        # send request
        sock.send(json.dumps(c.getDict()).encode())

        # if command is exit
        if c.getCommand().lower() == EXIT:
            break

        # Read data into `data`
        data = b""
        while True:
            temp = sock.recv(1024)
            if not temp:
                break

            data += temp
            if b"\n" in data:
                break

        # Convert to json
        data_json = json.loads(data.decode())

        if data_json["success"] and not c.getCommand().lower() == GET:
            # not download file
            print(base64.b64decode(data_json["result"]).decode())
        elif data_json["success"] and c.getCommand().lower() == GET:
            # download file
            file_name = data_json["fileName"]
            file_name_replaced = file_name.replace("/", "_")
            with open(file_name_replaced, "wb") as f:
                f.write(base64.b64decode(data_json["result"]))
        elif not data_json["success"]:
            # print error
            print(data_json["errorDescription"])

    sock.close()