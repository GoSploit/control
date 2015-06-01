package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"strconv"

	"github.com/gobs/args"
	"github.com/gobs/cmd"
	"github.com/gosploit/protocol"
	"github.com/peterh/liner"
)

var (
	Send *json.Encoder
	Recv *json.Decoder
)

/*
type sessions struct {
}

func (cmd sessions) Run() gribble.Value {

}

type select struct {
	name string `select`
	ID   int    `param:"1"`
}

func (cmd select) Run() gribble.Value {
	p := &protocol.Packet{
		ID: 1,
		Msg: protocol.SelectSessionRequest{
			ID: int64(cmd.ID),
		},
	}
	Send.Encode(p)
	Recv.Decode(&p)
	if p.Msg.(protocol.SelectSessionResponse).Error != "" {
		fmt.Println(p.Msg.(protocol.SelectSessionResponse).Error)
	} else {
		fmt.Println("Session selected")
	}
	return ""
}

type ls struct {
}

func List() gribble.Value {
	p := &protocol.Packet{
		ID:  1,
		Msg: protocol.ListCommand{},
	}
	Send.Encode(p)
	Recv.Decode(&p)
	for _, fileinfo := range p.Msg.(protocol.ListResponse).Files {
		dir := " "
		if fileinfo.IsDir {
			dir = "d"
		}
		fmt.Printf("%s\t%s\t%d\n", dir, fileinfo.Name, fileinfo.Size)
	}
	return ""
}

type help struct {
}

func (cmd help) Run() gribble.Value {
	fmt.Println(env.String())
	return ""
}
*/
func main() {
	conn, _ := net.Dial("tcp", "127.0.0.1:8989")
	Send = json.NewEncoder(conn)
	Recv = json.NewDecoder(conn)
	line := liner.NewLiner()
	line.SetCtrlCAborts(true)

	commander := &cmd.Cmd{Prompt: "> "}
	commander.Init()

	commander.Add(cmd.Command{
		Name: "sessions",
		Help: `List available sessions`,
		Call: func(line string) (stop bool) {
			p := &protocol.Packet{
				ID:  1,
				Msg: protocol.GetSessionsRequest{},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			for _, session := range p.Msg.(protocol.GetSessionsResponse).Sessions {
				fmt.Println(session.ID)
			}
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "select",
		Help: `Select a specific session`,
		Call: func(line string) (stop bool) {
			cmdargs := args.GetArgs(line)
			id, err := strconv.ParseInt(cmdargs[0], 10, 64)
			if err != nil {
				fmt.Println(err)
				return
			}
			p := &protocol.Packet{
				ID: 1,
				Msg: protocol.SelectSessionRequest{
					ID: id,
				},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			if p.Msg.(protocol.SelectSessionResponse).Error != "" {
				fmt.Println(p.Msg.(protocol.SelectSessionResponse).Error)
			} else {
				fmt.Println("Session selected")
			}
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "ls",
		Help: `List files and directories`,
		Call: func(line string) (stop bool) {
			p := &protocol.Packet{
				ID:  1,
				Msg: protocol.ListCommand{},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			for _, fileinfo := range p.Msg.(protocol.ListResponse).Files {
				dir := " "
				if fileinfo.IsDir {
					dir = "d"
				}
				fmt.Printf("%s\t%s\t%d\n", dir, fileinfo.Name, fileinfo.Size)
			}
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "cd",
		Help: `Change to a specific directory`,
		Call: func(line string) (stop bool) {
			cmdargs := args.GetArgs(line)
			if len(cmdargs) == 0 {
				fmt.Println("Directory is required")
				return
			}

			dir := cmdargs[0]
			p := &protocol.Packet{
				ID: 1,
				Msg: protocol.ChDirCommand{
					NewDir: dir,
				},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			if p.Msg.(protocol.ChDirResponse).Error != "" {
				fmt.Println(p.Msg.(protocol.ChDirResponse).Error)
			}
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "get",
		Help: `Get a file from the remote`,
		Call: func(line string) (stop bool) {
			cmdargs := args.GetArgs(line)
			if len(cmdargs) < 2 {
				fmt.Println("File is required")
				return
			}

			file := cmdargs[0]
			p := &protocol.Packet{
				ID: 1,
				Msg: protocol.GetCommand{
					File: file,
				},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			ioutil.WriteFile(cmdargs[1], p.Msg.(protocol.GetResponse).Data, 0666)
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "put",
		Help: `Put a file on the remote`,
		Call: func(line string) (stop bool) {
			cmdargs := args.GetArgs(line)
			if len(cmdargs) < 2 {
				fmt.Println("File is required")
				return
			}

			data, err := ioutil.ReadFile(cmdargs[0])
			if err != nil {
				fmt.Println(err)
				return
			}

			p := &protocol.Packet{
				ID: 1,
				Msg: protocol.PutCommand{
					File: cmdargs[1],
					Data: data,
				},
			}
			Send.Encode(p)
			Recv.Decode(&p)
			return
		},
	})

	commander.Add(cmd.Command{
		Name: "exit",
		Help: "exit",
		Call: func(line string) (stop bool) {
			return true
		},
	})

	commander.CmdLoop()
}
