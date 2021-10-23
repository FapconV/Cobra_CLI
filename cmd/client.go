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

	"github.com/spf13/cobra"

	"log"
	"net"

	"google.golang.org/protobuf/proto"

	pb "NetSim/telemetry"

	network "NetSim/packet"
)

const (
	address = "localhost:50051"
)

// clientCmd represents the client command
var clientCmd = &cobra.Command{
	Use:   "client",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("client called")
		var cid uint64 = 1
		bp := "tcp"
		sid := "venu"
		mv := "1.2"
		var csT uint64 = 7
		var mts uint64 = 5
		var ts uint64 = 2
		n := "venu"
		agd := true
		//vbt := pb.TelemetryField_BoolValue{
		//			BoolValue : true,
		//		}
		var ced uint64 = 5

		c := pb.Telemetry{
			CollectionId:           &cid,
			BasePath:               &bp,
			SubscriptionIdentifier: &sid,
			ModelVersion:           &mv,
			CollectionStartTime:    &csT,
			MsgTimestamp:           &mts,
			Fields: []*pb.TelemetryField{
				{
					Timestamp:   &ts,
					Name:        &n,
					AugmentData: &agd,
					//ValueByType: &vbt,
					Fields: []*pb.TelemetryField{},
				},
			},
			CollectionEndTime: &ced,
		}

		// Contact the server and print out its response.
		msg, err := proto.Marshal(&c)
		if err != nil {
			log.Fatalln("Failed to encode address book:", err)
		}

		log.Print(msg)

		// Set up a connection to the server.
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()

		s := network.NewStream(1024)
		s.OnError(func(err network.IOError) {
			conn.Close()
		})

		s.SetConnection(conn)

		s.Outgoing <- network.New(0, []byte(msg))

		recv := <-s.Incoming

		log.Print(recv.Data)
	},
}

func init() {
	rootCmd.AddCommand(clientCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// clientCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// clientCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
