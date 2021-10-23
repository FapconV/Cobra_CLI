/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"log"
	"net"

	"github.com/spf13/cobra"

	"google.golang.org/protobuf/proto"

	network "NetSim/packet"
	pb "NetSim/telemetry"
)

const (
	port = ":50051"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("server called")
		lis, err := net.Listen("tcp", port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		go func() {
			for {
				conn, err := lis.Accept()
				if err != nil {
					log.Fatalf("did not connect: %v", err)
				}

				if conn == nil {
					return
				}

				serv := network.NewStream(1024)

				serv.OnError(func(err network.IOError) {
					conn.Close()
				})

				serv.SetConnection(conn)

				go func() {
					for msg := range serv.Incoming {
						log.Print(msg.Data)
						out := &pb.Telemetry{}
						err := proto.Unmarshal(msg.Data, out)
						if err != nil {
							log.Fatalln("Failed to decode", err)
						}
						log.Print(out)

						serv.Outgoing <- network.New(0, []byte("Recieved"))
					}
				}()
			}
		}()
	},
}

func init() {
	startCmd.AddCommand(serverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
