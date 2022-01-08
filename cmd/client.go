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
	"strconv"

	"github.com/spf13/cobra"

	"log"
	"net"
	"time"

	"encoding/binary"

	telem "NetSim/telemetry"
	//        "samples"

	"bufio"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

const (
	ENC_ST_MAX_DGRAM          uint32 = 64 * 1024
	ENC_ST_MAX_PAYLOAD        uint32 = 1024 * 1024
	ENC_ST_HDR_MSG_FLAGS_NONE uint16 = 0
	ENC_ST_HDR_MSG_SIZE       uint32 = 12
	ENC_ST_HDR_VERSION        uint16 = 1
)

type encapSTHdrMsgType uint16

const (
	ENC_ST_HDR_MSG_TYPE_UNSED encapSTHdrMsgType = iota
	ENC_ST_HDR_MSG_TYPE_TELEMETRY_DATA
	ENC_ST_HDR_MSG_TYPE_HEARTBEAT
)

type encapSTHdrMsgEncap uint16

const (
	ENC_ST_HDR_MSG_ENCAP_UNSED encapSTHdrMsgEncap = iota
	ENC_ST_HDR_MSG_ENCAP_GPB
	ENC_ST_HDR_MSG_ENCAP_JSON
	ENC_ST_HDR_MSG_ENCAP_GPB_COMPACT
	ENC_ST_HDR_MSG_ENCAP_GPB_KV
)

const (
	//	address = "localhost:50051"
	//	address = "10.105.236.227:5432"
	address = "103.249.77.26:5432"
)

type encapSTHdr struct {
	MsgType       encapSTHdrMsgType
	MsgEncap      encapSTHdrMsgEncap
	MsgHdrVersion uint16
	Msgflag       uint16
	Msglen        uint32
}

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
		intargument, _ := strconv.ParseInt(args[0], 10, 0)
		fmt.Println(int(intargument))
		// Set up a connection to the server.
		conn, err := net.Dial("tcp", address)
		if err != nil {
			log.Fatalf("did not connect: %v", err)
			fmt.Printf("did not connect: %v", err)
		}
		defer conn.Close()

		for i := 0; i < int(intargument); i++ {
			addrow()
			fullmsg := MDTSampleTelemetryTableFetchOne(
				SAMPLE_TELEMETRY_DATABASE_BASIC, i)
			gpbMessage := fullmsg.SampleStreamGPB

			hdr := encapSTHdr{
				MsgType:       ENC_ST_HDR_MSG_TYPE_TELEMETRY_DATA,
				MsgEncap:      ENC_ST_HDR_MSG_ENCAP_GPB,
				MsgHdrVersion: ENC_ST_HDR_VERSION,
				Msgflag:       ENC_ST_HDR_MSG_FLAGS_NONE,
				Msglen:        uint32(len(gpbMessage)),
			}

			err2 := binary.Write(conn, binary.BigEndian, &hdr)
			if err2 != nil {
				fmt.Println("Failed to write data header")
				return
			}
			if err != nil {
				log.Fatalln("Failed to encode address book:", err)
			}
			wrote, err := conn.Write(gpbMessage)
			if err != nil {
				fmt.Println("Failed write data 1")
				return
			}
			fmt.Println("Wrote", wrote, "expect", len(gpbMessage), "for data 1")
			time.Sleep(1 * time.Second)
		}
	},
}

type sampleTelemetryTable []SampleTelemetryTableEntry

type SampleTelemetryTableEntry struct {
	Sample             *telem.Telemetry
	SampleStreamGPB    []byte
	SampleStreamJSON   []byte
	SampleStreamJSONKV []byte
	Leaves             int
	Events             int
}

type SampleTelemetryDatabaseID int

const (
	SAMPLE_TELEMETRY_DATABASE_BASIC SampleTelemetryDatabaseID = iota
)

var database_size int

var sampleTelemetryDatabase map[SampleTelemetryDatabaseID]sampleTelemetryTable

func MDTSampleTelemetryTableFetchOne(
	dbindex SampleTelemetryDatabaseID, rowindex int) *SampleTelemetryTableEntry {

	if len(sampleTelemetryDatabase) <= int(dbindex) {
		return nil
	}

	table := sampleTelemetryDatabase[dbindex]
	return &table[rowindex]
}

type MDTContext interface{}
type MDTSampleCallback func(sample *SampleTelemetryTableEntry, context MDTContext) (abort bool)

//
// MDTSampleTelemetryTableIterate iterates over table of samples
// calling caller with function MDTSampleCallback and opaque context
// MDTContext provided, for every known sample. The number of samples
// iterated over is returned.
func MDTSampleTelemetryTableIterate(
	dbindex SampleTelemetryDatabaseID,
	fn MDTSampleCallback,
	c MDTContext) (applied int) {

	if len(sampleTelemetryDatabase) <= int(dbindex) {
		return 0
	}
	count := 0
	table := sampleTelemetryDatabase[dbindex]
	for _, entry := range table {
		count++
		if fn(&entry, c) {
			break
		}
	}

	return count
}

func MDTLoadMetrics() string {
	b, e := ioutil.ReadFile("mdt_msg_samples/dump.metrics")
	if e == nil {
		return string(b)
	}
	return ""
}

