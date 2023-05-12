package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"time"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
	"solace.dev/go/messaging/pkg/solace/message"
	"solace.dev/go/messaging/pkg/solace/resource"
)

func getEnv(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

func main() {
	brokerConfig := config.ServicePropertyMap{
		config.TransportLayerPropertyHost:                getEnv("SOLACE_HOST", "tcp://localhost:55554"),
		config.ServicePropertyVPNName:                    getEnv("SOLACE_VPN", "default"),
		config.AuthenticationPropertySchemeBasicPassword: getEnv("SOLACE_PASSWORD", "default"),
		config.AuthenticationPropertySchemeBasicUserName: getEnv("SOLACE_USERNAME", "default"),
	}

	// Skip certificate validation
	messagingService, err := messaging.NewMessagingServiceBuilder().
		FromConfigurationProvider(brokerConfig).
		WithTransportSecurityStrategy(config.NewTransportSecurityStrategy().WithoutCertificateValidation()).
		Build()

	if err != nil {
		panic(err)
	}

	// Connect to the messaging serice
	if err := messagingService.Connect(); err != nil {
		panic(err)
	}

	defer messagingService.Disconnect()

	fmt.Println("Connected to the broker? ", messagingService.IsConnected())

	subscriber, err := messagingService.CreatePersistentMessageReceiverBuilder().
		WithSubscriptions(resource.TopicSubscriptionOf("*/Call/>")).
		Build(resource.QueueDurableNonExclusive("queue-call"))
	if err != nil {
		panic(err)
	}

	defer subscriber.Terminate(10 * time.Second)

	if err := subscriber.Start(); err != nil {
		panic(err)
	}

	publisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		panic(err)
	}

	defer publisher.Terminate(10 * time.Second)

	if err := publisher.Start(); err != nil {
		panic(err)
	}

	if err := subscriber.ReceiveAsync(func(message message.InboundMessage) {
		payload, ok := message.GetPayloadAsString()
		if ok {
			fmt.Printf("Message Payload %s \n", payload)
		}

		fmt.Printf("Message Destination %s \n", message.GetDestinationName())

		callTopics := strings.Split(message.GetDestinationName(), "/")
		logTopic := fmt.Sprintf("Call-Log/%s/%d/%s/%s",
			"af4092", time.Now().Unix(), callTopics[2], callTopics[3])

		msg, _ := messagingService.MessageBuilder().BuildWithStringPayload("log")
		if err := publisher.Publish(msg, resource.TopicOf(logTopic)); err != nil {
			panic(err)
		}
	}); err != nil {
		panic(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	// Block until a signal is received.
	<-c
}
