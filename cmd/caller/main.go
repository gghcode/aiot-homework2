package main

import (
	"fmt"
	"os"
	"solace.dev/go/messaging/pkg/solace/resource"
	"time"

	"solace.dev/go/messaging"
	"solace.dev/go/messaging/pkg/solace/config"
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

	publisher, err := messagingService.CreateDirectMessagePublisherBuilder().Build()
	if err != nil {
		panic(err)
	}

	defer publisher.Terminate(10 * time.Second)

	if err := publisher.Start(); err != nil {
		panic(err)
	}

	msg, _ := messagingService.MessageBuilder().BuildWithStringPayload("call request")
	// Publish the message to the topic my/topic/string

	if err := publisher.Publish(msg, resource.TopicOf("try-me")); err != nil {
		panic(err)
	}
}
