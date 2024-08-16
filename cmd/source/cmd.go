package source

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/morvencao/event-based-transport-demo/pkg/source"
	"github.com/morvencao/event-based-transport-demo/pkg/store"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/options/mqtt"
	"open-cluster-management.io/sdk-go/pkg/cloudevents/generic/types"
)

func NewSourceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "source",
		Short: "Start the source",
		Long:  "Start the Source.",
	}

	// create source options and add flags
	sourceOptions := newSourceOptions()
	sourceOptions.addSourceFlags(cmd.Flags())

	// run the source
	cmd.Run = sourceOptions.runSource

	return cmd
}

type sourceOptions struct {
	serverAddr    string
	sourceID      string
	transportType string
	transportAddr string
}

func newSourceOptions() *sourceOptions {
	return &sourceOptions{}
}

func (o *sourceOptions) addSourceFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.serverAddr, "server-addr", "localhost:8080", "Server address")
	fs.StringVar(&o.sourceID, "source-id", "source", "Source ID")
	fs.StringVar(&o.transportType, "transport-type", "mqtt", "Transport type")
	fs.StringVar(&o.transportAddr, "transport-addr", "localhost:1883", "Transport address")
}

func (o *sourceOptions) runSource(cmd *cobra.Command, args []string) {
	ctx, cancel := context.WithCancel(context.Background())
	var ceSourceOptions *options.CloudEventsSourceOptions
	switch o.transportType {
	case "mqtt":
		mqttOptions := &mqtt.MQTTOptions{
			KeepAlive: 60,
			PubQoS:    1,
			SubQoS:    1,
			Topics: types.Topics{
				SourceEvents: fmt.Sprintf("sources/%s/clusters/+/sourceevents", o.sourceID),
				AgentEvents:  fmt.Sprintf("sources/%s/clusters/+/agentevents", o.sourceID),
			},
			Dialer: &mqtt.MQTTDialer{
				BrokerHost: o.transportAddr,
				Timeout:    5 * time.Second,
			},
		}
		ceSourceOptions = mqtt.NewSourceOptions(mqttOptions, fmt.Sprintf("%s-client", o.sourceID), o.sourceID)
	default:
		log.Fatalf("Unsupported transport type: %s", o.transportType)
	}

	store := store.NewMemoryStore()
	eventController := source.NewEventController()
	apiServer := source.NewAPIServer(o.serverAddr, o.sourceID, store, eventController)

	// Start the source client
	resourceSourceClient, err := source.StartResourceSourceClient(ctx, ceSourceOptions, store)
	if err != nil {
		log.Fatalf("Failed to start source client: %v", err)
	}

	// Add event handlers
	eventController.AddEventHandler(source.CreateEvent, resourceSourceClient.OnCreate)
	eventController.AddEventHandler(source.UpdateEvent, resourceSourceClient.OnUpdate)
	eventController.AddEventHandler(source.DeleteEvent, resourceSourceClient.OnDelete)

	// Wait for SIGINT or SIGTERM signal
	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		defer cancel()
		<-stopCh
	}()

	// Start the event controller
	go eventController.Run(ctx)
	// Start the API server
	go apiServer.Start(ctx)

	<-ctx.Done()
}
