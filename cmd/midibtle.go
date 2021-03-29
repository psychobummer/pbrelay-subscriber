package cmd

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"time"

	"github.com/psychobummer/buttwork/device"
	"github.com/psychobummer/pbmidi"
	pbk "github.com/psychobummer/pbrelay/rpc/keystore"
	pbr "github.com/psychobummer/pbrelay/rpc/relay"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
)

var midibtleCmd = &cobra.Command{
	Use:   "midi-btle",
	Short: "Read MIDI data, and control local BTLE devices",
	Run:   doMidibtle,
}

func init() {
	midibtleCmd.Flags().StringP("connect", "c", "", "ip:port of the server to connect (e.g: 1.2.3.4:9999) (required)")
	midibtleCmd.Flags().StringP("producer", "p", "", "id of the producer stream (required)")

	midibtleCmd.MarkFlagRequired("connect")
	midibtleCmd.MarkFlagRequired("producer")
	rootCmd.AddCommand(midibtleCmd)
}

func doMidibtle(cmd *cobra.Command, args []string) {
	hostAddr, _ := cmd.Flags().GetString("connect")
	producer, _ := cmd.Flags().GetString("producer")
	conn, err := grpc.Dial(hostAddr, grpc.WithInsecure())
	if err != nil {
		log.Fatal().Msgf("couldnt dial server: %v", err)
	}
	defer conn.Close()

	keystoreClient := pbk.NewKeystoreServiceClient(conn)
	keyResp, err := keystoreClient.GetKey(context.Background(), &pbk.GetKeyRequest{
		Id: producer,
	})
	if err != nil {
		log.Fatal().Msg(err.Error())
	}

	relayClient := pbr.NewRelayServiceClient(conn)
	stream, err := relayClient.GetStream(context.Background(), &pbr.GetStreamRequest{
		Id: producer,
	})

	btleDevice, err := getBtleDevice("LVS")
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	defer func() {
		if err := btleDevice.Disconnect(); err != nil {
			log.Error().Msgf("couldnt disconnect from device; probably not a huge issue: %s", err.Error())
		}
	}()

	for {
		msg, err := stream.Recv()
		if err != nil {
			log.Error().Msgf("error reading from stream: %v", err)
			break
		}
		if !ed25519.Verify(keyResp.GetPublicKey(), msg.GetData(), msg.GetSignature()) {
			log.Error().Msg("could not validate signature of received mesage. Shutting down stream.")
			break
		}
		var midiMessage pbmidi.Message
		if err := json.Unmarshal(msg.GetData(), &midiMessage); err != nil {
			log.Error().Msgf("could not unmarshal payload data into midi message: %v", err)
			continue
		}

		var vibeLevel uint8
		switch midiMessage.State {
		case pbmidi.NoteOff:
			vibeLevel = 0
		case pbmidi.NoteOn:
			// scale the 0,127 midi velocity range to 0,100
			vibeLevel = uint8(100 * (float32(midiMessage.Veclocity) / 100))
		}
		if _, err := btleDevice.Vibrate(vibeLevel); err != nil {
			log.Error().Msgf("error talking to btle device: %s", err.Error())
		}
	}
}

func getBtleDevice(prefix string) (device.Device, error) {
	discovery, err := device.NewBLEDiscovery(device.TestConfig())
	if err != nil {
		return nil, err
	}
	identifiers, err := discovery.ScanOnce(2 * time.Second)
	if err != nil {
		return nil, err
	}
	filteredIdentifiers := identifiers.FilterPrefix(prefix)
	if len(filteredIdentifiers) == 0 {
		return nil, fmt.Errorf("No BTLE devices matching %s were found in the scan", prefix)
	}
	thisDevice, err := discovery.Connect(filteredIdentifiers[0])
	if err != nil {
		return nil, err
	}
	return thisDevice, nil
}