func getCurrTime() uint64 {
	now := time.Now()
	nanos := now.Unix()
	//millis := uint64(nanos / 1000000)
	return uint64(nanos)
}

func addrow() {
	marshaller := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
	}

	kv, err := os.Open("../samples/dump.jsonkv")
	if err != nil {
		fmt.Println(err)
	}
	defer kv.Close()

	dump := bufio.NewReader(kv)
	decoder := json.NewDecoder(dump)

	_, err = decoder.Token()
	if err != nil {
		fmt.Println(err)
	}

	// Read the messages and build the db.
	for decoder.More() {
		var m telem.Telemetry

		err := jsonpb.UnmarshalNext(decoder, &m)
		if err != nil {
			fmt.Println(err)
		}
		m.CollectionStartTime = getCurrTime()
		m.MsgTimestamp = getCurrTime()
		dataGpb := m.DataGpbkv
		var keys, content []*telem.TelemetryField
		for _, item := range dataGpb {
			item.Timestamp = getCurrTime()
			fileds := item.Fields
			for _, f := range fileds {
				f.Timestamp = getCurrTime()
				if f.Name == "keys" {
					keys = f.Fields
				} else if f.Name == "content" {
					content = f.Fields
				}
			}
		}

		fmt.Println("Keys:---------")
		for _, item := range keys {
			item.Timestamp = getCurrTime()
			fmt.Println(item)
		}

		fmt.Println("Content:---------")
		for _, item := range content {
			item.Timestamp = getCurrTime()
			if item.Name == "bytes-received" {
				RandomInt := rand.Intn(100000) + 100000
				randUnit64 := telem.TelemetryField_Uint64Value{
					Uint64Value: uint64(RandomInt),
				}
				item.ValueByType = &randUnit64
			}
			fmt.Println(item)
		}

		m.CollectionEndTime = getCurrTime() + 10
		gpbstream, err := proto.Marshal(&m)
		if err != nil {
			fmt.Println(err)
		}

		jsonstream, err := marshaller.MarshalToString(&m)
		if err != nil {
			fmt.Println(err)
		}

		entry := SampleTelemetryTableEntry{
			Sample:             &m,
			SampleStreamGPB:    gpbstream,
			SampleStreamJSONKV: json.RawMessage(jsonstream),
			Leaves:             strings.Count(jsonstream, "\"name\""),
			Events:             strings.Count(jsonstream, "\"content\""),
		}

		sampleTelemetryDatabase[SAMPLE_TELEMETRY_DATABASE_BASIC] =
			append(sampleTelemetryDatabase[SAMPLE_TELEMETRY_DATABASE_BASIC], entry)
	}
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
	sampleTelemetryDatabase = make(map[SampleTelemetryDatabaseID]sampleTelemetryTable)

	sampleTelemetryDatabase[SAMPLE_TELEMETRY_DATABASE_BASIC] = sampleTelemetryTable{}

	marshaller := &jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
	}

	kv, err := os.Open("../samples/dump.jsonkv")
	if err != nil {
		fmt.Println(err)
	}
	defer kv.Close()

	dump := bufio.NewReader(kv)
	decoder := json.NewDecoder(dump)

	_, err = decoder.Token()
	if err != nil {
		fmt.Println(err)
	}

	// Read the messages and build the db.
	for decoder.More() {
		var m telem.Telemetry

		err := jsonpb.UnmarshalNext(decoder, &m)
		if err != nil {
			fmt.Println(err)
		}
		m.CollectionStartTime = getCurrTime()
		m.MsgTimestamp = getCurrTime()
		dataGpb := m.DataGpbkv
		var keys, content []*telem.TelemetryField
		for _, item := range dataGpb {
			item.Timestamp = getCurrTime()
			fileds := item.Fields
			for _, f := range fileds {
				f.Timestamp = getCurrTime()
				if f.Name == "keys" {
					keys = f.Fields
				} else if f.Name == "content" {
					content = f.Fields
				}
			}
		}

		fmt.Println("Keys:---------")
		for _, item := range keys {
			item.Timestamp = getCurrTime()
			fmt.Println(item)
		}

		fmt.Println("Content:---------")
		for _, item := range content {
			item.Timestamp = getCurrTime()
			if item.Name == "bytes-received" {
				RandomInt := rand.Intn(100000) + 100000
				randUnit64 := telem.TelemetryField_Uint64Value{
					Uint64Value: uint64(RandomInt),
				}
				item.ValueByType = &randUnit64
			}
			fmt.Println(item)
		}

		m.CollectionEndTime = getCurrTime() + 10
		gpbstream, err := proto.Marshal(&m)
		if err != nil {
			fmt.Println(err)
		}

		jsonstream, err := marshaller.MarshalToString(&m)
		if err != nil {
			fmt.Println(err)
		}

		entry := SampleTelemetryTableEntry{
			Sample:             &m,
			SampleStreamGPB:    gpbstream,
			SampleStreamJSONKV: json.RawMessage(jsonstream),
			Leaves:             strings.Count(jsonstream, "\"name\""),
			Events:             strings.Count(jsonstream, "\"content\""),
		}

		sampleTelemetryDatabase[SAMPLE_TELEMETRY_DATABASE_BASIC] =
			append(sampleTelemetryDatabase[SAMPLE_TELEMETRY_DATABASE_BASIC], entry)
	}

}
